// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package unit

import (
	"github.com/juju/cmd"
	"github.com/juju/utils/voyeur"

	"github.com/juju/juju/agent"
	jujudagent "github.com/juju/juju/cmd/jujud/agent"
	"github.com/juju/juju/worker/logsender"
)

type (
	ManifoldsConfig = manifoldsConfig
	K8sUnitAgent    = k8sUnitAgent
)

type K8sUnitAgentTest interface {
	cmd.Command
	DataDir() string
	ApplicationName() string
	SetAgentConf(cfg jujudagent.AgentConf)
	ChangeConfig(change agent.ConfigMutator) error
}

func NewForTest(ctx *cmd.Context, bufferedLogger *logsender.BufferedLogWriter, configChangedVal *voyeur.Value) (K8sUnitAgentTest, error) {
	return &k8sUnitAgent{
		ctx:              ctx,
		AgentConf:        jujudagent.NewAgentConf(""),
		bufferedLogger:   bufferedLogger,
		configChangedVal: configChangedVal,
	}, nil
}

func (c *k8sUnitAgent) SetAgentConf(cfg jujudagent.AgentConf) {
	c.AgentConf = cfg
}

func (c *k8sUnitAgent) DataDir() string {
	return c.AgentConf.DataDir()
}

func (c *k8sUnitAgent) ApplicationName() string {
	return c.applicationName
}
