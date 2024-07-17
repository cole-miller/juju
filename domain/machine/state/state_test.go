// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"context"
	"database/sql"
	"sort"

	"github.com/canonical/sqlair"
	"github.com/juju/collections/transform"
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/core/blockdevice"
	"github.com/juju/juju/core/machine"
	"github.com/juju/juju/core/status"
	"github.com/juju/juju/domain/life"
	machineerrors "github.com/juju/juju/domain/machine/errors"
	schematesting "github.com/juju/juju/domain/schema/testing"
	loggertesting "github.com/juju/juju/internal/logger/testing"
	"github.com/juju/juju/internal/uuid"
)

type stateSuite struct {
	schematesting.ModelSuite

	state *State
}

var _ = gc.Suite(&stateSuite{})

// runQuery executes the provided SQL query string using the current state's database connection.
//
// It is a convenient function to setup test with a specific database state
func (s *stateSuite) runQuery(query string) error {
	db, err := s.state.DB()
	if err != nil {
		return err
	}
	stmt, err := sqlair.Prepare(query)
	if err != nil {
		return err
	}
	return db.Txn(context.Background(), func(ctx context.Context, tx *sqlair.TX) error {
		return tx.Query(ctx, stmt).Run()
	})
}

func (s *stateSuite) SetUpTest(c *gc.C) {
	s.ModelSuite.SetUpTest(c)

	s.state = NewState(s.TxnRunnerFactory(), loggertesting.WrapCheckLog(c))
}

// TestCreateMachine asserts the happy path of CreateMachine at the state layer.
func (s *stateSuite) TestCreateMachine(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	var (
		machineID string
	)
	err = s.TxnRunner().StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, "SELECT name FROM machine").Scan(&machineID)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	})
	c.Check(err, jc.ErrorIsNil)
	c.Assert(machineID, gc.Equals, "666")
}

// TestCreateMachineAlreadyExists asserts that a MachineAlreadyExists error is
// returned when the machine already exists.
func (s *stateSuite) TestCreateMachineAlreadyExists(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	err = s.state.CreateMachine(context.Background(), "666", "", "")
	c.Check(err, jc.ErrorIs, machineerrors.MachineAlreadyExists)
}

// TestCreateMachineWithParentSuccess asserts the happy path of
// CreateMachineWithParent at the state layer.
func (s *stateSuite) TestCreateMachineWithParentSuccess(c *gc.C) {
	// Create the parent first
	err := s.state.CreateMachine(context.Background(), "666", "3", "1")
	c.Assert(err, jc.ErrorIsNil)

	// Create the machine with the created parent
	err = s.state.CreateMachineWithParent(context.Background(), "667", "666", "4", "2")
	c.Check(err, jc.ErrorIsNil)
}

// TestCreateMachineWithParentNotFound asserts that a NotFound error is returned
// when the parent machine is not found.
func (s *stateSuite) TestCreateMachineWithParentNotFound(c *gc.C) {
	err := s.state.CreateMachineWithParent(context.Background(), "667", "666", "4", "2")
	c.Check(err, jc.ErrorIs, errors.NotFound)
}

// TestCreateMachineWithparentAlreadyExists asserts that a MachineAlreadyExists
// error is returned when the machine to be created already exists.
func (s *stateSuite) TestCreateMachineWithParentAlreadyExists(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	err = s.state.CreateMachineWithParent(context.Background(), "666", "357", "4", "2")
	c.Check(err, jc.ErrorIs, machineerrors.MachineAlreadyExists)
}

// TestDeleteMachine asserts the happy path of DeleteMachine at the state layer.
func (s *stateSuite) TestDeleteMachine(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	bd := blockdevice.BlockDevice{
		DeviceName:     "name-666",
		Label:          "label-666",
		UUID:           "device-666",
		HardwareId:     "hardware-666",
		WWN:            "wwn-666",
		BusAddress:     "bus-666",
		SizeMiB:        666,
		FilesystemType: "btrfs",
		InUse:          true,
		MountPoint:     "mount-666",
		SerialId:       "serial-666",
	}
	bdUUID := uuid.MustNewUUID().String()
	s.insertBlockDevice(c, bd, bdUUID, "666")

	err = s.state.DeleteMachine(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)

	var machineCount int
	err = s.TxnRunner().StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, "SELECT count(*) FROM machine WHERE name=?", "666").Scan(&machineCount)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	})
	c.Check(err, jc.ErrorIsNil)
	c.Assert(machineCount, gc.Equals, 0)
}

// TestDeleteMachineStatus asserts that DeleteMachine at the state layer removes
// any machine status and status data when deleting a machine.
func (s *stateSuite) TestDeleteMachineStatus(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	bd := blockdevice.BlockDevice{
		DeviceName:     "name-666",
		Label:          "label-666",
		UUID:           "device-666",
		HardwareId:     "hardware-666",
		WWN:            "wwn-666",
		BusAddress:     "bus-666",
		SizeMiB:        666,
		FilesystemType: "btrfs",
		InUse:          true,
		MountPoint:     "mount-666",
		SerialId:       "serial-666",
	}
	bdUUID := uuid.MustNewUUID().String()
	s.insertBlockDevice(c, bd, bdUUID, "666")

	s.state.SetMachineStatus(context.Background(), "666", status.StatusInfo{Status: status.Started, Message: "started", Data: map[string]interface{}{"key": "data"}})

	err = s.state.DeleteMachine(context.Background(), "666")
	c.Assert(err, jc.ErrorIsNil)

	var status int
	var statusData int
	err = s.TxnRunner().StdTxn(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, "SELECT count(*) FROM machine_status WHERE machine_uuid=?", "123").Scan(&status)
		if err != nil {
			return errors.Trace(err)
		}
		err = tx.QueryRowContext(ctx, "SELECT count(*) FROM machine_status_data WHERE machine_uuid=?", "123").Scan(&statusData)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	})
	c.Check(err, jc.ErrorIsNil)
	c.Assert(status, gc.Equals, 0)
	c.Assert(statusData, gc.Equals, 0)
}

func (s *stateSuite) insertBlockDevice(c *gc.C, bd blockdevice.BlockDevice, blockDeviceUUID, machineId string) {
	db := s.DB()

	inUse := 0
	if bd.InUse {
		inUse = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO block_device (uuid, name, label, device_uuid, hardware_id, wwn, bus_address, serial_id, mount_point, filesystem_type_id, Size_mib, in_use, machine_uuid)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 2, ?, ?, (SELECT uuid FROM machine WHERE name=?))
`, blockDeviceUUID, bd.DeviceName, bd.Label, bd.UUID, bd.HardwareId, bd.WWN, bd.BusAddress, bd.SerialId, bd.MountPoint, bd.SizeMiB, inUse, machineId)
	c.Assert(err, jc.ErrorIsNil)

	for _, link := range bd.DeviceLinks {
		_, err = db.ExecContext(context.Background(), `
INSERT INTO block_device_link_device (block_device_uuid, name)
VALUES (?, ?)
`, blockDeviceUUID, link)
		c.Assert(err, jc.ErrorIsNil)
	}
	c.Assert(err, jc.ErrorIsNil)
}

// TestGetMachineLifeSuccess asserts the happy path of GetMachineLife at the
// state layer.
func (s *stateSuite) TestGetMachineLifeSuccess(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	obtainedLife, err := s.state.GetMachineLife(context.Background(), "666")
	expectedLife := life.Alive
	c.Check(err, jc.ErrorIsNil)
	c.Assert(*obtainedLife, gc.Equals, expectedLife)
}

// TestGetMachineLifeNotFound asserts that a NotFound error is returned when the
// machine is not found.
func (s *stateSuite) TestGetMachineLifeNotFound(c *gc.C) {
	_, err := s.state.GetMachineLife(context.Background(), "666")
	c.Assert(err, jc.ErrorIs, machineerrors.NotFound)
}

func (s *stateSuite) TestListAllMachines(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "3", "1")
	c.Assert(err, jc.ErrorIsNil)

	err = s.state.CreateMachine(context.Background(), "667", "4", "2")
	c.Assert(err, jc.ErrorIsNil)

	machines, err := s.state.AllMachineNames(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	expectedMachines := []string{"666", "667"}
	ms := transform.Slice[machine.Name, string](machines, func(m machine.Name) string { return m.String() })

	sort.Strings(ms)
	sort.Strings(expectedMachines)
	c.Assert(ms, gc.DeepEquals, expectedMachines)
}

// TestGetMachineStatusSuccess asserts the happy path of GetMachineStatus at the
// state layer.
func (s *stateSuite) TestGetMachineStatusSuccess(c *gc.C) {
	db := s.DB()

	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	// Add a status value for this machine into the
	// machine_status table using the machineUUID and the status
	// value 2 for "running" (from instance_status_values table).
	_, err = db.ExecContext(context.Background(), "INSERT INTO machine_status VALUES('123', '1', 'started', '2024-07-12 12:00:00')")
	c.Assert(err, jc.ErrorIsNil)

	obtainedStatus, err := s.state.GetMachineStatus(context.Background(), "666")
	expectedStatus := status.StatusInfo{Status: status.Started, Message: "started"}
	c.Check(err, jc.ErrorIsNil)
	c.Assert(obtainedStatus.Status, gc.Equals, expectedStatus.Status)
	c.Assert(obtainedStatus.Message, gc.Equals, expectedStatus.Message)
}

// TestGetMachineStatusWithData asserts the happy path of GetMachineStatus at
// the state layer.
func (s *stateSuite) TestGetMachineStatusSuccessWithData(c *gc.C) {
	db := s.DB()

	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	// Add a status value for this machine into the
	// machine_status table using the machineUUID and the status
	// value 2 for "running" (from instance_status_values table).
	_, err = db.ExecContext(context.Background(), "INSERT INTO machine_status VALUES('123', '1', 'started', '2024-07-12 12:00:00')")
	c.Assert(err, jc.ErrorIsNil)

	// Add some status data for this machine into the machine_status_data table.
	_, err = db.ExecContext(context.Background(), "INSERT INTO machine_status_data VALUES('123', 'key', 'data')")
	c.Assert(err, jc.ErrorIsNil)

	obtainedStatus, err := s.state.GetMachineStatus(context.Background(), "666")
	expectedStatus := status.StatusInfo{Status: status.Started, Message: "started", Data: map[string]interface{}{"key": "data"}}
	c.Check(err, jc.ErrorIsNil)
	c.Assert(obtainedStatus.Status, gc.Equals, expectedStatus.Status)
	c.Assert(obtainedStatus.Message, gc.Equals, expectedStatus.Message)
	c.Assert(obtainedStatus.Data, gc.DeepEquals, expectedStatus.Data)
}

// TestGetMachineStatusNotFoundError asserts that a NotFound error is returned
// when the machine is not found.
func (s *stateSuite) TestGetMachineStatusNotFoundError(c *gc.C) {
	_, err := s.state.GetMachineStatus(context.Background(), "666")
	c.Assert(err, jc.ErrorIs, errors.NotFound)
}

// TestGetMachineStatusNotSetError asserts that a StatusNotSet error is returned
// when the status is not set.
func (s *stateSuite) TestGetMachineStatusNotSetError(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	_, err = s.state.GetMachineStatus(context.Background(), "666")
	c.Assert(err, jc.ErrorIs, machineerrors.StatusNotSet)
}

// TestSetMachineStatusSuccess asserts the happy path of SetMachineStatus at the
// state layer.
func (s *stateSuite) TestSetMachineStatusSuccess(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	expectedStatus := status.StatusInfo{Status: status.Started, Message: "started"}
	err = s.state.SetMachineStatus(context.Background(), "666", expectedStatus)
	c.Check(err, jc.ErrorIsNil)

	obtainedStatus, err := s.state.GetMachineStatus(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(obtainedStatus.Status, gc.Equals, expectedStatus.Status)
	c.Assert(obtainedStatus.Message, gc.Equals, expectedStatus.Message)
}

// TestSetMachineStatusSuccessWithData asserts the happy path of
// SetMachineStatus at the state layer.
func (s *stateSuite) TestSetMachineStatusSuccessWithData(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "123")
	c.Assert(err, jc.ErrorIsNil)

	expectedStatus := status.StatusInfo{Status: status.Started, Message: "started", Data: map[string]interface{}{"key": "data"}}
	err = s.state.SetMachineStatus(context.Background(), "666", expectedStatus)
	c.Check(err, jc.ErrorIsNil)

	obtainedStatus, err := s.state.GetMachineStatus(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(obtainedStatus.Status, gc.Equals, expectedStatus.Status)
	c.Assert(obtainedStatus.Message, gc.Equals, expectedStatus.Message)
	c.Assert(obtainedStatus.Data, gc.DeepEquals, expectedStatus.Data)
}

// TestSetMachineStatusNotFoundError asserts that a NotFound error is returned
// when the machine is not found.
func (s *stateSuite) TestSetMachineStatusNotFoundError(c *gc.C) {
	err := s.state.SetMachineStatus(context.Background(), "666", status.StatusInfo{})
	c.Assert(err, jc.ErrorIs, errors.NotFound)
}

// TestMachineStatusValues asserts the keys and values in the
// machine_status_values table, because we convert between core.status values
// and machine_status_values based on these associations. This test will catch
// any discrepancies between the two sets of values, and error if/when any of
// them ever change.
func (s *stateSuite) TestMachineStatusValues(c *gc.C) {
	db := s.DB()

	// Check that the status values in the machine_status_values table match
	// the instance status values in core status.
	rows, err := db.QueryContext(context.Background(), "SELECT id, status FROM machine_status_values")
	defer rows.Close()
	c.Assert(err, jc.ErrorIsNil)
	var statusValues []struct {
		ID   int
		Name string
	}
	for rows.Next() {
		var statusValue struct {
			ID   int
			Name string
		}
		err = rows.Scan(&statusValue.ID, &statusValue.Name)
		c.Assert(err, jc.ErrorIsNil)
		statusValues = append(statusValues, statusValue)
	}
	c.Check(statusValues, gc.HasLen, 5)
	c.Check(statusValues[0].ID, gc.Equals, 0)
	c.Check(statusValues[0].Name, gc.Equals, "error")
	c.Check(statusValues[1].ID, gc.Equals, 1)
	c.Check(statusValues[1].Name, gc.Equals, "started")
	c.Check(statusValues[2].ID, gc.Equals, 2)
	c.Check(statusValues[2].Name, gc.Equals, "pending")
	c.Check(statusValues[3].ID, gc.Equals, 3)
	c.Check(statusValues[3].Name, gc.Equals, "stopped")
	c.Check(statusValues[4].ID, gc.Equals, 4)
	c.Check(statusValues[4].Name, gc.Equals, "down")
}

// TestMachineStatusValuesConversion asserts the conversions to and from the
// core status values and the internal status values for machine stay intact.
func (s *stateSuite) TestMachineStatusValuesConversion(c *gc.C) {

	// Assert that toCoreMachineStatusValues returns the correct status value
	tests := []struct {
		statusValue int
		expected    status.Status
	}{
		{0, status.Error},
		{1, status.Started},
		{2, status.Pending},
		{3, status.Stopped},
		{4, status.Down},
	}

	for _, test := range tests {
		status := machineStatusWithData{Status: test.statusValue}
		c.Check(status.toCoreMachineStatusValue(), gc.Equals, test.expected)
	}

	// Assert that fromCoreMachineStatusValues returns the correct status value
	for _, test := range tests {
		c.Check(fromCoreMachineStatusValue(test.expected), gc.Equals, test.statusValue)
	}
}

// TestInstanceStatusValuesConversion asserts the conversions to and from the core status values and the internal status values for instances stay intact.
func (s *stateSuite) TestInstanceStatusValuesConversion(c *gc.C) {

	// Assert that toCoreInstanceStatusValues returns the correct status value
	tests := []struct {
		statusValue int
		expected    status.Status
	}{
		{0, status.Empty},
		{1, status.Allocating},
		{2, status.Running},
		{3, status.ProvisioningError},
	}

	for _, test := range tests {
		status := machineStatusWithData{Status: test.statusValue}
		c.Check(status.toCoreInstanceStatusValue(), gc.Equals, test.expected)
	}

	// Assert that fromCoreInstanceStatusValues returns the correct status value
	for _, test := range tests {
		c.Check(fromCoreInstanceStatusValue(test.expected), gc.Equals, test.statusValue)
	}
}

// TestSetMachineLifeSuccess asserts the happy path of SetMachineLife at the
// state layer.
func (s *stateSuite) TestSetMachineLifeSuccess(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	// Assert the life status is initially Alive
	obtainedLife, err := s.state.GetMachineLife(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(*obtainedLife, gc.Equals, life.Alive)

	// Set the machine's life to Dead
	err = s.state.SetMachineLife(context.Background(), "666", life.Dead)
	c.Check(err, jc.ErrorIsNil)

	// Assert we get the Dead as the machine's new life status.
	obtainedLife, err = s.state.GetMachineLife(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(*obtainedLife, gc.Equals, life.Dead)
}

// TestSetMachineLifeNotFoundError asserts that we get a NotFound if the
// provided machine doesn't exist.
func (s *stateSuite) TestSetMachineLifeNotFoundError(c *gc.C) {
	err := s.state.SetMachineLife(context.Background(), "666", life.Dead)
	c.Assert(err, jc.ErrorIs, machineerrors.NotFound)
}

// TestListAllMachinesEmpty asserts that AllMachineNames returns an empty list
// if there are no machines.
func (s *stateSuite) TestListAllMachinesEmpty(c *gc.C) {
	machines, err := s.state.AllMachineNames(context.Background())
	c.Check(err, jc.ErrorIsNil)
	c.Assert(machines, gc.HasLen, 0)
}

// TestListAllMachineNamesSuccess asserts the happy path of AllMachineNames at
// the state layer.
func (s *stateSuite) TestListAllMachineNamesSuccess(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "3", "1")
	c.Assert(err, jc.ErrorIsNil)

	err = s.state.CreateMachine(context.Background(), "667", "4", "2")
	c.Assert(err, jc.ErrorIsNil)

	machines, err := s.state.AllMachineNames(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	expectedMachines := []string{"666", "667"}
	ms := transform.Slice[machine.Name, string](machines, func(m machine.Name) string { return m.String() })

	sort.Strings(ms)
	sort.Strings(expectedMachines)
	c.Assert(ms, gc.DeepEquals, expectedMachines)
}

// TestIsControllerSuccess asserts the happy path of IsController at the state
// layer.
func (s *stateSuite) TestIsControllerSuccess(c *gc.C) {
	err := s.state.CreateMachine(context.Background(), "666", "", "")
	c.Assert(err, jc.ErrorIsNil)

	isController, err := s.state.IsMachineController(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(isController, gc.Equals, false)

	db := s.DB()

	updateIsController := `
UPDATE machine
SET is_controller = TRUE
WHERE name = $1;
`
	_, err = db.ExecContext(context.Background(), updateIsController, "666")
	c.Assert(err, jc.ErrorIsNil)
	isController, err = s.state.IsMachineController(context.Background(), "666")
	c.Check(err, jc.ErrorIsNil)
	c.Assert(isController, gc.Equals, true)
}

// TestIsControllerNotFound asserts that a NotFound error is returned when the
// machine is not found.
func (s *stateSuite) TestIsControllerNotFound(c *gc.C) {
	_, err := s.state.IsMachineController(context.Background(), "666")
	c.Assert(err, jc.ErrorIs, machineerrors.NotFound)
}
