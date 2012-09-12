package container

import (
	"fmt"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/log"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/upstart"
	"os"
	"path/filepath"
	"strings"
)

// Container contains running juju service units.
type Container interface {
	// Deploy deploys the unit into a new container.
	Deploy(unit *state.Unit, info *state.Info, tools *state.Tools) error

	// Destroy destroys the unit's container.
	Destroy(unit *state.Unit) error
}

// Simple is a Container that knows how deploy units within	
// the current machine.
type Simple struct {
	DataDir string
	// InitDir holds the directory where upstart scripts
	// will be deployed. If blank, the system default will
	// be used.
	InitDir string

	// TODO(rog) add LogDir?
}

func (c *Simple) service(unit *state.Unit) *upstart.Service {
	svc := upstart.NewService("juju-" + unit.AgentName())
	if c.InitDir != "" {
		svc.InitDir = c.InitDir
	}
	return svc
}

func (c *Simple) dirName(unit *state.Unit) string {
	return filepath.Join(c.DataDir, "agents", unit.AgentName())
}

// Deploy deploys a unit running the given tools unit into a new container.
// The unit will use the given info to connect to the state.
func (c *Simple) Deploy(unit *state.Unit, info *state.Info, tools *state.Tools) (err error) {
	if info.UseSSH {
		return fmt.Errorf("cannot deploy unit agent connecting with ssh")
	}
	toolsDir := environs.AgentToolsDir(c.DataDir, unit.AgentName())
	err = os.Symlink(tools.Binary.String(), toolsDir)
	if err != nil {
		return fmt.Errorf("cannot make agent tools symlink: %v", err)
	}
	defer func() {
		if err != nil {
			if err := os.Remove(toolsDir); err != nil {
				log.Printf("container: cannot remove tools symlink: %v", err)
			}
		}
	}()
	cmd := fmt.Sprintf(
		"%s unit --zookeeper-servers '%s' --log-file %s --unit-name %s",
		filepath.Join(toolsDir, "jujud"),
		strings.Join(info.Addrs, ","),
		filepath.Join("/var/log/juju", unit.AgentName()+".log"),
		unit.Name())

	conf := &upstart.Conf{
		Service: *c.service(unit),
		Desc:    "juju unit agent for " + unit.Name(),
		Cmd:     cmd,
	}
	dir := c.dirName(unit)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err := os.Remove(dir); err != nil {
				log.Printf("container: cannot remove agent dir: %v", err)
			}
		}
	}()
	return conf.Install()
}

// Destroy destroys the unit's container.
func (c *Simple) Destroy(unit *state.Unit) error {
	if err := c.service(unit).Remove(); err != nil {
		return err
	}
	return os.RemoveAll(c.dirName(unit))
}
