// Copyright 2023 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package errors

import (
	"github.com/juju/errors"
)

const (
	// NotFound describes an error that occurs when the machine being operated
	// on does not exist.
	NotFound = errors.ConstError("machine not found")

	// NotProvisioned describes an error that occurs when the machine being
	// operated on is not provisioned yet.
	NotProvisioned = errors.ConstError("machine not provisioned")

	// StatusNotSet describes an error that occurs when the status of a machine
	// or a cloud instance is not set yet.
	StatusNotSet = errors.ConstError("status not set")

	// InvalidStatus describes a status that is not valid
	InvalidStatus = errors.ConstError("invalid status")

	// GrandParentNotSupported describes an error that occurs when the operation
	// found a grandparent machine, as it is not currently supported.
	GrandParentNotSupported = errors.ConstError("grandparent machine are not supported currently")

	// MachineAlreadyExists describes an error that occurs when creating a
	// machine if a machine with the same name already exists.
	MachineAlreadyExists = errors.ConstError("machine already exists")

	// MachineHasNoParent describes an error that occurs when a machine has no
	// parent.
	MachineHasNoParent = errors.ConstError("machine has no parent")

	// GrandParentNotAllowed describes an error that occurs when a parent is
	// detected for a parent of a machine.
	GrandParentNotAllowed = errors.ConstError("grandparent not allowed")
)
