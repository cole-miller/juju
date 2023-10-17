// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package upgradedatabase

import (
	"context"
	"time"

	"github.com/juju/clock"
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	"github.com/juju/version/v2"
	"github.com/juju/worker/v3"
	"github.com/juju/worker/v3/catacomb"
	"github.com/juju/worker/v3/dependency"

	"github.com/juju/juju/agent"
	coredatabase "github.com/juju/juju/core/database"
	"github.com/juju/juju/core/upgrade"
	"github.com/juju/juju/core/watcher"
	"github.com/juju/juju/domain/model"
	"github.com/juju/juju/domain/schema"
	domainupgrade "github.com/juju/juju/domain/upgrade"
	upgradeerrors "github.com/juju/juju/domain/upgrade/errors"
	jujuversion "github.com/juju/juju/version"
	"github.com/juju/juju/worker/gate"
)

const (
	// defaultUpgradeTimeout is the default timeout for the upgrade to complete.
	// 20 minutes should be enough for the db upgrade to complete.
	defaultUpgradeTimeout = 20 * time.Minute
)

// UpgradeService is the interface for the upgrade service.
type UpgradeService interface {
	// CreateUpgrade creates an upgrade to and from specified versions
	// If an upgrade is already running/pending, return an AlreadyExists err
	CreateUpgrade(ctx context.Context, previousVersion, targetVersion version.Number) (domainupgrade.UUID, error)
	// SetControllerReady marks the supplied controllerID as being ready
	// to start its upgrade. All provisioned controllers need to be ready
	// before an upgrade can start
	SetControllerReady(ctx context.Context, upgradeUUID domainupgrade.UUID, controllerID string) error
	// StartUpgrade starts the current upgrade if it exists
	StartUpgrade(ctx context.Context, upgradeUUID domainupgrade.UUID) error
	// SetDBUpgradeCompleted marks the upgrade as completed in the database
	SetDBUpgradeCompleted(ctx context.Context, upgradeUUID domainupgrade.UUID) error
	// SetDBUpgradeFailed marks the upgrade as failed in the database
	SetDBUpgradeFailed(ctx context.Context, upgradeUUID domainupgrade.UUID) error
	// ActiveUpgrade returns the uuid of the current active upgrade.
	// If there are no active upgrades, return a NotFound error
	ActiveUpgrade(ctx context.Context) (domainupgrade.UUID, error)
	// WatchForUpgradeReady creates a watcher which notifies when all controller
	// nodes have been registered, meaning the upgrade is ready to start.
	WatchForUpgradeReady(ctx context.Context, upgradeUUID domainupgrade.UUID) (watcher.NotifyWatcher, error)
	// WatchForUpgradeState creates a watcher which notifies when the upgrade
	// has reached the given state.
	WatchForUpgradeState(ctx context.Context, upgradeUUID domainupgrade.UUID, state upgrade.State) (watcher.NotifyWatcher, error)
}

// ModelManagerService is the interface for the model manager service.
type ModelManagerService interface {
	// ModelList returns a list of all model UUIDs.
	// This only includes active models from the perspective of dqlite. These
	// are not the same as alive models.
	ModelList(context.Context) ([]model.UUID, error)
}

// NewLock returns a new gate.Lock that is unlocked if the agent has not the same version as juju
func NewLock(agentConfig agent.Config) gate.Lock {
	lock := gate.NewLock()

	// Build numbers are irrelevant to upgrade steps.
	upgradedToVersion := agentConfig.UpgradedToVersion().ToPatch()
	currentVersion := jujuversion.Current.ToPatch()

	if upgradedToVersion == currentVersion {
		lock.Unlock()
	}

	return lock
}

// Config holds the configuration for the worker.
type Config struct {
	// DBUpgradeCompleteLock is a lock used to synchronise workers that must
	// start after database upgrades are verified as completed.
	DBUpgradeCompleteLock gate.Lock

	// Agent is the running machine agent.
	Agent agent.Agent

	// ModelManagerService is the model manager service used to identify
	// the model uuids required to upgrade.
	ModelManagerService ModelManagerService

	// UpgradeService is the upgrade service used to drive the upgrade.
	UpgradeService UpgradeService

	// DBGetter is the database getter used to get the database for each model.
	DBGetter coredatabase.DBGetter

	// Tag holds the controller tag information.
	Tag names.Tag

	// Versions of the source and destination.
	FromVersion version.Number
	ToVersion   version.Number

	Logger Logger
	Clock  clock.Clock
}

// Validate validates the worker configuration.
func (c Config) Validate() error {
	if c.DBUpgradeCompleteLock == nil {
		return errors.NotValidf("nil DBUpgradeCompleteLock")
	}
	if c.Agent == nil {
		return errors.NotValidf("nil Agent")
	}
	if c.Logger == nil {
		return errors.NotValidf("nil Logger")
	}
	if c.Clock == nil {
		return errors.NotValidf("nil Clock")
	}
	if c.FromVersion == version.Zero {
		return errors.NotValidf("invalid FromVersion")
	}
	if c.ToVersion == version.Zero {
		return errors.NotValidf("invalid ToVersion")
	}
	if c.Tag == nil {
		return errors.NotValidf("invalid Tag")
	}
	return nil
}

type upgradeDBWorker struct {
	catacomb catacomb.Catacomb

	dbUpgradeCompleteLock gate.Lock

	controllerID string

	fromVersion version.Number
	toVersion   version.Number

	dbGetter coredatabase.DBGetter

	modelManagerService ModelManagerService
	upgradeService      UpgradeService

	logger Logger
	clock  clock.Clock
}

// NewUpgradeDatabaseWorker returns a new Worker.
func NewUpgradeDatabaseWorker(config Config) (worker.Worker, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Trace(err)
	}

	w := &upgradeDBWorker{
		dbUpgradeCompleteLock: config.DBUpgradeCompleteLock,

		controllerID: config.Tag.Id(),

		fromVersion: config.FromVersion,
		toVersion:   config.ToVersion,

		dbGetter: config.DBGetter,

		modelManagerService: config.ModelManagerService,
		upgradeService:      config.UpgradeService,

		logger: config.Logger,
		clock:  config.Clock,
	}

	if err := catacomb.Invoke(catacomb.Plan{
		Site: &w.catacomb,
		Work: w.loop,
	}); err != nil {
		return nil, errors.Trace(err)
	}

	return w, nil
}

// Kill implements worker.Worker.Kill.
func (w *upgradeDBWorker) Kill() {
	w.catacomb.Kill(nil)
}

// Wait implements worker.Worker.Wait.
func (w *upgradeDBWorker) Wait() error {
	return w.catacomb.Wait()
}

// loop implements Worker main loop.
func (w *upgradeDBWorker) loop() error {
	if w.upgradeDone() {
		// We're already upgraded, so we can uninstall this worker. This will
		// prevent it from running again, without an agent restart.
		return dependency.ErrUninstall
	}

	ctx, cancel := w.scopedContext()
	defer cancel()

	w.logger.Infof("creating upgrade from: %v to: %v", w.fromVersion, w.toVersion)

	// Create an upgrade for this controller. If another controller has already
	// created the upgrade, we will get an ErrUpgradeAlreadyStarted error. The
	// job of this controller is just to wait for the upgrade to be done and
	// then unlock the DBUpgradeCompleteLock.
	upgradeUUID, err := w.upgradeService.CreateUpgrade(ctx, w.fromVersion, w.toVersion)
	if err != nil {
		if errors.Is(err, upgradeerrors.ErrUpgradeAlreadyStarted) {
			// We're already running the upgrade, so we can just watch the
			// upgrade and wait for it to complete.
			w.logger.Tracef("upgrade already started, watching upgrade")
			return w.watchUpgrade()
		}
		return errors.Annotatef(err, "create upgrade from: %v to: %v", w.fromVersion, w.toVersion)
	}

	return w.runUpgrade(upgradeUUID)
}

// watchUpgrade watches the upgrade until it is complete.
// Once the upgrade is complete, the DBUpgradeCompleteLock is unlocked.
func (w *upgradeDBWorker) watchUpgrade() error {
	w.logger.Infof("watching upgrade from: %v to: %v", w.fromVersion, w.toVersion)

	ctx, cancel := w.scopedContext()
	defer cancel()

	modelUUID, err := w.upgradeService.ActiveUpgrade(ctx)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			// This currently no active upgrade, so we can't watch anything.
			// If this happens, it's probably in a bad state. We can't really
			// do anything about it, so we'll just bounce and hope that we
			// see if we've performed the upgrade already and that
			// we just didn't know about it in time.
			return dependency.ErrBounce
		}
		return errors.Trace(err)
	}

	completedWatcher, err := w.upgradeService.WatchForUpgradeState(ctx, modelUUID, upgrade.DBCompleted)
	if err != nil {
		return errors.Annotate(err, "watch completed upgrade")
	}

	if err := w.catacomb.Add(completedWatcher); err != nil {
		return errors.Trace(err)
	}

	failedWatcher, err := w.upgradeService.WatchForUpgradeState(ctx, modelUUID, upgrade.Error)
	if err != nil {
		return errors.Annotate(err, "watch failed upgrade")
	}

	for {
		select {
		case <-w.catacomb.Dying():
			return w.catacomb.ErrDying()

		case <-completedWatcher.Changes():
			// The upgrade is complete, so we can unlock the lock.
			w.logger.Infof("database upgrade complete")
			w.dbUpgradeCompleteLock.Unlock()
			return dependency.ErrUninstall

		case <-failedWatcher.Changes():
			// If the upgrade failed, we can't do anything about it, so we'll
			// just bounce and hope we get a better result next time.
			w.logger.Errorf("database upgrade failed, check logs for details")
			return dependency.ErrBounce
		}
	}
}

// upgradeDone returns true if this worker does not need to run any upgrade
// logic.
func (w *upgradeDBWorker) upgradeDone() bool {
	// If we are already unlocked, there is nothing to do.
	if w.dbUpgradeCompleteLock.IsUnlocked() {
		return true
	}

	if w.fromVersion == w.toVersion {
		w.logger.Infof("database upgrade for %v already completed", w.toVersion)
		w.dbUpgradeCompleteLock.Unlock()
		return true
	}

	return false
}

func (w *upgradeDBWorker) runUpgrade(upgradeUUID domainupgrade.UUID) error {
	w.logger.Infof("running database upgrade from: %v to: %v", w.fromVersion, w.toVersion)

	ctx, cancel := w.scopedContext()
	defer cancel()

	if err := w.upgradeService.SetControllerReady(ctx, upgradeUUID, w.controllerID); err != nil {
		return errors.Annotatef(err, "set controller ready")
	}

	// Watch for the upgrade to be ready. This should ensure that all
	// controllers are sync'd and waiting for the leader to start the upgrade.
	watcher, err := w.upgradeService.WatchForUpgradeReady(ctx, upgradeUUID)
	if err != nil {
		return errors.Trace(err)
	}

	for {
		select {
		case <-w.catacomb.Dying():
			return w.catacomb.ErrDying()

		case <-w.clock.After(defaultUpgradeTimeout):
			if err := w.upgradeService.SetDBUpgradeFailed(ctx, upgradeUUID); err != nil {
				return errors.Annotatef(err, "set db upgrade failed")
			}

			return errors.Errorf("timed out waiting for upgrade from: %v to: %v", w.fromVersion, w.toVersion)

		case <-watcher.Changes():
			if err := w.upgradeService.StartUpgrade(ctx, upgradeUUID); err != nil {
				return errors.Annotatef(err, "start upgrade")
			}

			// Upgrade the controller database first.
			if err := w.upgradeController(ctx); err != nil {
				return errors.Trace(err)
			}
			// Then upgrade the models databases.
			if err := w.upgradeModels(ctx); err != nil {
				return errors.Trace(err)
			}

			if err := w.upgradeService.SetDBUpgradeCompleted(ctx, upgradeUUID); err != nil {
				return errors.Annotatef(err, "set db upgrade completed")
			}

			w.logger.Infof("database upgrade already completed")
			w.dbUpgradeCompleteLock.Unlock()

			return nil
		}
	}
}

func (w *upgradeDBWorker) upgradeController(ctx context.Context) error {
	w.logger.Infof("upgrading controller database from: %v to: %v", w.fromVersion, w.toVersion)

	db, err := w.dbGetter.GetDB(coredatabase.ControllerNS)
	if err != nil {
		return errors.Annotatef(err, "controller db")
	}

	schema := schema.ControllerDDL()
	changeSet, err := schema.Ensure(ctx, db)
	if err != nil {
		return errors.Annotatef(err, "applying controller schema")
	}
	w.logger.Infof("applied controller schema changes from: %d to: %d", changeSet.Post, changeSet.Current)
	return nil
}

func (w *upgradeDBWorker) upgradeModels(ctx context.Context) error {
	w.logger.Infof("upgrading model databases from: %v to: %v", w.fromVersion, w.toVersion)

	models, err := w.modelManagerService.ModelList(ctx)
	if err != nil {
		return errors.Annotatef(err, "getting model list")
	}

	for _, modelUUID := range models {
		if err := w.upgradeModel(ctx, modelUUID); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (w *upgradeDBWorker) upgradeModel(ctx context.Context, modelUUID model.UUID) error {
	db, err := w.dbGetter.GetDB(modelUUID.String())
	if err != nil {
		return errors.Annotatef(err, "model db %s", modelUUID)
	}

	schema := schema.ModelDDL()
	changeSet, err := schema.Ensure(ctx, db)
	if err != nil {
		return errors.Annotatef(err, "applying model schema %s", modelUUID)
	}
	w.logger.Infof("applied model schema changes from: %d to: %d for model %s", changeSet.Post, changeSet.Current, modelUUID)
	return nil
}

func (w *upgradeDBWorker) scopedContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(w.catacomb.Context(context.Background()))
}
