// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package client

import (
	"context"

	"github.com/juju/juju/core/blockdevice"
	"github.com/juju/juju/core/network"
)

// BlockDeviceService instances can fetch block devices for a machine.
type BlockDeviceService interface {
	BlockDevices(ctx context.Context, machineId string) ([]blockdevice.BlockDevice, error)
}

// NetworkService is the interface that is used to interact with the
// network spaces/subnets.
type NetworkService interface {
	// GetAllSpaces returns all spaces for the model.
	GetAllSpaces(ctx context.Context) (network.SpaceInfos, error)
	// GetAllSubnets returns all the subnets for the model.
	GetAllSubnets(ctx context.Context) (network.SubnetInfos, error)
}
