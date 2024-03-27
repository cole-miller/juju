// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package application_test

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/juju/charm/v8"
	charmresource "github.com/juju/charm/v8/resource"
	csparams "github.com/juju/charmrepo/v6/csclient/params"
	"github.com/juju/errors"
	"github.com/juju/loggo"
	"github.com/juju/mgo/v2"
	"github.com/juju/names/v4"
	jtesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/v3"
	gc "gopkg.in/check.v1"

	unitassignerapi "github.com/juju/juju/api/agent/unitassigner"
	"github.com/juju/juju/api/client/client"
	"github.com/juju/juju/apiserver/common"
	commontesting "github.com/juju/juju/apiserver/common/testing"
	"github.com/juju/juju/apiserver/facades/client/application"
	apiservertesting "github.com/juju/juju/apiserver/testing"
	"github.com/juju/juju/core/arch"
	"github.com/juju/juju/core/constraints"
	"github.com/juju/juju/core/instance"
	"github.com/juju/juju/core/model"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/core/network/firewall"
	"github.com/juju/juju/core/status"
	jujutesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/rpc/params"
	"github.com/juju/juju/state"
	"github.com/juju/juju/state/stateenvirons"
	statestorage "github.com/juju/juju/state/storage"
	statetesting "github.com/juju/juju/state/testing"
	"github.com/juju/juju/storage"
	"github.com/juju/juju/storage/poolmanager"
	"github.com/juju/juju/testcharms"
	"github.com/juju/juju/testing"
	"github.com/juju/juju/testing/factory"
	jujuversion "github.com/juju/juju/version"
)

type applicationSuite struct {
	jujutesting.JujuConnSuite
	commontesting.BlockHelper

	applicationAPI *application.APIv15
	application    *state.Application
	authorizer     *apiservertesting.FakeAuthorizer
	repo           *mockRepo
}

var _ = gc.Suite(&applicationSuite{})

func (s *applicationSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)
	s.BlockHelper = commontesting.NewBlockHelper(s.APIState)
	s.AddCleanup(func(*gc.C) { s.BlockHelper.Close() })

	s.application = s.Factory.MakeApplication(c, nil)

	s.authorizer = &apiservertesting.FakeAuthorizer{
		Tag: s.AdminUserTag(c),
	}
	s.applicationAPI = s.makeAPI(c)

	var logger loggo.Logger
	s.repo = &mockRepo{
		CallMocker: jtesting.NewCallMocker(logger),
		revisions:  make(map[string]int),
	}

	s.PatchValue(&application.OpenCSRepo, func(args application.OpenCSRepoParams) (application.Repository, error) {
		return s.repo, nil
	})
}

func (s *applicationSuite) openRepo(args application.OpenCSRepoParams) (application.Repository, error) {
	return s.repo, nil
}

func (s *applicationSuite) UploadCharm(c *gc.C, url, name string) (*charm.URL, charm.Charm) {
	resultURL := charm.MustParseURL(url)
	if resultURL.User == "" {
		resultURL.User = "who"
	}
	if resultURL.Revision < 0 {
		base := *resultURL
		rev, ok := s.repo.revisions[base.String()]
		if !ok {
			resultURL.Revision = 0
		} else {
			resultURL.Revision = rev + 1
		}
		s.repo.revisions[base.String()] = resultURL.Revision
	}
	ch, err := charm.ReadCharmArchive(
		testcharms.RepoWithSeries("quantal").CharmArchivePath(c.MkDir(), name))
	c.Assert(err, jc.ErrorIsNil)
	s.repo.Call("DownloadCharm", resultURL.String()).Returns(
		ch,
		error(nil),
	)
	return resultURL, ch
}

func (s *applicationSuite) UploadCharmMultiSeries(c *gc.C, url, name string) (*charm.URL, charm.Charm) {
	return s.UploadCharm(c, url, name)
}

func (s *applicationSuite) makeAPI(c *gc.C) *application.APIv15 {
	resources := common.NewResources()
	c.Assert(resources.RegisterNamed("dataDir", common.StringResource(c.MkDir())), jc.ErrorIsNil)
	storageAccess, err := application.GetStorageState(s.State)
	c.Assert(err, jc.ErrorIsNil)
	model, err := s.State.Model()
	c.Assert(err, jc.ErrorIsNil)
	blockChecker := common.NewBlockChecker(s.State)
	registry := stateenvirons.NewStorageProviderRegistry(s.Environ)
	pm := poolmanager.New(state.NewStateSettings(s.State), registry)
	api, err := application.NewAPIBase(
		application.GetState(s.State),
		storageAccess,
		s.authorizer,
		nil,
		blockChecker,
		application.GetModel(model),
		nil, // leadership not used in these tests.
		application.CharmToStateCharm,
		application.DeployApplication,
		pm,
		registry,
		common.NewResources(),
		nil, // CAAS Broker not used in this suite.
	)
	c.Assert(err, jc.ErrorIsNil)
	return &application.APIv15{api}
}

func (s *applicationSuite) TestCharmConfig(c *gc.C) {
	s.setUpConfigTest(c)

	branchName := "test-branch"
	c.Assert(s.State.AddBranch(branchName, "test-user"), jc.ErrorIsNil)

	results, err := s.applicationAPI.CharmConfig(params.ApplicationGetArgs{
		Args: []params.ApplicationGet{
			{ApplicationName: "foo", BranchName: branchName},
			{ApplicationName: "bar", BranchName: branchName},
			{ApplicationName: "wat", BranchName: branchName},
		},
	})
	assertConfigTest(c, results, err, []params.ConfigResult{})
}

func (s *applicationSuite) TestCharmConfigV8(c *gc.C) {
	s.setUpConfigTest(c)
	api := &application.APIv8{
		APIv9: &application.APIv9{
			APIv10: &application.APIv10{
				APIv11: &application.APIv11{
					APIv12: &application.APIv12{
						APIv13: &application.APIv13{
							APIv14: &application.APIv14{
								s.applicationAPI,
							},
						},
					},
				},
			},
		},
	}
	results, err := api.CharmConfig(params.Entities{
		Entities: []params.Entity{
			{Tag: "wat"}, {Tag: "machine-0"}, {Tag: "user-foo"},
			{Tag: "application-foo"}, {Tag: "application-bar"}, {Tag: "application-wat"},
		},
	})
	assertConfigTest(c, results, err, []params.ConfigResult{
		{Error: &params.Error{Message: `"wat" is not a valid tag`}},
		{Error: &params.Error{Message: `unexpected tag type, expected application, got machine`}},
		{Error: &params.Error{Message: `unexpected tag type, expected application, got user`}},
	})
}

func (s *applicationSuite) TestGetConfig(c *gc.C) {
	s.setUpConfigTest(c)
	results, err := s.applicationAPI.GetConfig(params.Entities{
		Entities: []params.Entity{
			{Tag: "wat"}, {Tag: "machine-0"}, {Tag: "user-foo"},
			{Tag: "application-foo"}, {Tag: "application-bar"}, {Tag: "application-wat"},
		},
	})
	assertConfigTest(c, results, err, []params.ConfigResult{
		{Error: &params.Error{Message: `"wat" is not a valid tag`}},
		{Error: &params.Error{Message: `unexpected tag type, expected application, got machine`}},
		{Error: &params.Error{Message: `unexpected tag type, expected application, got user`}},
	})
}

func (s *applicationSuite) setUpConfigTest(c *gc.C) {
	fooConfig := map[string]interface{}{
		"title":       "foo",
		"skill-level": 42,
	}
	dummy := s.Factory.MakeCharm(c, &factory.CharmParams{
		Name: "dummy",
	})
	s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Name:        "foo",
		Charm:       dummy,
		CharmConfig: fooConfig,
	})
	barConfig := map[string]interface{}{
		"title":   "bar",
		"outlook": "fantastic",
	}
	s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Name:        "bar",
		Charm:       dummy,
		CharmConfig: barConfig,
	})
}

func assertConfigTest(c *gc.C, results params.ApplicationGetConfigResults, err error, resPrefix []params.ConfigResult) {
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, params.ApplicationGetConfigResults{
		Results: append(resPrefix, []params.ConfigResult{
			{
				Config: map[string]interface{}{
					"outlook": map[string]interface{}{
						"description": "No default outlook.",
						"source":      "unset",
						"type":        "string",
					},
					"skill-level": map[string]interface{}{
						"description": "A number indicating skill.",
						"source":      "user",
						"type":        "int",
						"value":       42,
					},
					"title": map[string]interface{}{
						"default":     "My Title",
						"description": "A descriptive title used for the application.",
						"source":      "user",
						"type":        "string",
						"value":       "foo",
					},
					"username": map[string]interface{}{
						"default":     "admin001",
						"description": "The name of the initial account (given admin permissions).",
						"source":      "default",
						"type":        "string",
						"value":       "admin001",
					},
				},
			}, {
				Config: map[string]interface{}{
					"outlook": map[string]interface{}{
						"description": "No default outlook.",
						"source":      "user",
						"type":        "string",
						"value":       "fantastic",
					},
					"skill-level": map[string]interface{}{
						"description": "A number indicating skill.",
						"source":      "unset",
						"type":        "int",
					},
					"title": map[string]interface{}{
						"default":     "My Title",
						"description": "A descriptive title used for the application.",
						"source":      "user",
						"type":        "string",
						"value":       "bar",
					},
					"username": map[string]interface{}{
						"default":     "admin001",
						"description": "The name of the initial account (given admin permissions).",
						"source":      "default",
						"type":        "string",
						"value":       "admin001",
					},
				},
			}, {
				Error: &params.Error{Message: `application "wat" not found`, Code: "not found"},
			},
		}...)})
}

func (s *applicationSuite) TestSetMetricCredentials(c *gc.C) {
	ch := s.Factory.MakeCharm(c, &factory.CharmParams{Name: "wordpress"})
	wordpress := s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Charm: ch,
	})
	tests := []struct {
		about   string
		args    params.ApplicationMetricCredentials
		results params.ErrorResults
	}{
		{
			"test one argument and it passes",
			params.ApplicationMetricCredentials{Creds: []params.ApplicationMetricCredential{{
				ApplicationName:   s.application.Name(),
				MetricCredentials: []byte("creds 1234"),
			}}},
			params.ErrorResults{Results: []params.ErrorResult{{Error: nil}}},
		},
		{
			"test two arguments and both pass",
			params.ApplicationMetricCredentials{Creds: []params.ApplicationMetricCredential{
				{
					ApplicationName:   s.application.Name(),
					MetricCredentials: []byte("creds 1234"),
				},
				{
					ApplicationName:   wordpress.Name(),
					MetricCredentials: []byte("creds 4567"),
				},
			}},
			params.ErrorResults{Results: []params.ErrorResult{
				{Error: nil},
				{Error: nil},
			}},
		},
		{
			"test two arguments and second one fails",
			params.ApplicationMetricCredentials{Creds: []params.ApplicationMetricCredential{
				{
					ApplicationName:   s.application.Name(),
					MetricCredentials: []byte("creds 1234"),
				},
				{
					ApplicationName:   "not-a-application",
					MetricCredentials: []byte("creds 4567"),
				},
			}},
			params.ErrorResults{Results: []params.ErrorResult{
				{Error: nil},
				{Error: &params.Error{Message: `application "not-a-application" not found`, Code: "not found"}},
			}},
		},
	}
	for i, t := range tests {
		c.Logf("Running test %d %v", i, t.about)
		results, err := s.applicationAPI.SetMetricCredentials(t.args)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(results.Results, gc.HasLen, len(t.results.Results))
		c.Assert(results, gc.DeepEquals, t.results)

		for i, a := range t.args.Creds {
			if t.results.Results[i].Error == nil {
				app, err := s.State.Application(a.ApplicationName)
				c.Assert(err, jc.ErrorIsNil)
				creds := app.MetricCredentials()
				c.Assert(creds, gc.DeepEquals, a.MetricCredentials)
			}
		}
	}
}

func (s *applicationSuite) TestCompatibleSettingsParsing(c *gc.C) {
	// Test the exported settings parsing in a compatible way.
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	app, err := s.State.Application("dummy")
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ch.String(), gc.Equals, "local:quantal/dummy-1")

	// Empty string will be returned as nil.
	options := map[string]string{
		"title":    "foobar",
		"username": "",
	}
	settings, err := application.ParseSettingsCompatible(ch.Config(), options)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, charm.Settings{
		"title":    "foobar",
		"username": nil,
	})

	// Illegal settings lead to an error.
	options = map[string]string{
		"yummy": "didgeridoo",
	}
	_, err = application.ParseSettingsCompatible(ch.Config(), options)
	c.Assert(err, gc.ErrorMatches, `unknown option "yummy"`)
}

func (s *applicationSuite) TestApplicationDeployWithStorage(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:utopic/storage-block-10", "storage-block")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	storageConstraints := map[string]storage.Constraints{
		"data": {
			Count: 1,
			Size:  1024,
			Pool:  "modelscoped-block",
		},
	}

	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Storage:         storageConstraints,
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.DeepEquals, params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	})
	app := apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, s.constraintsWithDefaultArch(c))
	storageConstraintsOut, err := app.StorageConstraints()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(storageConstraintsOut, gc.DeepEquals, map[string]state.StorageConstraints{
		"data": {
			Count: 1,
			Size:  1024,
			Pool:  "modelscoped-block",
		},
		"allecto": {
			Count: 0,
			Size:  1024,
			Pool:  "loop",
		},
	})
}

func (s *applicationSuite) TestMinJujuVersionTooHigh(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/minjujuversion-0", "minjujuversion")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	vers := jujuversion.Current
	vers.Tag = ""
	match := fmt.Sprintf(`charm's min version (999.999.999) is higher than this juju model's version (%s)`, vers)
	c.Assert(err, gc.ErrorMatches, regexp.QuoteMeta(match))
}

func (s *applicationSuite) TestApplicationDeployWithInvalidStoragePool(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:utopic/storage-block-0", "storage-block")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	storageConstraints := map[string]storage.Constraints{
		"data": {
			Pool:  "foo",
			Count: 1,
			Size:  1024,
		},
	}

	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Storage:         storageConstraints,
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.ErrorMatches, `.* pool "foo" not found`)
}

func (s *applicationSuite) TestApplicationDeployDefaultFilesystemStorage(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:trusty/storage-filesystem-1", "storage-filesystem")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.DeepEquals, params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	})
	app := apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, s.constraintsWithDefaultArch(c))
	storageConstraintsOut, err := app.StorageConstraints()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(storageConstraintsOut, gc.DeepEquals, map[string]state.StorageConstraints{
		"data": {
			Count: 1,
			Size:  1024,
			Pool:  "rootfs",
		},
	})
}

func (s *applicationSuite) TestApplicationDeploy(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Placement: []*instance.Placement{
			{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "valid"},
		},
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.DeepEquals, params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	})
	app := apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, s.constraintsWithDefaultArch(c))
	units, err := app.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 1)
}

func (s *applicationSuite) constraintsWithDefaultArch(c *gc.C) constraints.Value {
	a := arch.DefaultArchitecture
	return constraints.Value{
		Arch: &a,
	}
}

func (s *applicationSuite) TestApplicationDeployWithInvalidPlacement(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Placement: []*instance.Placement{
			{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "invalid"},
		},
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.NotNil)
	c.Assert(results.Results[0].Error.Error(), gc.Matches, ".* invalid placement is invalid")
}

func (s *applicationSuite) TestApplicationDeployWithMachinePlacementLockedError(c *gc.C) {
	s.testApplicationDeployWithPlacementLockedError(c, instance.Placement{Scope: "#", Directive: "0"}, false)
}

func (s *applicationSuite) TestApplicationDeployWithMachineContainerPlacementLockedError(c *gc.C) {
	s.testApplicationDeployWithPlacementLockedError(c, instance.Placement{Scope: "lxd", Directive: "0"}, false)
}

func (s *applicationSuite) TestApplicationDeployWithExtantMachineContainerLockedParentError(c *gc.C) {
	s.testApplicationDeployWithPlacementLockedError(c, instance.Placement{Scope: "#", Directive: "0/lxd/0"}, true)
}

func (s *applicationSuite) testApplicationDeployWithPlacementLockedError(
	c *gc.C, placement instance.Placement, addContainer bool,
) {
	m, err := s.BackingState.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)

	if addContainer {
		template := state.MachineTemplate{
			Series: "xenial",
			Jobs:   []state.MachineJob{state.JobHostUnits},
		}
		_, err := s.State.AddMachineInsideMachine(template, m.Id(), "lxd")
		c.Assert(err, jc.ErrorIsNil)
	}

	c.Assert(m.CreateUpgradeSeriesLock(nil, "trusty"), jc.ErrorIsNil)

	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Placement:       []*instance.Placement{&placement},
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.NotNil)
	c.Assert(results.Results[0].Error.Error(), gc.Matches, ".* machine is locked for series upgrade")
}

func (s *applicationSuite) TestApplicationDeploymentRemovesPendingResourcesOnFailure(c *gc.C) {
	charm := s.AddTestingCharm(c, "dummy-resource")
	resources := s.State.Resources()
	pendingID, err := resources.AddPendingResource("haha/borken", "user", charmresource.Resource{
		Meta:   charm.Meta().Resources["dummy"],
		Origin: charmresource.OriginUpload,
	})
	c.Assert(err, jc.ErrorIsNil)

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			ApplicationName: "haha/borken",
			NumUnits:        1,
			CharmURL:        charm.String(),
			CharmOrigin:     createCharmOriginFromURL(c, charm.URL()),
			Resources:       map[string]string{"dummy": pendingID},
		}},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.ErrorMatches, `cannot add application "haha/borken": invalid name`)

	res, err := resources.ListPendingResources("haha/borken")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(res.Resources, gc.HasLen, 0)
	c.Assert(res.UnitResources, gc.HasLen, 0)
}

func (s *applicationSuite) TestApplicationDeploymentLeavesResourcesOnSuccess(c *gc.C) {
	charm := s.AddTestingCharm(c, "dummy-resource")
	resources := s.State.Resources()
	pendingID, err := resources.AddPendingResource("unborken", "user", charmresource.Resource{
		Meta:   charm.Meta().Resources["dummy"],
		Origin: charmresource.OriginUpload,
	})
	c.Assert(err, jc.ErrorIsNil)

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			ApplicationName: "unborken",
			NumUnits:        1,
			CharmURL:        charm.String(),
			CharmOrigin:     createCharmOriginFromURL(c, charm.URL()),
			Resources:       map[string]string{"dummy": pendingID},
		}},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	res, err := resources.ListResources("unborken")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(res.Resources, gc.HasLen, 1)
}

func (s *applicationSuite) TestApplicationDeploymentWithTrust(c *gc.C) {
	// This test should fail if the configuration parsing does not
	// understand the "trust" configuration parameter
	curl, ch := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	config := map[string]string{"trust": "true"}
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Config:          config,
		Constraints:     cons,
		Placement: []*instance.Placement{
			{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "valid"},
		},
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.DeepEquals, params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	})

	app := apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, s.constraintsWithDefaultArch(c))

	appConfig, err := app.ApplicationConfig()
	c.Assert(err, jc.ErrorIsNil)

	trust := appConfig.GetBool("trust", false)
	c.Assert(trust, jc.IsTrue)
}

func (s *applicationSuite) TestApplicationDeploymentNoTrust(c *gc.C) {
	// This test should fail if the trust configuration setting defaults to
	// anything other than "false" when no configuration parameter for trust
	// is set at deployment.
	curl, ch := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     createCharmOriginFromURL(c, curl),
		NumUnits:        1,
		Constraints:     cons,
		Placement: []*instance.Placement{
			{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "valid"},
		},
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, gc.DeepEquals, params.ErrorResults{
		Results: []params.ErrorResult{{Error: nil}},
	})

	app := apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, s.constraintsWithDefaultArch(c))
	appConfig, err := app.ApplicationConfig()
	c.Assert(err, jc.ErrorIsNil)
	trust := appConfig.GetBool(application.TrustConfigOptionName, true)
	c.Assert(trust, jc.IsFalse)
}

func (s *applicationSuite) testClientApplicationsDeployWithBindings(c *gc.C, endpointBindings, expected map[string]string) {
	curl, _ := s.UploadCharm(c, "cs:quantal/riak-42", "riak")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	var cons constraints.Value
	args := params.ApplicationDeploy{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin: &params.CharmOrigin{
			Source: "charm-store",
			Base: params.Base{
				Name: "ubuntu", Channel: "12.10",
			},
		},
		NumUnits:         1,
		Constraints:      cons,
		EndpointBindings: endpointBindings,
	}

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{args}},
	)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	app, err := s.State.Application(args.ApplicationName)
	c.Assert(err, jc.ErrorIsNil)

	retrievedBindings, err := app.EndpointBindings()
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(retrievedBindings.Map(), jc.DeepEquals, expected)
}

func (s *applicationSuite) TestClientApplicationsDeployWithOldBindings(c *gc.C) {
	space, err := s.State.AddSpace("a-space", "", nil, true)
	c.Assert(err, jc.ErrorIsNil)
	expected := map[string]string{
		"":         network.AlphaSpaceId,
		"endpoint": space.Id(),
		"ring":     network.AlphaSpaceId,
		"admin":    network.AlphaSpaceId,
	}
	endpointBindings := map[string]string{
		"endpoint": space.Name(),
		"ring":     "",
		"admin":    "",
	}
	s.testClientApplicationsDeployWithBindings(c, endpointBindings, expected)
}

func (s *applicationSuite) TestClientApplicationsDeployWithBindings(c *gc.C) {
	space, err := s.State.AddSpace("a-space", "", nil, true)
	c.Assert(err, jc.ErrorIsNil)
	expected := map[string]string{
		"":         network.AlphaSpaceId,
		"endpoint": space.Id(),
		"ring":     network.AlphaSpaceId,
		"admin":    network.AlphaSpaceId,
	}
	endpointBindings := map[string]string{"endpoint": space.Id()}
	s.testClientApplicationsDeployWithBindings(c, endpointBindings, expected)
}

func (s *applicationSuite) TestClientApplicationsDeployWithDefaultBindings(c *gc.C) {
	expected := map[string]string{
		"":         network.AlphaSpaceId,
		"endpoint": network.AlphaSpaceId,
		"ring":     network.AlphaSpaceId,
		"admin":    network.AlphaSpaceId,
	}
	s.testClientApplicationsDeployWithBindings(c, nil, expected)
}

// TODO(wallyworld) - the following charm tests have been moved from the apiserver/client
// package in order to use the fake charm store testing infrastructure. They are legacy tests
// written to use the api client instead of the apiserver logic. They need to be rewritten and
// feature tests added.

func (s *applicationSuite) TestAddCharm(c *gc.C) {
	var blobs blobs
	s.PatchValue(application.NewStateStorage, func(uuid string, session *mgo.Session) statestorage.Storage {
		storage := statestorage.NewStorage(uuid, session)
		return &recordingStorage{Storage: storage, blobs: &blobs}
	})

	client := client.NewClient(s.APIState, testing.NoopLogger{})
	// First test the sanity checks.
	err := client.AddCharm(&charm.URL{Name: "nonsense"}, csparams.StableChannel, false)
	c.Assert(err, gc.ErrorMatches, `cannot parse charm or bundle URL: ":nonsense-0"`)
	err = client.AddCharm(charm.MustParseURL("local:precise/dummy"), csparams.StableChannel, false)
	c.Assert(err, gc.ErrorMatches, "only charm store charm URLs are supported, with cs: schema")
	err = client.AddCharm(charm.MustParseURL("cs:precise/wordpress"), csparams.StableChannel, false)
	c.Assert(err, gc.ErrorMatches, "charm URL must include a revision")

	// Add a charm, without uploading it to storage, to
	// check that AddCharm does not try to do it.
	charmDir := testcharms.Repo.CharmDir("dummy")
	ident := fmt.Sprintf("%s-%d", charmDir.Meta().Name, charmDir.Revision())
	curl := charm.MustParseURL("cs:quantal/" + ident)
	info := state.CharmInfo{
		Charm:       charmDir,
		ID:          curl,
		StoragePath: "/storage/path",
		SHA256:      ident + "-sha256",
	}
	sch, err := s.State.AddCharm(info)
	c.Assert(err, jc.ErrorIsNil)

	// AddCharm should see the charm in state and not upload it.
	err = client.AddCharm(sch.URL(), csparams.StableChannel, false)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(blobs.m, gc.HasLen, 0)

	// Now try adding another charm completely.
	curl, _ = s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err = client.AddCharm(curl, csparams.StableChannel, false)
	c.Assert(err, jc.ErrorIsNil)

	// Verify it's in state and it got uploaded.
	storage := statestorage.NewStorage(s.State.ModelUUID(), s.State.MongoSession())
	sch, err = s.State.Charm(curl)
	c.Assert(err, jc.ErrorIsNil)
	s.assertUploaded(c, storage, sch.StoragePath(), sch.BundleSha256())
}

func (s *applicationSuite) TestAddCharmConcurrently(c *gc.C) {
	c.Skip("see lp:1596960 -- bad test for bad code")

	var putBarrier sync.WaitGroup
	var blobs blobs
	s.PatchValue(application.NewStateStorage, func(uuid string, session *mgo.Session) statestorage.Storage {
		storage := statestorage.NewStorage(uuid, session)
		return &recordingStorage{Storage: storage, blobs: &blobs, putBarrier: &putBarrier}
	})

	client := client.NewClient(s.APIState, testing.NoopLogger{})
	curl, _ := s.UploadCharm(c, "trusty/wordpress-3", "wordpress")

	// Try adding the same charm concurrently from multiple goroutines
	// to test no "duplicate key errors" are reported (see lp bug
	// #1067979) and also at the end only one charm document is
	// created.

	var wg sync.WaitGroup
	// We don't add them 1-by-1 because that would allow each goroutine to
	// finish separately without actually synchronizing between them
	putBarrier.Add(10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			c.Assert(client.AddCharm(curl, csparams.StableChannel, false), gc.IsNil, gc.Commentf("goroutine %d", index))
			sch, err := s.State.Charm(curl)
			c.Assert(err, gc.IsNil, gc.Commentf("goroutine %d", index))
			c.Assert(sch.URL(), jc.DeepEquals, curl, gc.Commentf("goroutine %d", index))
		}(i)
	}
	wg.Wait()

	blobs.Lock()

	c.Assert(blobs.m, gc.HasLen, 10)

	// Verify there is only a single uploaded charm remains and it
	// contains the correct data.
	sch, err := s.State.Charm(curl)
	c.Assert(err, jc.ErrorIsNil)
	storagePath := sch.StoragePath()
	c.Assert(blobs.m[storagePath], jc.IsTrue)
	for path, exists := range blobs.m {
		if path != storagePath {
			c.Assert(exists, jc.IsFalse)
		}
	}

	storage := statestorage.NewStorage(s.State.ModelUUID(), s.State.MongoSession())
	s.assertUploaded(c, storage, sch.StoragePath(), sch.BundleSha256())
}

func (s *applicationSuite) assertUploaded(c *gc.C, storage statestorage.Storage, storagePath, expectedSHA256 string) {
	reader, _, err := storage.Get(storagePath)
	c.Assert(err, jc.ErrorIsNil)
	defer reader.Close()
	downloadedSHA256, _, err := utils.ReadSHA256(reader)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(downloadedSHA256, gc.Equals, expectedSHA256)
}

func (s *applicationSuite) TestAddCharmOverwritesPlaceholders(c *gc.C) {
	client := client.NewClient(s.APIState, testing.NoopLogger{})
	curl, _ := s.UploadCharm(c, "cs:trusty/wordpress-42", "wordpress")

	// Add a placeholder with the same charm URL.
	err := s.State.AddCharmPlaceholder(curl)
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.Charm(curl)
	c.Assert(err, jc.Satisfies, errors.IsNotFound)

	// Now try to add the charm, which will convert the placeholder to
	// a pending charm.
	err = client.AddCharm(curl, csparams.StableChannel, false)
	c.Assert(err, jc.ErrorIsNil)

	// Make sure the document's flags were reset as expected.
	sch, err := s.State.Charm(curl)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sch.URL(), jc.DeepEquals, curl)
	c.Assert(sch.IsPlaceholder(), jc.IsFalse)
	c.Assert(sch.IsUploaded(), jc.IsTrue)
}

func (s *applicationSuite) TestApplicationSetCharm(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	numUnits := 3
	for i := 0; i < numUnits; i++ {
		_, err := s.State.AddMachine("quantal", state.JobHostUnits)
		c.Assert(err, jc.ErrorIsNil)
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL: curl.String(),
			CharmOrigin: &params.CharmOrigin{
				Source: "charm-store",
				Base: params.Base{
					Name: "ubuntu", Channel: "12.10",
				},
			},
			ApplicationName: "application",
			NumUnits:        numUnits,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	curl, _ = s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{
		names.NewUnitTag("application/0"),
		names.NewUnitTag("application/1"),
		names.NewUnitTag("application/2"),
	})
	c.Assert(errs, gc.DeepEquals, []error{error(nil), error(nil), error(nil)})
	c.Assert(err, jc.ErrorIsNil)
	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin: &params.CharmOrigin{
			Source: "charm-store",
			Base: params.Base{
				Name: "ubuntu", Channel: "12.10",
			},
		},
	})
	c.Assert(err, jc.ErrorIsNil)

	// Ensure that the charm is not marked as forced.
	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)
	charm, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(charm.String(), gc.Equals, curl.String())
	c.Assert(force, jc.IsFalse)
}

func (s *applicationSuite) setupApplicationSetCharm(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	numUnits := 3
	for i := 0; i < numUnits; i++ {
		_, err := s.State.AddMachine("quantal", state.JobHostUnits)
		c.Assert(err, jc.ErrorIsNil)
	}
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL: curl.String(),
			CharmOrigin: &params.CharmOrigin{
				Source: "charm-store",
				Base: params.Base{
					Name: "ubuntu", Channel: "12.10",
				},
			},
			ApplicationName: "application",
			NumUnits:        numUnits,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{
		names.NewUnitTag("application/0"),
		names.NewUnitTag("application/1"),
		names.NewUnitTag("application/2"),
	})
	c.Assert(errs, gc.DeepEquals, []error{error(nil), error(nil), error(nil)})
	c.Assert(err, jc.ErrorIsNil)
	curl, _ = s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *applicationSuite) assertApplicationSetCharm(c *gc.C, forceUnits bool) {
	err := s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        "cs:~who/quantal/wordpress-3",
		CharmOrigin:     &params.CharmOrigin{Source: "charm-store", OS: "ubuntu", Channel: "12.10"},
		ForceUnits:      forceUnits,
	})
	c.Assert(err, jc.ErrorIsNil)
	// Ensure that the charm is not marked as forced.
	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)
	charm, _, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(charm.String(), gc.Equals, "cs:~who/quantal/wordpress-3")
}

func (s *applicationSuite) assertApplicationSetCharmBlocked(c *gc.C, msg string) {
	err := s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        "cs:~who/quantal/wordpress-3",
		CharmOrigin: &params.CharmOrigin{
			Source: "charm-store",
			Base: params.Base{
				Name: "ubuntu", Channel: "12.10",
			},
		},
	})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) TestBlockDestroyApplicationSetCharm(c *gc.C) {
	s.setupApplicationSetCharm(c)
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationSetCharm")
	s.assertApplicationSetCharm(c, false)
}

func (s *applicationSuite) TestBlockRemoveApplicationSetCharm(c *gc.C) {
	s.setupApplicationSetCharm(c)
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationSetCharm")
	s.assertApplicationSetCharm(c, false)
}

func (s *applicationSuite) TestBlockChangesApplicationSetCharm(c *gc.C) {
	s.setupApplicationSetCharm(c)
	s.BlockAllChanges(c, "TestBlockChangesApplicationSetCharm")
	s.assertApplicationSetCharmBlocked(c, "TestBlockChangesApplicationSetCharm")
}

func (s *applicationSuite) TestApplicationSetCharmForceUnits(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	numUnits := 3
	for i := 0; i < numUnits; i++ {
		_, err := s.State.AddMachine("quantal", state.JobHostUnits)
		c.Assert(err, jc.ErrorIsNil)
	}
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL: curl.String(),
			CharmOrigin: &params.CharmOrigin{
				Source: "charm-store",
				Base: params.Base{
					Name: "ubuntu", Channel: "12.10",
				},
			},
			ApplicationName: "application",
			NumUnits:        numUnits,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	curl, _ = s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{
		names.NewUnitTag("application/0"),
		names.NewUnitTag("application/1"),
		names.NewUnitTag("application/2"),
	})
	c.Assert(errs, gc.DeepEquals, []error{error(nil), error(nil), error(nil)})
	c.Assert(err, jc.ErrorIsNil)
	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     &params.CharmOrigin{Source: "charm-store", OS: "ubuntu", Channel: "12.10"},
		ForceUnits:      true,
	})
	c.Assert(err, jc.ErrorIsNil)

	// Ensure that the charm is marked as forced.
	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)
	charm, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(charm.String(), gc.Equals, curl.String())
	c.Assert(force, jc.IsTrue)
}

func (s *applicationSuite) TestBlockApplicationSetCharmForce(c *gc.C) {
	s.setupApplicationSetCharm(c)

	// block all changes
	s.BlockAllChanges(c, "TestBlockApplicationSetCharmForce")
	s.BlockRemoveObject(c, "TestBlockApplicationSetCharmForce")
	s.BlockDestroyModel(c, "TestBlockApplicationSetCharmForce")

	s.assertApplicationSetCharm(c, true)
}

func (s *applicationSuite) TestApplicationSetCharmInvalidApplication(c *gc.C) {
	err := s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "badapplication",
		CharmURL:        "cs:quantal/wordpress-3",
		CharmOrigin: &params.CharmOrigin{
			Source: "charm-store",
			Base: params.Base{
				Name: "ubuntu", Channel: "12.10",
			},
		}, ForceSeries: true,
		ForceUnits: true,
	})
	c.Assert(err, gc.ErrorMatches, `application "badapplication" not found`)
}

func (s *applicationSuite) TestApplicationAddCharmErrors(c *gc.C) {
	for url, expect := range map[string]string{
		"wordpress":                   "only charm store charm URLs are supported, with cs: schema",
		"cs:wordpress":                "charm URL must include a revision",
		"cs:precise/wordpress":        "charm URL must include a revision",
		"cs:precise/wordpress-999999": `cannot retrieve "cs:precise/wordpress-999999": charm not found`,
	} {
		c.Logf("test %s", url)
		err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
			URL: url,
		}, s.openRepo)
		c.Check(err, gc.ErrorMatches, expect)
	}
}

func (s *applicationSuite) TestApplicationSetCharmLegacy(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL: curl.String(),
			CharmOrigin: &params.CharmOrigin{
				Source: "charm-store",
				Base: params.Base{
					Name: "ubuntu", Channel: "12.10",
				},
			},
			ApplicationName: "application",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	curl, _ = s.UploadCharm(c, "cs:trusty/dummy-1", "dummy")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	// Even with forceSeries = true, we can't change a charm where
	// the series is specified in the URL.
	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin: &params.CharmOrigin{
			Source: "charm-store",
			Base: params.Base{
				Name: "ubuntu", Channel: "12.10",
			},
		}, ForceSeries: true,
	})
	c.Assert(err, gc.ErrorMatches, `cannot upgrade application "application" to charm "cs:~who/trusty/dummy-1": cannot change an application's series`)
}

func (s *applicationSuite) TestApplicationSetCharmUnsupportedSeries(c *gc.C) {
	curl, _ := s.UploadCharmMultiSeries(c, "cs:~who/multi-series", "multi-series")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     &params.CharmOrigin{Source: "charm-store", Series: "bionic"},
			ApplicationName: "application",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	curl, _ = s.UploadCharmMultiSeries(c, "cs:~who/multi-series", "multi-series2")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     &params.CharmOrigin{Source: "charm-store", OS: "ubuntu", Channel: "12.10"},
	})
	c.Assert(err, gc.ErrorMatches, `cannot upgrade application "application" to charm "cs:~who/multi-series-1": only these series are supported: trusty, wily`)
}

func (s *applicationSuite) assertApplicationSetCharmSeries(c *gc.C, upgradeCharm, series string) {
	curl, _ := s.UploadCharmMultiSeries(c, "cs:~who/multi-series", "multi-series")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     &params.CharmOrigin{Source: "charm-store", Series: "quantal"},
			ApplicationName: "application",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	url := upgradeCharm
	if series != "" {
		url = series + "/" + upgradeCharm
	}
	curl, _ = s.UploadCharmMultiSeries(c, "cs:~who/"+url, upgradeCharm)
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     &params.CharmOrigin{Source: "charm-store", OS: "ubuntu", Channel: "12.10"},
		ForceSeries:     true,
	})
	c.Assert(err, jc.ErrorIsNil)
	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ch.String(), gc.Equals, "cs:~who/"+url+"-0")
}

func (s *applicationSuite) TestApplicationSetCharmUnsupportedSeriesForce(c *gc.C) {
	s.assertApplicationSetCharmSeries(c, "multi-series2", "")
}

func (s *applicationSuite) TestApplicationSetCharmNoExplicitSupportedSeries(c *gc.C) {
	s.assertApplicationSetCharmSeries(c, "dummy", "quantal")
}

func (s *applicationSuite) TestApplicationSetCharmWrongOS(c *gc.C) {
	curl, _ := s.UploadCharmMultiSeries(c, "cs:~who/multi-series", "multi-series")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     &params.CharmOrigin{Source: "charm-store", Series: "bionic"},
			ApplicationName: "application",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	curl, _ = s.UploadCharmMultiSeries(c, "cs:~who/multi-series-windows", "multi-series-windows")
	err = application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	err = s.applicationAPI.SetCharm(params.ApplicationSetCharm{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		CharmOrigin:     &params.CharmOrigin{Source: "charm-store", OS: "ubuntu", Channel: "12.10"},
		ForceSeries:     true,
	})
	c.Assert(err, gc.ErrorMatches, `cannot upgrade application "application" to charm "cs:~who/multi-series-windows-0": OS "Ubuntu" not supported by charm`)
}

func (s *applicationSuite) setupApplicationDeploy(c *gc.C, args string) (*charm.URL, charm.Charm, constraints.Value) {
	curl, ch := s.UploadCharm(c, "cs:quantal/dummy-42", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	cons := constraints.MustParse(args)
	return curl, ch, cons
}

func createCharmOriginFromURL(c *gc.C, curl *charm.URL) *params.CharmOrigin {
	switch curl.Schema {
	case "cs":
		return &params.CharmOrigin{Source: "charm-store", Series: curl.Series}
	case "local":
		return &params.CharmOrigin{Source: "local", Series: curl.Series}
	default:
		return &params.CharmOrigin{Source: "charm-hub", Series: curl.Series}
	}
}

func (s *applicationSuite) assertApplicationDeployPrincipal(c *gc.C, curl *charm.URL, ch charm.Charm, mem4g constraints.Value) {
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application",
			NumUnits:        3,
			Constraints:     mem4g,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
	apiservertesting.AssertPrincipalApplicationDeployed(c, s.State, "application", curl, false, ch, mem4g)
}

func (s *applicationSuite) assertApplicationDeployPrincipalBlocked(c *gc.C, msg string, curl *charm.URL, mem4g constraints.Value) {
	_, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application",
			NumUnits:        3,
			Constraints:     mem4g,
		}}})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) TestBlockDestroyApplicationDeployPrincipal(c *gc.C) {
	curl, bundle, cons := s.setupApplicationDeploy(c, "arch=amd64 mem=4G")
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationDeployPrincipal")
	s.assertApplicationDeployPrincipal(c, curl, bundle, cons)
}

func (s *applicationSuite) TestBlockRemoveApplicationDeployPrincipal(c *gc.C) {
	curl, bundle, cons := s.setupApplicationDeploy(c, "arch=amd64 mem=4G")
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationDeployPrincipal")
	s.assertApplicationDeployPrincipal(c, curl, bundle, cons)
}

func (s *applicationSuite) TestBlockChangesApplicationDeployPrincipal(c *gc.C) {
	curl, _, cons := s.setupApplicationDeploy(c, "mem=4G")
	s.BlockAllChanges(c, "TestBlockChangesApplicationDeployPrincipal")
	s.assertApplicationDeployPrincipalBlocked(c, "TestBlockChangesApplicationDeployPrincipal", curl, cons)
}

func (s *applicationSuite) TestApplicationDeploySubordinate(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:utopic/logging-47", "logging")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	app, err := s.State.Application("application-name")
	c.Assert(err, jc.ErrorIsNil)
	charm, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(force, jc.IsFalse)
	c.Assert(charm.URL(), gc.DeepEquals, curl)
	c.Assert(charm.Meta(), gc.DeepEquals, ch.Meta())
	c.Assert(charm.Config(), gc.DeepEquals, ch.Config())

	units, err := app.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 0)
}

func (s *applicationSuite) combinedSettings(ch *state.Charm, inSettings charm.Settings) charm.Settings {
	result := ch.Config().DefaultSettings()
	for name, value := range inSettings {
		result[name] = value
	}
	return result
}

func (s *applicationSuite) TestApplicationDeployConfig(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
			NumUnits:        1,
			ConfigYAML:      "application-name:\n  username: fred",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	app, err := s.State.Application("application-name")
	c.Assert(err, jc.ErrorIsNil)
	settings, err := app.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{"username": "fred"}))
}

func (s *applicationSuite) TestApplicationDeployConfigError(c *gc.C) {
	// TODO(fwereade): test Config/ConfigYAML handling directly on srvClient.
	// Can't be done cleanly until it's extracted similarly to Machiner.
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
			NumUnits:        1,
			ConfigYAML:      "application-name:\n  skill-level: fred",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.ErrorMatches, `option "skill-level" expected int, got "fred"`)
	_, err = s.State.Application("application-name")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func (s *applicationSuite) TestApplicationDeployToMachine(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:quantal/dummy-0", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	machine, err := s.State.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)

	arch := arch.DefaultArchitecture
	hwChar := &instance.HardwareCharacteristics{
		Arch: &arch,
	}
	instId := instance.Id("i-host-machine")
	err = machine.SetProvisioned(instId, "", "fake-nonce", hwChar)
	c.Assert(err, jc.ErrorIsNil)

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
			NumUnits:        1,
			ConfigYAML:      "application-name:\n  username: fred",
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	app, err := s.State.Application("application-name")
	c.Assert(err, jc.ErrorIsNil)
	charm, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(force, jc.IsFalse)
	c.Assert(charm.URL(), gc.DeepEquals, curl)
	c.Assert(charm.Meta(), gc.DeepEquals, ch.Meta())
	c.Assert(charm.Config(), gc.DeepEquals, ch.Config())

	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{names.NewUnitTag("application-name/0")})
	c.Assert(errs, gc.DeepEquals, []error{nil})
	c.Assert(err, jc.ErrorIsNil)

	units, err := app.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 1)

	mid, err := units[0].AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mid, gc.Equals, machine.Id())
}

func (s *applicationSuite) TestApplicationDeployToMachineWithLXDProfile(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:quantal/lxd-profile-0", "lxd-profile")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	machine, err := s.State.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)

	arch := arch.DefaultArchitecture
	hwChar := &instance.HardwareCharacteristics{
		Arch: &arch,
	}
	instId := instance.Id("i-host-machine")
	err = machine.SetProvisioned(instId, "", "fake-nonce", hwChar)
	c.Assert(err, jc.ErrorIsNil)

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
			NumUnits:        1,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	application, err := s.State.Application("application-name")
	c.Assert(err, jc.ErrorIsNil)
	expected, force, err := application.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(force, jc.IsFalse)
	c.Assert(expected.URL(), gc.DeepEquals, curl)
	c.Assert(expected.Meta(), gc.DeepEquals, ch.Meta())
	c.Assert(expected.Config(), gc.DeepEquals, ch.Config())

	expectedProfile := ch.(charm.LXDProfiler).LXDProfile()
	c.Assert(expected.LXDProfile(), gc.DeepEquals, &state.LXDProfile{
		Description: expectedProfile.Description,
		Config:      expectedProfile.Config,
		Devices:     expectedProfile.Devices,
	})

	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{names.NewUnitTag("application-name/0")})
	c.Assert(errs, gc.DeepEquals, []error{nil})
	c.Assert(err, jc.ErrorIsNil)

	units, err := application.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 1)

	mid, err := units[0].AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mid, gc.Equals, machine.Id())
}

func (s *applicationSuite) TestApplicationDeployToMachineWithInvalidLXDProfile(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/lxd-profile-fail-0", "lxd-profile-fail")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, gc.ErrorMatches, `.*invalid lxd-profile.yaml: contains device type "unix-disk"`)
}

func (s *applicationSuite) TestApplicationDeployToMachineWithInvalidLXDProfileAndForceStillSucceeds(c *gc.C) {
	curl, ch := s.UploadCharm(c, "cs:quantal/lxd-profile-fail-0", "lxd-profile-fail")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL:   curl.String(),
		Force: true,
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	machine, err := s.State.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)

	arch := arch.DefaultArchitecture
	hwChar := &instance.HardwareCharacteristics{
		Arch: &arch,
	}
	instId := instance.Id("i-host-machine")
	err = machine.SetProvisioned(instId, "", "fake-nonce", hwChar)
	c.Assert(err, jc.ErrorIsNil)

	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application-name",
			NumUnits:        1,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)

	app, err := s.State.Application("application-name")
	c.Assert(err, jc.ErrorIsNil)
	expected, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(force, jc.IsFalse)
	c.Assert(expected.URL(), gc.DeepEquals, curl)
	c.Assert(expected.Meta(), gc.DeepEquals, ch.Meta())
	c.Assert(expected.Config(), gc.DeepEquals, ch.Config())

	expectedProfile := ch.(charm.LXDProfiler).LXDProfile()
	c.Assert(expected.LXDProfile(), gc.DeepEquals, &state.LXDProfile{
		Description: expectedProfile.Description,
		Config:      expectedProfile.Config,
		Devices:     expectedProfile.Devices,
	})

	errs, err := unitassignerapi.New(s.APIState).AssignUnits([]names.UnitTag{names.NewUnitTag("application-name/0")})
	c.Assert(errs, gc.DeepEquals, []error{nil})
	c.Assert(err, jc.ErrorIsNil)

	units, err := app.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 1)

	mid, err := units[0].AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mid, gc.Equals, machine.Id())
}

func (s *applicationSuite) TestApplicationDeployToMachineNotFound(c *gc.C) {
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL: "cs:quantal/application-name-1",
			CharmOrigin: &params.CharmOrigin{
				Source: "charm-store",
				Base: params.Base{
					Name: "ubuntu", Channel: "12.10",
				},
			},
			ApplicationName: "application-name",
			NumUnits:        1,
			Placement:       []*instance.Placement{instance.MustParsePlacement("42")},
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.ErrorMatches, `cannot deploy "application-name" to machine 42: machine 42 not found`)

	_, err = s.State.Application("application-name")
	c.Assert(err, gc.ErrorMatches, `application "application-name" not found`)
}

func (s *applicationSuite) deployApplicationForUpdateTests(c *gc.C) {
	curl, _ := s.UploadCharm(c, "cs:quantal/dummy-1", "dummy")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	results, err := s.applicationAPI.Deploy(params.ApplicationsDeploy{
		Applications: []params.ApplicationDeploy{{
			CharmURL:        curl.String(),
			CharmOrigin:     createCharmOriginFromURL(c, curl),
			ApplicationName: "application",
			NumUnits:        1,
		}}})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.Results, gc.HasLen, 1)
	c.Assert(results.Results[0].Error, gc.IsNil)
}

func (s *applicationSuite) checkClientApplicationUpdateSetCharm(c *gc.C, forceCharmURL bool) {
	s.deployApplicationForUpdateTests(c)
	curl, _ := s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	// Update the charm for the application.
	minUnits := 3
	args := params.ApplicationUpdate{
		ApplicationName: "application",
		MinUnits:        &minUnits,
		ForceCharmURL:   forceCharmURL,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err = api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the charm has been updated and and the force flag correctly set.
	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(app.MinUnits(), gc.Equals, minUnits)
}

func (s *applicationSuite) TestApplicationUpdateSetCharm(c *gc.C) {
	s.checkClientApplicationUpdateSetCharm(c, false)
}

func (s *applicationSuite) TestBlockDestroyApplicationUpdate(c *gc.C) {
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationUpdate")
	s.checkClientApplicationUpdateSetCharm(c, false)
}

func (s *applicationSuite) TestBlockRemoveApplicationUpdate(c *gc.C) {
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationUpdate")
	s.checkClientApplicationUpdateSetCharm(c, false)
}

func (s *applicationSuite) setupApplicationUpdate(c *gc.C) string {
	s.deployApplicationForUpdateTests(c)
	curl, _ := s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)
	return curl.String()
}

func (s *applicationSuite) TestBlockChangeApplicationUpdate(c *gc.C) {
	curl := s.setupApplicationUpdate(c)
	s.BlockAllChanges(c, "TestBlockChangeApplicationUpdate")
	// Update the charm for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "application",
		CharmURL:        curl,
		ForceCharmURL:   false,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	s.AssertBlocked(c, err, "TestBlockChangeApplicationUpdate")
}

func (s *applicationSuite) TestApplicationUpdateSetMinUnits(c *gc.C) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))

	// Set minimum units for the application.
	minUnits := 2
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		MinUnits:        &minUnits,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the minimum number of units has been set.
	c.Assert(app.Refresh(), gc.IsNil)
	c.Assert(app.MinUnits(), gc.Equals, minUnits)
}

func (s *applicationSuite) TestApplicationUpdateSetMinUnitsWithLXDProfile(c *gc.C) {
	app := s.AddTestingApplication(c, "lxd-profile", s.AddTestingCharm(c, "lxd-profile"))

	// Set minimum units for the application.
	minUnits := 2
	args := params.ApplicationUpdate{
		ApplicationName: "lxd-profile",
		MinUnits:        &minUnits,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the minimum number of units has been set.
	c.Assert(app.Refresh(), gc.IsNil)
	c.Assert(app.MinUnits(), gc.Equals, minUnits)
}

func (s *applicationSuite) TestApplicationUpdateDoesNotSetMinUnitsWithLXDProfile(c *gc.C) {
	series := "quantal"
	repo := testcharms.RepoForSeries(series)
	ch := repo.CharmDir("lxd-profile-fail")
	ident := fmt.Sprintf("%s-%d", ch.Meta().Name, ch.Revision())
	curl := charm.MustParseURL(fmt.Sprintf("local:%s/%s", series, ident))
	_, err := jujutesting.PutCharm(s.State, curl, ch)
	c.Assert(err, gc.ErrorMatches, `invalid lxd-profile.yaml: contains device type "unix-disk"`)
}

func (s *applicationSuite) TestApplicationUpdateSetMinUnitsError(c *gc.C) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))

	// Set a negative minimum number of units for the application.
	minUnits := -1
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		MinUnits:        &minUnits,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, gc.ErrorMatches,
		`cannot set minimum units for application "dummy": cannot set a negative minimum number of units`)

	// Ensure the minimum number of units has not been set.
	c.Assert(app.Refresh(), gc.IsNil)
	c.Assert(app.MinUnits(), gc.Equals, 0)
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsStringsExplicitMaster(c *gc.C) {
	s.testApplicationUpdateSetSettingsStrings(c, model.GenerationMaster)
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsStringsEmptyBranchUsesMaster(c *gc.C) {
	s.testApplicationUpdateSetSettingsStrings(c, "")
}

func (s *applicationSuite) testApplicationUpdateSetSettingsStrings(c *gc.C, branchName string) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsStrings: map[string]string{"title": "s-title", "username": "s-user"},
		Generation:      branchName,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "s-title", "username": "s-user"}
	obtained, err := app.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsStringsBranch(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	const newBranch = "newBranch"
	c.Assert(s.State.AddBranch(newBranch, "user"), jc.ErrorIsNil)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsStrings: map[string]string{"title": "s-title", "username": "s-user"},
		Generation:      newBranch,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "s-title", "username": "s-user"}
	obtained, err := app.CharmConfig(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))

	// Check that the application is recorded against the generation.
	gen, err := s.State.Branch(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(gen.AssignedUnits(), jc.DeepEquals, map[string][]string{"dummy": {}})
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsYAMLExplicitMaster(c *gc.C) {
	s.testApplicationUpdateSetSettingsYAML(c, model.GenerationMaster)
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsYAMLEmptyBranchUsesMaster(c *gc.C) {
	s.testApplicationUpdateSetSettingsYAML(c, "")
}

func (s *applicationSuite) testApplicationUpdateSetSettingsYAML(c *gc.C, branchName string) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsYAML:    "dummy:\n  title: y-title\n  username: y-user",
		Generation:      branchName,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "y-title", "username": "y-user"}
	obtained, err := app.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))
}

func (s *applicationSuite) TestApplicationUpdateSetSettingsYAMLBranch(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	const newBranch = "newBranch"
	c.Assert(s.State.AddBranch(newBranch, "user"), jc.ErrorIsNil)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsYAML:    "dummy:\n  title: y-title\n  username: y-user",
		Generation:      newBranch,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "y-title", "username": "y-user"}
	obtained, err := app.CharmConfig(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))

	// Check that the application is recorded against the generation.
	gen, err := s.State.Branch(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(gen.AssignedUnits(), jc.DeepEquals, map[string][]string{"dummy": {}})
}

func (s *applicationSuite) TestClientApplicationUpdateSetSettingsGetYAML(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsYAML:    "charm: dummy\napplication: dummy\nsettings:\n  title:\n    value: y-title\n    type: string\n  username:\n    value: y-user\n  ignore:\n    blah: true",
		Generation:      model.GenerationMaster,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "y-title", "username": "y-user"}
	obtained, err := app.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))
}

func (s *applicationSuite) TestApplicationUpdateCombinedStringAndYAMLSettings(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy", ch)

	const newBranch = "newBranch"
	c.Assert(s.State.AddBranch(newBranch, "user"), jc.ErrorIsNil)

	// Update settings for the application.
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		SettingsStrings: map[string]string{
			"username": "s-user",
		},
		SettingsYAML: "dummy:\n  title: s-title",
		Generation:   newBranch,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the settings have been correctly updated.
	expected := charm.Settings{"title": "s-title", "username": "s-user"}
	obtained, err := app.CharmConfig(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, s.combinedSettings(ch, expected))

	// Check that the application is recorded against the generation.
	gen, err := s.State.Branch(newBranch)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(gen.AssignedUnits(), jc.DeepEquals, map[string][]string{"dummy": {}})
}

func (s *applicationSuite) TestApplicationUpdateSetConstraints(c *gc.C) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))

	// Update constraints for the application.
	cons, err := constraints.Parse("mem=4096", "cores=2")
	c.Assert(err, jc.ErrorIsNil)
	args := params.ApplicationUpdate{
		ApplicationName: "dummy",
		Constraints:     &cons,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err = api.Update(args)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the constraints have been correctly updated.
	obtained, err := app.Constraints()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, cons)
}

func (s *applicationSuite) TestApplicationUpdateAllParams(c *gc.C) {
	s.deployApplicationForUpdateTests(c)
	curl, _ := s.UploadCharm(c, "cs:quantal/wordpress-3", "wordpress")
	err := application.AddCharmWithAuthorization(application.NewStateShim(s.State), params.AddCharmWithAuthorization{
		URL: curl.String(),
	}, s.openRepo)
	c.Assert(err, jc.ErrorIsNil)

	app, err := s.State.Application("application")
	c.Assert(err, jc.ErrorIsNil)

	expectCurl, expectForce, _ := app.Charm()
	expectMinUnits := app.MinUnits()

	// Update all the application attributes.
	minUnits := 3
	cons, err := constraints.Parse("mem=4096", "cores=2")
	c.Assert(err, jc.ErrorIsNil)
	args := params.ApplicationUpdate{
		ApplicationName: "application",
		CharmURL:        curl.String(),
		ForceCharmURL:   true,
		MinUnits:        &minUnits,
		SettingsStrings: map[string]string{"blog-title": "string-title"},
		SettingsYAML:    "application:\n  blog-title: yaml-title\n",
		Constraints:     &cons,
		Generation:      model.GenerationMaster,
	}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err = api.Update(args)
	c.Assert(err, jc.Satisfies, errors.IsNotSupported)

	// Ensure the application not been.
	err = app.Refresh()
	c.Assert(err, jc.ErrorIsNil)

	// Check the charm.
	ch, force, err := app.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ch.String(), gc.Equals, expectCurl.String())
	c.Assert(force, gc.Equals, expectForce)

	// Check the minimum number of units.
	c.Assert(app.MinUnits(), gc.Equals, expectMinUnits)
}

func (s *applicationSuite) TestApplicationUpdateNoParams(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))

	// Calling Update with no parameters set is a no-op.
	args := params.ApplicationUpdate{ApplicationName: "wordpress"}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *applicationSuite) TestApplicationUpdateNoApplication(c *gc.C) {
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(params.ApplicationUpdate{})
	c.Assert(err, gc.ErrorMatches, `"" is not a valid application name`)
}

func (s *applicationSuite) TestApplicationUpdateInvalidApplication(c *gc.C) {
	args := params.ApplicationUpdate{ApplicationName: "no-such-application"}
	api := &application.APIv12{&application.APIv13{&application.APIv14{s.applicationAPI}}}
	err := api.Update(args)
	c.Assert(err, gc.ErrorMatches, `application "no-such-application" not found`)
}

var (
	validSetTestValue = "a value with spaces\nand newline\nand UTF-8 characters: \U0001F604 / \U0001F44D"
)

func (s *applicationSuite) TestApplicationSet(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	dummy := s.AddTestingApplication(c, "dummy", ch)

	err := s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy",
		Options: map[string]string{
			"title":    "foobar",
			"username": validSetTestValue,
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err := dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title":    "foobar",
		"username": validSetTestValue,
	}))

	err = s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy", Options: map[string]string{
			"title":    "barfoo",
			"username": "",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err = dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title":    "barfoo",
		"username": "",
	}))
}

func (s *applicationSuite) assertApplicationSetBlocked(c *gc.C, dummy *state.Application, msg string) {
	err := s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy",
		Options: map[string]string{
			"title":    "foobar",
			"username": validSetTestValue}})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) assertApplicationSet(c *gc.C, dummy *state.Application) {
	err := s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy",
		Options: map[string]string{
			"title":    "foobar",
			"username": validSetTestValue,
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err := dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := dummy.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title":    "foobar",
		"username": validSetTestValue,
	}))
}

func (s *applicationSuite) TestBlockDestroyApplicationSet(c *gc.C) {
	dummy := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationSet")
	s.assertApplicationSet(c, dummy)
}

func (s *applicationSuite) TestBlockRemoveApplicationSet(c *gc.C) {
	dummy := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationSet")
	s.assertApplicationSet(c, dummy)
}

func (s *applicationSuite) TestBlockChangesApplicationSet(c *gc.C) {
	dummy := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockAllChanges(c, "TestBlockChangesApplicationSet")
	s.assertApplicationSetBlocked(c, dummy, "TestBlockChangesApplicationSet")
}

func (s *applicationSuite) TestServerUnset(c *gc.C) {
	ch := s.AddTestingCharm(c, "dummy")
	dummy := s.AddTestingApplication(c, "dummy", ch)

	err := s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy",
		Options: map[string]string{
			"title":    "foobar",
			"username": "user name",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err := dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title":    "foobar",
		"username": "user name",
	}))

	err = s.applicationAPI.Unset(params.ApplicationUnset{
		ApplicationName: "dummy",
		Options:         []string{"username"},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err = dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title": "foobar",
	}))
}

func (s *applicationSuite) setupServerUnsetBlocked(c *gc.C) *state.Application {
	dummy := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))

	err := s.applicationAPI.Set(params.ApplicationSet{
		ApplicationName: "dummy",
		Options: map[string]string{
			"title":    "foobar",
			"username": "user name",
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err := dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := dummy.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title":    "foobar",
		"username": "user name",
	}))
	return dummy
}

func (s *applicationSuite) assertServerUnset(c *gc.C, dummy *state.Application) {
	err := s.applicationAPI.Unset(params.ApplicationUnset{
		ApplicationName: "dummy",
		Options:         []string{"username"},
	})
	c.Assert(err, jc.ErrorIsNil)
	settings, err := dummy.CharmConfig(model.GenerationMaster)
	c.Assert(err, jc.ErrorIsNil)
	ch, _, err := dummy.Charm()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(settings, gc.DeepEquals, s.combinedSettings(ch, charm.Settings{
		"title": "foobar",
	}))
}

func (s *applicationSuite) assertServerUnsetBlocked(c *gc.C, dummy *state.Application, msg string) {
	err := s.applicationAPI.Unset(params.ApplicationUnset{
		ApplicationName: "dummy",
		Options:         []string{"username"},
	})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) TestBlockDestroyServerUnset(c *gc.C) {
	dummy := s.setupServerUnsetBlocked(c)
	s.BlockDestroyModel(c, "TestBlockDestroyServerUnset")
	s.assertServerUnset(c, dummy)
}

func (s *applicationSuite) TestBlockRemoveServerUnset(c *gc.C) {
	dummy := s.setupServerUnsetBlocked(c)
	s.BlockRemoveObject(c, "TestBlockRemoveServerUnset")
	s.assertServerUnset(c, dummy)
}

func (s *applicationSuite) TestBlockChangesServerUnset(c *gc.C) {
	dummy := s.setupServerUnsetBlocked(c)
	s.BlockAllChanges(c, "TestBlockChangesServerUnset")
	s.assertServerUnsetBlocked(c, dummy, "TestBlockChangesServerUnset")
}

var clientAddApplicationUnitsTests = []struct {
	about       string
	application string // if not set, defaults to 'dummy'
	numUnits    int
	expected    []string
	to          string
	err         string
}{
	{
		about:    "returns unit names",
		numUnits: 3,
		expected: []string{"dummy/0", "dummy/1", "dummy/2"},
	},
	{
		about: "fails trying to add zero units",
		err:   "must add at least one unit",
	},
	{
		// Note: chained-state, we add 1 unit here, but the 3 units
		// from the first condition still exist
		about:    "force the unit onto bootstrap machine",
		numUnits: 1,
		expected: []string{"dummy/3"},
		to:       "0",
	},
	{
		about:       "unknown application name",
		application: "unknown-application",
		numUnits:    1,
		err:         `application "unknown-application" not found`,
	},
}

func (s *applicationSuite) TestClientAddApplicationUnits(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	for i, t := range clientAddApplicationUnitsTests {
		c.Logf("test %d. %s", i, t.about)
		applicationName := t.application
		if applicationName == "" {
			applicationName = "dummy"
		}
		args := params.AddApplicationUnits{
			ApplicationName: applicationName,
			NumUnits:        t.numUnits,
		}
		if t.to != "" {
			args.Placement = []*instance.Placement{instance.MustParsePlacement(t.to)}
		}
		result, err := s.applicationAPI.AddUnits(args)
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(result.Units, gc.DeepEquals, t.expected)
	}
	// Test that we actually assigned the unit to machine 0
	forcedUnit, err := s.BackingState.Unit("dummy/3")
	c.Assert(err, jc.ErrorIsNil)
	assignedMachine, err := forcedUnit.AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(assignedMachine, gc.Equals, "0")
}

func (s *applicationSuite) TestAddApplicationUnitsToNewContainer(c *gc.C) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	machine, err := s.State.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)

	_, err = s.applicationAPI.AddUnits(params.AddApplicationUnits{
		ApplicationName: "dummy",
		NumUnits:        1,
		Placement:       []*instance.Placement{instance.MustParsePlacement("lxd:" + machine.Id())},
	})
	c.Assert(err, jc.ErrorIsNil)

	units, err := app.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	mid, err := units[0].AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(mid, gc.Equals, machine.Id()+"/lxd/0")
}

var addApplicationUnitTests = []struct {
	about       string
	application string // if not set, defaults to 'dummy'
	expected    []string
	machineIds  []string
	placement   []*instance.Placement
	err         string
}{
	{
		about:      "valid placement directives",
		expected:   []string{"dummy/0"},
		placement:  []*instance.Placement{{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "valid"}},
		machineIds: []string{"1"},
	}, {
		about:      "direct machine assignment placement directive",
		expected:   []string{"dummy/1", "dummy/2"},
		placement:  []*instance.Placement{{Scope: "#", Directive: "1"}, {Scope: "lxd", Directive: "1"}},
		machineIds: []string{"1", "1/lxd/0"},
	}, {
		about:     "invalid placement directive",
		err:       ".* invalid placement is invalid",
		expected:  []string{"dummy/3"},
		placement: []*instance.Placement{{Scope: "deadbeef-0bad-400d-8000-4b1d0d06f00d", Directive: "invalid"}},
	},
}

func (s *applicationSuite) TestAddApplicationUnits(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	// Add a machine for the units to be placed on.
	_, err := s.State.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	for i, t := range addApplicationUnitTests {
		c.Logf("test %d. %s", i, t.about)
		applicationName := t.application
		if applicationName == "" {
			applicationName = "dummy"
		}
		result, err := s.applicationAPI.AddUnits(params.AddApplicationUnits{
			ApplicationName: applicationName,
			NumUnits:        len(t.expected),
			Placement:       t.placement,
		})
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
			continue
		}
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(result.Units, gc.DeepEquals, t.expected)
		for i, unitName := range result.Units {
			u, err := s.BackingState.Unit(unitName)
			c.Assert(err, jc.ErrorIsNil)
			assignedMachine, err := u.AssignedMachineId()
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(assignedMachine, gc.Equals, t.machineIds[i])
		}
	}
}

func (s *applicationSuite) assertAddApplicationUnits(c *gc.C) {
	result, err := s.applicationAPI.AddUnits(params.AddApplicationUnits{
		ApplicationName: "dummy",
		NumUnits:        3,
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result.Units, gc.DeepEquals, []string{"dummy/0", "dummy/1", "dummy/2"})

	// Test that we actually assigned the unit to machine 0
	forcedUnit, err := s.BackingState.Unit("dummy/0")
	c.Assert(err, jc.ErrorIsNil)
	assignedMachine, err := forcedUnit.AssignedMachineId()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(assignedMachine, gc.Equals, "0")
}

func (s *applicationSuite) TestApplicationCharmRelations(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)

	_, err = s.applicationAPI.CharmRelations(params.ApplicationCharmRelations{ApplicationName: "blah"})
	c.Assert(err, gc.ErrorMatches, `application "blah" not found`)

	result, err := s.applicationAPI.CharmRelations(params.ApplicationCharmRelations{ApplicationName: "wordpress"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result.CharmRelations, gc.DeepEquals, []string{
		"cache", "db", "juju-info", "logging-dir", "monitoring-port", "url",
	})
}

func (s *applicationSuite) assertAddApplicationUnitsBlocked(c *gc.C, msg string) {
	_, err := s.applicationAPI.AddUnits(params.AddApplicationUnits{
		ApplicationName: "dummy",
		NumUnits:        3,
	})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) TestBlockDestroyAddApplicationUnits(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockDestroyModel(c, "TestBlockDestroyAddApplicationUnits")
	s.assertAddApplicationUnits(c)
}

func (s *applicationSuite) TestBlockRemoveAddApplicationUnits(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockRemoveObject(c, "TestBlockRemoveAddApplicationUnits")
	s.assertAddApplicationUnits(c)
}

func (s *applicationSuite) TestBlockChangeAddApplicationUnits(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	s.BlockAllChanges(c, "TestBlockChangeAddApplicationUnits")
	s.assertAddApplicationUnitsBlocked(c, "TestBlockChangeAddApplicationUnits")
}

func (s *applicationSuite) TestAddUnitToMachineNotFound(c *gc.C) {
	s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	_, err := s.applicationAPI.AddUnits(params.AddApplicationUnits{
		ApplicationName: "dummy",
		NumUnits:        3,
		Placement:       []*instance.Placement{instance.MustParsePlacement("42")},
	})
	c.Assert(err, gc.ErrorMatches, `acquiring machine to host unit "dummy/0": machine 42 not found`)
}

func (s *applicationSuite) TestApplicationExpose(c *gc.C) {
	charm := s.AddTestingCharm(c, "dummy")
	applicationNames := []string{"dummy-application", "exposed-application"}
	apps := make([]*state.Application, len(applicationNames))
	var err error
	for i, name := range applicationNames {
		apps[i] = s.AddTestingApplication(c, name, charm)
		c.Assert(apps[i].IsExposed(), jc.IsFalse)
	}
	err = apps[1].MergeExposeSettings(nil)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(apps[1].IsExposed(), jc.IsTrue)

	s.assertApplicationExpose(c)
}

func (s *applicationSuite) TestApplicationExposeEndpoints(c *gc.C) {
	charm := s.AddTestingCharm(c, "wordpress")
	app := s.AddTestingApplication(c, "wordpress", charm)
	c.Assert(app.IsExposed(), jc.IsFalse)

	err := s.applicationAPI.Expose(params.ApplicationExpose{
		ApplicationName: app.Name(),
		ExposedEndpoints: map[string]params.ExposedEndpoint{
			// Exposing an endpoint with no expose options implies
			// expose to 0.0.0.0/0 and ::/0.
			"monitoring-port": {},
		},
	})
	c.Assert(err, jc.ErrorIsNil)

	got, err := s.State.Application(app.Name())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(got.IsExposed(), gc.Equals, true)
	c.Assert(got.ExposedEndpoints(), gc.DeepEquals, map[string]state.ExposedEndpoint{
		"monitoring-port": {
			ExposeToCIDRs: []string{firewall.AllNetworksIPV4CIDR, firewall.AllNetworksIPV6CIDR},
		},
	})
}

func (s *applicationSuite) TestApplicationExposeEndpointsWithPre29Client(c *gc.C) {
	charm := s.AddTestingCharm(c, "wordpress")
	app := s.AddTestingApplication(c, "wordpress", charm)
	c.Assert(app.IsExposed(), jc.IsFalse)

	err := s.applicationAPI.Expose(params.ApplicationExpose{
		ApplicationName: app.Name(),
		// If no endpoint-specific expose params are provided, the call
		// will emulate the behavior of a pre 2.9 controller where all
		// ports are exposed to 0.0.0.0/0 and ::/0.
	})
	c.Assert(err, jc.ErrorIsNil)

	got, err := s.State.Application(app.Name())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(got.IsExposed(), gc.Equals, true)
	c.Assert(got.ExposedEndpoints(), gc.DeepEquals, map[string]state.ExposedEndpoint{
		"": {
			ExposeToCIDRs: []string{firewall.AllNetworksIPV4CIDR, firewall.AllNetworksIPV6CIDR},
		},
	})
}

func (s *applicationSuite) setupApplicationExpose(c *gc.C) {
	charm := s.AddTestingCharm(c, "dummy")
	applicationNames := []string{"dummy-application", "exposed-application"}
	apps := make([]*state.Application, len(applicationNames))
	var err error
	for i, name := range applicationNames {
		apps[i] = s.AddTestingApplication(c, name, charm)
		c.Assert(apps[i].IsExposed(), jc.IsFalse)
	}
	err = apps[1].MergeExposeSettings(nil)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(apps[1].IsExposed(), jc.IsTrue)
}

var applicationExposeTests = []struct {
	about                 string
	application           string
	exposedEndpointParams map[string]params.ExposedEndpoint
	//
	expExposed          bool
	expExposedEndpoints map[string]state.ExposedEndpoint
	expErr              string
}{
	{
		about:       "unknown application name",
		application: "unknown-application",
		expErr:      `application "unknown-application" not found`,
	},
	{
		about:       "expose all endpoints of an application ",
		application: "dummy-application",
		expExposed:  true,
		expExposedEndpoints: map[string]state.ExposedEndpoint{
			"": {
				ExposeToCIDRs: []string{"0.0.0.0/0", "::/0"},
			},
		},
	},
	{
		about:       "expose an already exposed application",
		application: "exposed-application",
		expExposed:  true,
		expExposedEndpoints: map[string]state.ExposedEndpoint{
			"": {
				ExposeToCIDRs: []string{"0.0.0.0/0", "::/0"},
			},
		},
	},
	{
		about:       "unknown endpoint name in expose parameters",
		application: "dummy-application",
		exposedEndpointParams: map[string]params.ExposedEndpoint{
			"bogus": {},
		},
		expErr: `endpoint "bogus" not found`,
	},
	{
		about:       "unknown space name in expose parameters",
		application: "dummy-application",
		exposedEndpointParams: map[string]params.ExposedEndpoint{
			"": {
				ExposeToSpaces: []string{"invaders"},
			},
		},
		expErr: `space "invaders" not found`,
	},
	{
		about:       "expose an application and provide expose parameters",
		application: "exposed-application",
		exposedEndpointParams: map[string]params.ExposedEndpoint{
			"": {
				ExposeToSpaces: []string{network.AlphaSpaceName},
				ExposeToCIDRs:  []string{"13.37.0.0/16"},
			},
		},
		expExposed: true,
		expExposedEndpoints: map[string]state.ExposedEndpoint{
			"": {
				ExposeToSpaceIDs: []string{network.AlphaSpaceId},
				ExposeToCIDRs:    []string{"13.37.0.0/16"},
			},
		},
	},
}

func (s *applicationSuite) assertApplicationExpose(c *gc.C) {
	for i, t := range applicationExposeTests {
		c.Logf("test %d. %s", i, t.about)
		err := s.applicationAPI.Expose(params.ApplicationExpose{
			ApplicationName:  t.application,
			ExposedEndpoints: t.exposedEndpointParams,
		})
		if t.expErr != "" {
			c.Assert(err, gc.ErrorMatches, t.expErr)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			app, err := s.State.Application(t.application)
			c.Assert(err, jc.ErrorIsNil)
			c.Assert(app.IsExposed(), gc.Equals, t.expExposed)
			c.Assert(app.ExposedEndpoints(), gc.DeepEquals, t.expExposedEndpoints)
		}
	}
}

func (s *applicationSuite) assertApplicationExposeBlocked(c *gc.C, msg string) {
	for i, t := range applicationExposeTests {
		c.Logf("test %d. %s", i, t.about)
		err := s.applicationAPI.Expose(params.ApplicationExpose{
			ApplicationName:  t.application,
			ExposedEndpoints: t.exposedEndpointParams,
		})
		s.AssertBlocked(c, err, msg)
	}
}

func (s *applicationSuite) TestBlockDestroyApplicationExpose(c *gc.C) {
	s.setupApplicationExpose(c)
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationExpose")
	s.assertApplicationExpose(c)
}

func (s *applicationSuite) TestBlockRemoveApplicationExpose(c *gc.C) {
	s.setupApplicationExpose(c)
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationExpose")
	s.assertApplicationExpose(c)
}

func (s *applicationSuite) TestBlockChangesApplicationExpose(c *gc.C) {
	s.setupApplicationExpose(c)
	s.BlockAllChanges(c, "TestBlockChangesApplicationExpose")
	s.assertApplicationExposeBlocked(c, "TestBlockChangesApplicationExpose")
}

var applicationUnexposeTests = []struct {
	about               string
	application         string
	err                 string
	initial             map[string]state.ExposedEndpoint
	unexposeEndpoints   []string
	expExposed          bool
	expExposedEndpoints map[string]state.ExposedEndpoint
}{
	{
		about:       "unknown application name",
		application: "unknown-application",
		err:         `application "unknown-application" not found`,
	},
	{
		about:       "unexpose a application without specifying any endpoints",
		application: "dummy-application",
		initial: map[string]state.ExposedEndpoint{
			"": {},
		},
		expExposed: false,
	},
	{
		about:       "unexpose specific application endpoint",
		application: "dummy-application",
		initial: map[string]state.ExposedEndpoint{
			"server":       {},
			"server-admin": {},
		},
		unexposeEndpoints: []string{"server"},
		// The server-admin (and hence the app) should remain exposed
		expExposed: true,
		expExposedEndpoints: map[string]state.ExposedEndpoint{
			"server-admin": {ExposeToCIDRs: []string{"0.0.0.0/0", "::/0"}},
		},
	},
	{
		about:       "unexpose all currently exposed application endpoints",
		application: "dummy-application",
		initial: map[string]state.ExposedEndpoint{
			"server":       {},
			"server-admin": {},
		},
		unexposeEndpoints: []string{"server", "server-admin"},
		// Application should now be unexposed as all its endpoints have
		// been unexposed.
		expExposed: false,
	},
	{
		about:       "unexpose an already unexposed application",
		application: "dummy-application",
		initial:     nil,
		expExposed:  false,
	},
}

func (s *applicationSuite) TestApplicationUnexpose(c *gc.C) {
	charm := s.AddTestingCharm(c, "mysql")
	for i, t := range applicationUnexposeTests {
		c.Logf("test %d. %s", i, t.about)
		app := s.AddTestingApplication(c, "dummy-application", charm)
		if len(t.initial) != 0 {
			err := app.MergeExposeSettings(t.initial)
			c.Assert(err, jc.ErrorIsNil)
		}
		c.Assert(app.IsExposed(), gc.Equals, len(t.initial) != 0)
		err := s.applicationAPI.Unexpose(params.ApplicationUnexpose{
			ApplicationName:  t.application,
			ExposedEndpoints: t.unexposeEndpoints,
		})
		if t.err == "" {
			c.Assert(err, jc.ErrorIsNil)
			app.Refresh()
			c.Assert(app.IsExposed(), gc.Equals, t.expExposed)
			c.Assert(app.ExposedEndpoints(), gc.DeepEquals, t.expExposedEndpoints)
		} else {
			c.Assert(err, gc.ErrorMatches, t.err)
		}
		err = app.Destroy()
		c.Assert(err, jc.ErrorIsNil)
	}
}

func (s *applicationSuite) setupApplicationUnexpose(c *gc.C) *state.Application {
	charm := s.AddTestingCharm(c, "dummy")
	app := s.AddTestingApplication(c, "dummy-application", charm)
	app.MergeExposeSettings(nil)
	c.Assert(app.IsExposed(), gc.Equals, true)
	return app
}

func (s *applicationSuite) assertApplicationUnexpose(c *gc.C, app *state.Application) {
	err := s.applicationAPI.Unexpose(params.ApplicationUnexpose{ApplicationName: "dummy-application"})
	c.Assert(err, jc.ErrorIsNil)
	app.Refresh()
	c.Assert(app.IsExposed(), gc.Equals, false)
	err = app.Destroy()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *applicationSuite) assertApplicationUnexposeBlocked(c *gc.C, app *state.Application, msg string) {
	err := s.applicationAPI.Unexpose(params.ApplicationUnexpose{ApplicationName: "dummy-application"})
	s.AssertBlocked(c, err, msg)
	err = app.Destroy()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *applicationSuite) TestBlockDestroyApplicationUnexpose(c *gc.C) {
	app := s.setupApplicationUnexpose(c)
	s.BlockDestroyModel(c, "TestBlockDestroyApplicationUnexpose")
	s.assertApplicationUnexpose(c, app)
}

func (s *applicationSuite) TestBlockRemoveApplicationUnexpose(c *gc.C) {
	app := s.setupApplicationUnexpose(c)
	s.BlockRemoveObject(c, "TestBlockRemoveApplicationUnexpose")
	s.assertApplicationUnexpose(c, app)
}

func (s *applicationSuite) TestBlockChangesApplicationUnexpose(c *gc.C) {
	app := s.setupApplicationUnexpose(c)
	s.BlockAllChanges(c, "TestBlockChangesApplicationUnexpose")
	s.assertApplicationUnexposeBlocked(c, app, "TestBlockChangesApplicationUnexpose")
}

var applicationDestroyTests = []struct {
	about       string
	application string
	err         string
}{
	{
		about:       "unknown application name",
		application: "unknown-application",
		err:         `application "unknown-application" not found`,
	},
	{
		about:       "destroy an application",
		application: "dummy-application",
	},
	{
		about:       "destroy an already destroyed application",
		application: "dummy-application",
		err:         `application "dummy-application" not found`,
	},
}

func (s *applicationSuite) TestApplicationDestroy(c *gc.C) {
	s.AddTestingApplication(c, "dummy-application", s.AddTestingCharm(c, "dummy"))
	_, err := s.State.AddRemoteApplication(state.AddRemoteApplicationParams{
		Name:        "remote-application",
		SourceModel: s.Model.ModelTag(),
		Token:       "t0",
	})
	c.Assert(err, jc.ErrorIsNil)

	for i, t := range applicationDestroyTests {
		c.Logf("test %d. %s", i, t.about)
		err := s.applicationAPI.Destroy(params.ApplicationDestroy{ApplicationName: t.application})
		if t.err != "" {
			c.Assert(err, gc.ErrorMatches, t.err)
		} else {
			c.Assert(err, jc.ErrorIsNil)
		}
	}

	// Now do Destroy on an application with units. Destroy will
	// cause the application to be not-Alive, but will not remove its
	// document.
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	applicationName := "wordpress"
	app, err := s.State.Application(applicationName)
	c.Assert(err, jc.ErrorIsNil)
	err = s.applicationAPI.Destroy(params.ApplicationDestroy{ApplicationName: applicationName})
	c.Assert(err, jc.ErrorIsNil)
	err = app.Refresh()
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func assertLife(c *gc.C, entity state.Living, life state.Life) {
	err := entity.Refresh()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(entity.Life(), gc.Equals, life)
}

func (s *applicationSuite) TestBlockApplicationDestroy(c *gc.C) {
	s.AddTestingApplication(c, "dummy-application", s.AddTestingCharm(c, "dummy"))

	// block remove-objects
	s.BlockRemoveObject(c, "TestBlockApplicationDestroy")
	err := s.applicationAPI.Destroy(params.ApplicationDestroy{ApplicationName: "dummy-application"})
	s.AssertBlocked(c, err, "TestBlockApplicationDestroy")
	// Tests may have invalid application names.
	app, err := s.State.Application("dummy-application")
	if err == nil {
		// For valid application names, check that application is alive :-)
		assertLife(c, app, state.Alive)
	}
}

func (s *applicationSuite) TestDestroyPrincipalUnits(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	units := make([]*state.Unit, 5)
	for i := range units {
		unit, err := wordpress.AddUnit(state.AddUnitParams{})
		c.Assert(err, jc.ErrorIsNil)
		unit.AssignToNewMachine()
		c.Assert(err, jc.ErrorIsNil)
		now := time.Now()
		sInfo := status.StatusInfo{
			Status:  status.Idle,
			Message: "",
			Since:   &now,
		}
		err = unit.SetAgentStatus(sInfo)
		c.Assert(err, jc.ErrorIsNil)
		units[i] = unit
	}
	s.assertDestroyPrincipalUnits(c, units)
}

func (s *applicationSuite) TestDestroySubordinateUnits(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	wordpress0, err := wordpress.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	ru, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)
	err = ru.EnterScope(nil)
	c.Assert(err, jc.ErrorIsNil)
	logging0, err := s.State.Unit("logging/0")
	c.Assert(err, jc.ErrorIsNil)

	// Try to destroy the subordinate alone; check it fails.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"logging/0"},
	})
	c.Assert(err, gc.ErrorMatches, `no units were destroyed: unit "logging/0" is a subordinate, .*`)
	assertLife(c, logging0, state.Alive)

	s.assertDestroySubordinateUnits(c, wordpress0, logging0)
}

func (s *applicationSuite) assertDestroyPrincipalUnits(c *gc.C, units []*state.Unit) {
	// Destroy 2 of them; check they become Dying.
	err := s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "wordpress/1"},
	})
	c.Assert(err, jc.ErrorIsNil)
	assertLife(c, units[0], state.Dying)
	assertLife(c, units[1], state.Dying)

	// Try to destroy an Alive one and a Dying one; check
	// it destroys the Alive one and ignores the Dying one.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/2", "wordpress/0"},
	})
	c.Assert(err, jc.ErrorIsNil)
	assertLife(c, units[2], state.Dying)

	// Try to destroy an Alive one along with a nonexistent one; check that
	// the valid instruction is followed but the invalid one is warned about.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"boojum/123", "wordpress/3"},
	})
	c.Assert(err, gc.ErrorMatches, `some units were not destroyed: unit "boojum/123" does not exist`)
	assertLife(c, units[3], state.Dying)

	// Make one Dead, and destroy an Alive one alongside it; check no errors.
	wp0, err := s.State.Unit("wordpress/0")
	c.Assert(err, jc.ErrorIsNil)
	err = wp0.EnsureDead()
	c.Assert(err, jc.ErrorIsNil)
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "wordpress/4"},
	})
	c.Assert(err, jc.ErrorIsNil)
	assertLife(c, units[0], state.Dead)
	assertLife(c, units[4], state.Dying)
}

func (s *applicationSuite) setupDestroyPrincipalUnits(c *gc.C) []*state.Unit {
	units := make([]*state.Unit, 5)
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	for i := range units {
		unit, err := wordpress.AddUnit(state.AddUnitParams{})
		c.Assert(err, jc.ErrorIsNil)
		err = unit.AssignToNewMachine()
		c.Assert(err, jc.ErrorIsNil)
		now := time.Now()
		sInfo := status.StatusInfo{
			Status:  status.Idle,
			Message: "",
			Since:   &now,
		}
		err = unit.SetAgentStatus(sInfo)
		c.Assert(err, jc.ErrorIsNil)
		units[i] = unit
	}
	return units
}

func (s *applicationSuite) assertBlockedErrorAndLiveliness(
	c *gc.C,
	err error,
	msg string,
	living1 state.Living,
	living2 state.Living,
	living3 state.Living,
	living4 state.Living,
) {
	s.AssertBlocked(c, err, msg)
	assertLife(c, living1, state.Alive)
	assertLife(c, living2, state.Alive)
	assertLife(c, living3, state.Alive)
	assertLife(c, living4, state.Alive)
}

func (s *applicationSuite) TestBlockChangesDestroyPrincipalUnits(c *gc.C) {
	units := s.setupDestroyPrincipalUnits(c)
	s.BlockAllChanges(c, "TestBlockChangesDestroyPrincipalUnits")
	err := s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "wordpress/1"},
	})
	s.assertBlockedErrorAndLiveliness(c, err, "TestBlockChangesDestroyPrincipalUnits", units[0], units[1], units[2], units[3])
}

func (s *applicationSuite) TestBlockRemoveDestroyPrincipalUnits(c *gc.C) {
	units := s.setupDestroyPrincipalUnits(c)
	s.BlockRemoveObject(c, "TestBlockRemoveDestroyPrincipalUnits")
	err := s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "wordpress/1"},
	})
	s.assertBlockedErrorAndLiveliness(c, err, "TestBlockRemoveDestroyPrincipalUnits", units[0], units[1], units[2], units[3])
}

func (s *applicationSuite) TestBlockDestroyDestroyPrincipalUnits(c *gc.C) {
	units := s.setupDestroyPrincipalUnits(c)
	s.BlockDestroyModel(c, "TestBlockDestroyDestroyPrincipalUnits")
	err := s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "wordpress/1"},
	})
	c.Assert(err, jc.ErrorIsNil)
	assertLife(c, units[0], state.Dying)
	assertLife(c, units[1], state.Dying)
}

func (s *applicationSuite) assertDestroySubordinateUnits(c *gc.C, wordpress0, logging0 *state.Unit) {
	// Try to destroy the principal and the subordinate together; check it warns
	// about the subordinate, but destroys the one it can. (The principal unit
	// agent will be responsible for destroying the subordinate.)
	err := s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "logging/0"},
	})
	c.Assert(err, gc.ErrorMatches, `some units were not destroyed: unit "logging/0" is a subordinate, .*`)
	assertLife(c, wordpress0, state.Dying)
	assertLife(c, logging0, state.Alive)
}

func (s *applicationSuite) TestBlockRemoveDestroySubordinateUnits(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	wordpress0, err := wordpress.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	ru, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)
	err = ru.EnterScope(nil)
	c.Assert(err, jc.ErrorIsNil)
	logging0, err := s.State.Unit("logging/0")
	c.Assert(err, jc.ErrorIsNil)

	s.BlockRemoveObject(c, "TestBlockRemoveDestroySubordinateUnits")
	// Try to destroy the subordinate alone; check it fails.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"logging/0"},
	})
	s.AssertBlocked(c, err, "TestBlockRemoveDestroySubordinateUnits")
	assertLife(c, rel, state.Alive)
	assertLife(c, wordpress0, state.Alive)
	assertLife(c, logging0, state.Alive)

	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "logging/0"},
	})
	s.AssertBlocked(c, err, "TestBlockRemoveDestroySubordinateUnits")
	assertLife(c, wordpress0, state.Alive)
	assertLife(c, logging0, state.Alive)
	assertLife(c, rel, state.Alive)
}

func (s *applicationSuite) TestBlockChangesDestroySubordinateUnits(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	wordpress0, err := wordpress.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	ru, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)
	err = ru.EnterScope(nil)
	c.Assert(err, jc.ErrorIsNil)
	logging0, err := s.State.Unit("logging/0")
	c.Assert(err, jc.ErrorIsNil)

	s.BlockAllChanges(c, "TestBlockChangesDestroySubordinateUnits")
	// Try to destroy the subordinate alone; check it fails.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"logging/0"},
	})
	s.AssertBlocked(c, err, "TestBlockChangesDestroySubordinateUnits")
	assertLife(c, rel, state.Alive)
	assertLife(c, wordpress0, state.Alive)
	assertLife(c, logging0, state.Alive)

	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"wordpress/0", "logging/0"},
	})
	s.AssertBlocked(c, err, "TestBlockChangesDestroySubordinateUnits")
	assertLife(c, wordpress0, state.Alive)
	assertLife(c, logging0, state.Alive)
	assertLife(c, rel, state.Alive)
}

func (s *applicationSuite) TestBlockDestroyDestroySubordinateUnits(c *gc.C) {
	wordpress := s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	wordpress0, err := wordpress.AddUnit(state.AddUnitParams{})
	c.Assert(err, jc.ErrorIsNil)
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rel, err := s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	ru, err := rel.Unit(wordpress0)
	c.Assert(err, jc.ErrorIsNil)
	err = ru.EnterScope(nil)
	c.Assert(err, jc.ErrorIsNil)
	logging0, err := s.State.Unit("logging/0")
	c.Assert(err, jc.ErrorIsNil)

	s.BlockDestroyModel(c, "TestBlockDestroyDestroySubordinateUnits")
	// Try to destroy the subordinate alone; check it fails.
	err = s.applicationAPI.DestroyUnits(params.DestroyApplicationUnits{
		UnitNames: []string{"logging/0"},
	})
	c.Assert(err, gc.ErrorMatches, `no units were destroyed: unit "logging/0" is a subordinate, .*`)
	assertLife(c, logging0, state.Alive)

	s.assertDestroySubordinateUnits(c, wordpress0, logging0)
}

func (s *applicationSuite) TestClientSetApplicationConstraints(c *gc.C) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))

	// Update constraints for the application.
	cons, err := constraints.Parse("mem=4096", "cores=2")
	c.Assert(err, jc.ErrorIsNil)
	err = s.applicationAPI.SetConstraints(params.SetConstraints{ApplicationName: "dummy", Constraints: cons})
	c.Assert(err, jc.ErrorIsNil)

	// Ensure the constraints have been correctly updated.
	obtained, err := app.Constraints()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, cons)
}

func (s *applicationSuite) setupSetApplicationConstraints(c *gc.C) (*state.Application, constraints.Value) {
	app := s.AddTestingApplication(c, "dummy", s.AddTestingCharm(c, "dummy"))
	// Update constraints for the application.
	cons, err := constraints.Parse("mem=4096", "cores=2")
	c.Assert(err, jc.ErrorIsNil)
	return app, cons
}

func (s *applicationSuite) assertSetApplicationConstraints(c *gc.C, application *state.Application, cons constraints.Value) {
	err := s.applicationAPI.SetConstraints(params.SetConstraints{ApplicationName: "dummy", Constraints: cons})
	c.Assert(err, jc.ErrorIsNil)
	// Ensure the constraints have been correctly updated.
	obtained, err := application.Constraints()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(obtained, gc.DeepEquals, cons)
}

func (s *applicationSuite) assertSetApplicationConstraintsBlocked(c *gc.C, msg string, application *state.Application, cons constraints.Value) {
	err := s.applicationAPI.SetConstraints(params.SetConstraints{ApplicationName: "dummy", Constraints: cons})
	s.AssertBlocked(c, err, msg)
}

func (s *applicationSuite) TestBlockDestroySetApplicationConstraints(c *gc.C) {
	app, cons := s.setupSetApplicationConstraints(c)
	s.BlockDestroyModel(c, "TestBlockDestroySetApplicationConstraints")
	s.assertSetApplicationConstraints(c, app, cons)
}

func (s *applicationSuite) TestBlockRemoveSetApplicationConstraints(c *gc.C) {
	app, cons := s.setupSetApplicationConstraints(c)
	s.BlockRemoveObject(c, "TestBlockRemoveSetApplicationConstraints")
	s.assertSetApplicationConstraints(c, app, cons)
}

func (s *applicationSuite) TestBlockChangesSetApplicationConstraints(c *gc.C) {
	app, cons := s.setupSetApplicationConstraints(c)
	s.BlockAllChanges(c, "TestBlockChangesSetApplicationConstraints")
	s.assertSetApplicationConstraintsBlocked(c, "TestBlockChangesSetApplicationConstraints", app, cons)
}

func (s *applicationSuite) TestClientGetApplicationConstraints(c *gc.C) {
	fooConstraints := constraints.MustParse("arch=amd64", "mem=4G")
	s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Name:        "foo",
		Constraints: fooConstraints,
	})
	barConstraints := constraints.MustParse("arch=amd64", "mem=128G", "cores=64")
	s.Factory.MakeApplication(c, &factory.ApplicationParams{
		Name:        "bar",
		Constraints: barConstraints,
	})

	results, err := s.applicationAPI.GetConstraints(params.Entities{
		Entities: []params.Entity{
			{Tag: "wat"}, {Tag: "machine-0"}, {Tag: "user-foo"},
			{Tag: "application-foo"}, {Tag: "application-bar"}, {Tag: "application-wat"},
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results, jc.DeepEquals, params.ApplicationGetConstraintsResults{
		Results: []params.ApplicationConstraint{
			{
				Error: &params.Error{Message: `"wat" is not a valid tag`},
			}, {
				Error: &params.Error{Message: `unexpected tag type, expected application, got machine`},
			}, {
				Error: &params.Error{Message: `unexpected tag type, expected application, got user`},
			}, {
				Constraints: fooConstraints,
			}, {
				Constraints: barConstraints,
			}, {
				Error: &params.Error{Message: `application "wat" not found`, Code: "not found"},
			},
		}})
}

func (s *applicationSuite) checkEndpoints(c *gc.C, mysqlAppName string, endpoints map[string]params.CharmRelation) {
	c.Assert(endpoints["wordpress"], gc.DeepEquals, params.CharmRelation{
		Name:      "db",
		Role:      "requirer",
		Interface: "mysql",
		Optional:  false,
		Limit:     1,
		Scope:     "global",
	})
	ep := params.CharmRelation{
		Name:      "server",
		Role:      "provider",
		Interface: "mysql",
		Scope:     "global",
	}
	// Remote applications don't use scope.
	if mysqlAppName == "hosted-mysql" {
		ep.Scope = ""
	}
	c.Assert(endpoints[mysqlAppName], gc.DeepEquals, ep)
}

func (s *applicationSuite) setupRelationScenario(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	eps, err := s.State.InferEndpoints("logging", "wordpress")
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *applicationSuite) assertAddRelation(c *gc.C, endpoints, viaCIDRs []string) {
	s.setupRelationScenario(c)

	res, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints, ViaCIDRs: viaCIDRs})
	c.Assert(err, jc.ErrorIsNil)
	// Show that the relation was added.
	wpApp, err := s.State.Application("wordpress")
	c.Assert(err, jc.ErrorIsNil)
	rels, err := wpApp.Relations()
	c.Assert(err, jc.ErrorIsNil)
	// There are 2 relations - the logging-wordpress one set up in the
	// scenario and the one created in this test.
	c.Assert(len(rels), gc.Equals, 2)

	// We may be related to a local application or a remote offer
	// or an application in another model.
	var mySqlApplication state.ApplicationEntity
	mySqlApplication, err = s.State.RemoteApplication("hosted-mysql")
	if errors.IsNotFound(err) {
		mySqlApplication, err = s.State.RemoteApplication("othermysql")
		if errors.IsNotFound(err) {
			mySqlApplication, err = s.State.Application("mysql")
			c.Assert(err, jc.ErrorIsNil)
			s.checkEndpoints(c, "mysql", res.Endpoints)
		} else {
			c.Assert(err, jc.ErrorIsNil)
			s.checkEndpoints(c, "othermysql", res.Endpoints)
		}
	} else {
		c.Assert(err, jc.ErrorIsNil)
		s.checkEndpoints(c, "hosted-mysql", res.Endpoints)
	}
	c.Assert(err, jc.ErrorIsNil)
	rels, err = mySqlApplication.Relations()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(len(rels), gc.Equals, 1)
}

func (s *applicationSuite) TestSuccessfullyAddRelation(c *gc.C) {
	endpoints := []string{"wordpress", "mysql"}
	s.assertAddRelation(c, endpoints, nil)
}

func (s *applicationSuite) TestBlockDestroyAddRelation(c *gc.C) {
	s.BlockDestroyModel(c, "TestBlockDestroyAddRelation")
	s.assertAddRelation(c, []string{"wordpress", "mysql"}, nil)
}
func (s *applicationSuite) TestBlockRemoveAddRelation(c *gc.C) {
	s.BlockRemoveObject(c, "TestBlockRemoveAddRelation")
	s.assertAddRelation(c, []string{"wordpress", "mysql"}, nil)
}

func (s *applicationSuite) TestBlockChangesAddRelation(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	s.BlockAllChanges(c, "TestBlockChangesAddRelation")
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: []string{"wordpress", "mysql"}})
	s.AssertBlocked(c, err, "TestBlockChangesAddRelation")
}

func (s *applicationSuite) TestSuccessfullyAddRelationSwapped(c *gc.C) {
	// Show that the order of the applications listed in the AddRelation call
	// does not matter.  This is a repeat of the previous test with the application
	// names swapped.
	endpoints := []string{"mysql", "wordpress"}
	s.assertAddRelation(c, endpoints, nil)
}

func (s *applicationSuite) TestCallWithOnlyOneEndpoint(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	endpoints := []string{"wordpress"}
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, "no relations found")
}

func (s *applicationSuite) TestCallWithOneEndpointTooMany(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	s.AddTestingApplication(c, "logging", s.AddTestingCharm(c, "logging"))
	endpoints := []string{"wordpress", "mysql", "logging"}
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, "cannot relate 3 endpoints")
}

func (s *applicationSuite) TestAddAlreadyAddedRelation(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	// Add a relation between wordpress and mysql.
	endpoints := []string{"wordpress", "mysql"}
	eps, err := s.State.InferEndpoints(endpoints...)
	c.Assert(err, jc.ErrorIsNil)
	_, err = s.State.AddRelation(eps...)
	c.Assert(err, jc.ErrorIsNil)
	// And try to add it again.
	_, err = s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, `cannot add relation "wordpress:db mysql:server": relation wordpress:db mysql:server`)
}

func (s *applicationSuite) setupRemoteApplication(c *gc.C) {
	results, err := s.applicationAPI.Consume(params.ConsumeApplicationArgs{
		Args: []params.ConsumeApplicationArg{
			{ApplicationOfferDetails: params.ApplicationOfferDetails{
				SourceModelTag:         testing.ModelTag.String(),
				OfferName:              "hosted-mysql",
				OfferUUID:              "hosted-mysql-uuid",
				ApplicationDescription: "A pretty popular database",
				Endpoints: []params.RemoteEndpoint{
					{Name: "server", Interface: "mysql", Role: "provider"},
				},
			}},
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.OneError(), gc.IsNil)
}

func (s *applicationSuite) TestAddRemoteRelation(c *gc.C) {
	s.setupRemoteApplication(c)
	// There's already a wordpress in the scenario this assertion sets up.
	s.assertAddRelation(c, []string{"wordpress", "hosted-mysql"}, nil)
}

func (s *applicationSuite) TestAddRemoteRelationWithRelName(c *gc.C) {
	s.setupRemoteApplication(c)
	s.assertAddRelation(c, []string{"wordpress", "hosted-mysql:server"}, nil)
}

func (s *applicationSuite) TestAddRemoteRelationVia(c *gc.C) {
	s.setupRemoteApplication(c)
	s.assertAddRelation(c, []string{"wordpress", "hosted-mysql:server"}, []string{"192.168.0.0/16"})

	rel, err := s.State.KeyRelation("wordpress:db hosted-mysql:server")
	c.Assert(err, jc.ErrorIsNil)
	w := rel.WatchRelationEgressNetworks()
	defer statetesting.AssertStop(c, w)
	wc := statetesting.NewStringsWatcherC(c, s.State, w)
	wc.AssertChange("192.168.0.0/16")
	wc.AssertNoChange()
}

func (s *applicationSuite) TestAddRemoteRelationOnlyOneEndpoint(c *gc.C) {
	s.setupRemoteApplication(c)
	endpoints := []string{"hosted-mysql"}
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, "no relations found")
}

func (s *applicationSuite) TestAlreadyAddedRemoteRelation(c *gc.C) {
	s.setupRemoteApplication(c)
	endpoints := []string{"wordpress", "hosted-mysql"}
	s.assertAddRelation(c, endpoints, nil)

	// And try to add it again.
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, regexp.QuoteMeta(`cannot add relation "wordpress:db hosted-mysql:server": relation wordpress:db hosted-mysql:server`))
}

func (s *applicationSuite) TestRemoteRelationInvalidEndpoint(c *gc.C) {
	s.setupRemoteApplication(c)
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))

	endpoints := []string{"wordpress", "hosted-mysql:nope"}
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, `saas application "hosted-mysql" has no "nope" relation`)
}

func (s *applicationSuite) TestRemoteRelationNoMatchingEndpoint(c *gc.C) {
	results, err := s.applicationAPI.Consume(params.ConsumeApplicationArgs{
		Args: []params.ConsumeApplicationArg{
			{ApplicationOfferDetails: params.ApplicationOfferDetails{
				SourceModelTag: testing.ModelTag.String(),
				OfferName:      "hosted-db2",
				OfferUUID:      "hosted-db2-uuid",
				Endpoints: []params.RemoteEndpoint{
					{Name: "database", Interface: "db2", Role: "provider"},
				},
			}},
		},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(results.OneError(), gc.IsNil)

	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	endpoints := []string{"wordpress", "hosted-db2"}
	_, err = s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, "no relations found")
}

func (s *applicationSuite) TestRemoteRelationApplicationNotFound(c *gc.C) {
	s.AddTestingApplication(c, "wordpress", s.AddTestingCharm(c, "wordpress"))
	endpoints := []string{"wordpress", "unknown"}
	_, err := s.applicationAPI.AddRelation(params.AddRelation{Endpoints: endpoints})
	c.Assert(err, gc.ErrorMatches, `application "unknown" not found`)
}
