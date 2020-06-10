// Copyright 2020 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package application

import (
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/gnuflag"
	"github.com/juju/names/v4"
	"github.com/juju/naturalsort"

	"github.com/juju/juju/api/application"
	"github.com/juju/juju/apiserver/params"
	jujucmd "github.com/juju/juju/cmd"
	"github.com/juju/juju/cmd/modelcmd"
)

const showUnitDoc = `
The command takes deployed unit names as an argument.

Optionally, relation data for only a specified endpoint
or related unit may be shown, or just the application data. 

Examples:
    $ juju show-unit mysql/0
    $ juju show-unit mysql/0 wordpress/1
    $ juju show-unit mysql/0 --app
    $ juju show-unit mysql/0 --endpoint db
    $ juju show-unit mysql/0 --related-unit wordpress/2
`

// NewShowUnitCommand returns a command that displays unit info.
func NewShowUnitCommand() cmd.Command {
	s := &showUnitCommand{}
	s.newAPIFunc = func() (UnitsInfoAPI, error) {
		return s.newUnitAPI()
	}
	return modelcmd.Wrap(s)
}

type showUnitCommand struct {
	modelcmd.ModelCommandBase

	out         cmd.Output
	units       []string
	endpoint    string
	relatedUnit string
	appOnly     bool

	newAPIFunc func() (UnitsInfoAPI, error)
}

// Info implements Command.Info.
func (c *showUnitCommand) Info() *cmd.Info {
	showCmd := &cmd.Info{
		Name:    "show-unit",
		Args:    "<unit name>",
		Purpose: "Displays information about a unit.",
		Doc:     showUnitDoc,
	}
	return jujucmd.Info(showCmd)
}

// Init implements Command.Init.
func (c *showUnitCommand) Init(args []string) error {
	if len(args) < 1 {
		return errors.Errorf("an unit name must be supplied")
	}
	c.units = args
	if c.relatedUnit != "" && !names.IsValidUnit(c.relatedUnit) {
		return errors.NotValidf("related unit name %v", c.relatedUnit)
	}
	var invalid []string
	for _, one := range c.units {
		if !names.IsValidUnit(one) {
			invalid = append(invalid, one)
		}
	}
	if len(invalid) == 0 {
		return nil
	}
	plural := "s"
	if len(invalid) == 1 {
		plural = ""
	}
	return errors.NotValidf(`unit name%v %v`, plural, strings.Join(invalid, `, `))
}

// SetFlags implements Command.SetFlags.
func (c *showUnitCommand) SetFlags(f *gnuflag.FlagSet) {
	c.ModelCommandBase.SetFlags(f)
	c.out.AddFlags(f, "yaml", cmd.DefaultFormatters.Formatters())
	f.StringVar(&c.endpoint, "endpoint", "", "Only show relation data for the specified endpoint")
	f.StringVar(&c.relatedUnit, "related-unit", "", "Only show relation data for the specified unit")
	f.BoolVar(&c.appOnly, "app", false, "Only show application relation data")
}

// UnitsInfoAPI defines the API methods that show-unit command uses.
type UnitsInfoAPI interface {
	Close() error
	BestAPIVersion() int
	UnitsInfo([]names.UnitTag) ([]params.UnitInfoResult, error)
}

func (c *showUnitCommand) newUnitAPI() (UnitsInfoAPI, error) {
	root, err := c.NewAPIRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return application.NewClient(root), nil
}

// Info implements Command.Run.
func (c *showUnitCommand) Run(ctx *cmd.Context) error {
	client, err := c.newAPIFunc()
	if err != nil {
		return err
	}
	defer client.Close()

	if v := client.BestAPIVersion(); v < 12 {
		// old client does not support showing applications.
		return errors.NotSupportedf("show unit on API server version %v", v)
	}

	tags, err := c.getUnitTags()
	if err != nil {
		return err
	}

	results, err := client.UnitsInfo(tags)
	if err != nil {
		return errors.Trace(err)
	}

	var errs params.ErrorResults
	var valid []params.UnitResult
	for _, result := range results {
		if result.Error != nil {
			errs.Results = append(errs.Results, params.ErrorResult{result.Error})
			continue
		}
		valid = append(valid, *result.Result)
	}
	if len(errs.Results) > 0 {
		return errs.Combine()
	}

	output, err := c.formatUnitInfos(valid)
	if err != nil {
		return err
	}
	return c.out.Write(ctx, output)
}

func (c *showUnitCommand) getUnitTags() ([]names.UnitTag, error) {
	tags := make([]names.UnitTag, len(c.units))
	for i, one := range c.units {
		if !names.IsValidUnit(one) {
			return nil, errors.Errorf("invalid unit name %v", one)
		}
		tags[i] = names.NewUnitTag(one)
	}
	return tags, nil
}

func (c *showUnitCommand) formatUnitInfos(all []params.UnitResult) (map[string]UnitInfo, error) {
	if len(all) == 0 {
		return nil, nil
	}
	output := make(map[string]UnitInfo)
	for _, one := range all {
		tag, info, err := c.createUnitInfo(one)
		if err != nil {
			return nil, errors.Trace(err)
		}
		output[tag.Id()] = info
	}
	return output, nil
}

type UnitRelationData struct {
	InScope  bool                   `yaml:"in-scope" json:"in-scope"`
	UnitData map[string]interface{} `yaml:"data" json:"data"`
}

type RelationData struct {
	Endpoint                string                      `yaml:"endpoint" json:"endpoint"`
	CrossModel              bool                        `yaml:"cross-model,omitempty" json:"cross-model,omitempty"`
	RelatedEndpoint         string                      `yaml:"related-endpoint" json:"related-endpoint"`
	ApplicationRelationData map[string]interface{}      `yaml:"application-data,omitempty" json:"application-data,omitempty"`
	MyData                  UnitRelationData            `yaml:"local-unit,omitempty" json:"local-unit,omitempty"`
	Data                    map[string]UnitRelationData `yaml:"related-units,omitempty" json:"related-units,omitempty"`
}

// ApplicationInfo defines the serialization behaviour of the application information.
type UnitInfo struct {
	WorkloadVersion string         `yaml:"workload-version,omitempty" json:"workload-version,omitempty"`
	Machine         string         `yaml:"machine,omitempty" json:"machine,omitempty"`
	OpenedPorts     []string       `yaml:"opened-ports" json:"opened-ports"`
	PublicAddress   string         `yaml:"public-address,omitempty" json:"public-address,omitempty"`
	Charm           string         `yaml:"charm" json:"charm"`
	Leader          bool           `yaml:"leader" json:"leader"`
	RelationData    []RelationData `yaml:"relation-info,omitempty" json:"relation-info,omitempty"`

	// The following are for CAAS models.
	ProviderId string `yaml:"provider-id,omitempty" json:"provider-id,omitempty"`
	Address    string `yaml:"address,omitempty" json:"address,omitempty"`
}

func (c *showUnitCommand) createUnitInfo(details params.UnitResult) (names.UnitTag, UnitInfo, error) {
	tag, err := names.ParseUnitTag(details.Tag)
	if err != nil {
		return names.UnitTag{}, UnitInfo{}, errors.Trace(err)
	}

	info := UnitInfo{
		WorkloadVersion: details.WorkloadVersion,
		Machine:         details.Machine,
		OpenedPorts:     details.OpenedPorts,
		PublicAddress:   details.PublicAddress,
		Charm:           details.Charm,
		Leader:          details.Leader,
		ProviderId:      details.ProviderId,
		Address:         details.Address,
	}
	for _, rdparams := range details.RelationData {
		if c.endpoint != "" && rdparams.Endpoint != c.endpoint {
			continue
		}
		rd := RelationData{
			Endpoint:                rdparams.Endpoint,
			RelatedEndpoint:         rdparams.RelatedEndpoint,
			CrossModel:              rdparams.CrossModel,
			ApplicationRelationData: make(map[string]interface{}),
			Data:                    make(map[string]UnitRelationData),
		}
		for k, v := range rdparams.ApplicationData {
			rd.ApplicationRelationData[k] = v
		}
		if c.appOnly {
			info.RelationData = append(info.RelationData, rd)
			continue
		}
		var unitNames []string
		for remoteUnit := range rdparams.UnitRelationData {
			if c.relatedUnit != "" && remoteUnit != c.relatedUnit {
				continue
			}
			if remoteUnit == tag.Id() {
				data := rdparams.UnitRelationData[remoteUnit]
				urd := UnitRelationData{
					InScope:  data.InScope,
					UnitData: make(map[string]interface{}),
				}
				for k, v := range data.UnitData {
					urd.UnitData[k] = v
				}
				rd.MyData = urd
				continue
			}
			unitNames = append(unitNames, remoteUnit)
		}
		naturalsort.Sort(unitNames)
		for _, remoteUnit := range unitNames {
			data := rdparams.UnitRelationData[remoteUnit]
			urd := UnitRelationData{
				InScope:  data.InScope,
				UnitData: make(map[string]interface{}),
			}
			for k, v := range data.UnitData {
				urd.UnitData[k] = v
			}
			rd.Data[remoteUnit] = urd
		}
		info.RelationData = append(info.RelationData, rd)
	}

	return tag, info, nil
}
