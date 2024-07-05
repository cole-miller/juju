// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/domain/machine/service (interfaces: State)
//
// Generated by this command:
//
//	mockgen -typed -package service -destination package_mock_test.go github.com/juju/juju/domain/machine/service State
//

// Package service is a generated GoMock package.
package service

import (
	context "context"
	reflect "reflect"

	instance "github.com/juju/juju/core/instance"
	machine "github.com/juju/juju/core/machine"
	life "github.com/juju/juju/domain/life"
	gomock "go.uber.org/mock/gomock"
)

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// AllMachineNames mocks base method.
func (m *MockState) AllMachineNames(arg0 context.Context) ([]machine.Name, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllMachineNames", arg0)
	ret0, _ := ret[0].([]machine.Name)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllMachineNames indicates an expected call of AllMachineNames.
func (mr *MockStateMockRecorder) AllMachineNames(arg0 any) *MockStateAllMachineNamesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllMachineNames", reflect.TypeOf((*MockState)(nil).AllMachineNames), arg0)
	return &MockStateAllMachineNamesCall{Call: call}
}

// MockStateAllMachineNamesCall wrap *gomock.Call
type MockStateAllMachineNamesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateAllMachineNamesCall) Return(arg0 []machine.Name, arg1 error) *MockStateAllMachineNamesCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateAllMachineNamesCall) Do(f func(context.Context) ([]machine.Name, error)) *MockStateAllMachineNamesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateAllMachineNamesCall) DoAndReturn(f func(context.Context) ([]machine.Name, error)) *MockStateAllMachineNamesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CancelMachineReboot mocks base method.
func (m *MockState) CancelMachineReboot(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CancelMachineReboot", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CancelMachineReboot indicates an expected call of CancelMachineReboot.
func (mr *MockStateMockRecorder) CancelMachineReboot(arg0, arg1 any) *MockStateCancelMachineRebootCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelMachineReboot", reflect.TypeOf((*MockState)(nil).CancelMachineReboot), arg0, arg1)
	return &MockStateCancelMachineRebootCall{Call: call}
}

// MockStateCancelMachineRebootCall wrap *gomock.Call
type MockStateCancelMachineRebootCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateCancelMachineRebootCall) Return(arg0 error) *MockStateCancelMachineRebootCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateCancelMachineRebootCall) Do(f func(context.Context, string) error) *MockStateCancelMachineRebootCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateCancelMachineRebootCall) DoAndReturn(f func(context.Context, string) error) *MockStateCancelMachineRebootCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CreateMachine mocks base method.
func (m *MockState) CreateMachine(arg0 context.Context, arg1 machine.Name, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMachine", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMachine indicates an expected call of CreateMachine.
func (mr *MockStateMockRecorder) CreateMachine(arg0, arg1, arg2, arg3 any) *MockStateCreateMachineCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMachine", reflect.TypeOf((*MockState)(nil).CreateMachine), arg0, arg1, arg2, arg3)
	return &MockStateCreateMachineCall{Call: call}
}

// MockStateCreateMachineCall wrap *gomock.Call
type MockStateCreateMachineCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateCreateMachineCall) Return(arg0 error) *MockStateCreateMachineCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateCreateMachineCall) Do(f func(context.Context, machine.Name, string, string) error) *MockStateCreateMachineCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateCreateMachineCall) DoAndReturn(f func(context.Context, machine.Name, string, string) error) *MockStateCreateMachineCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DeleteMachine mocks base method.
func (m *MockState) DeleteMachine(arg0 context.Context, arg1 machine.Name) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMachine", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMachine indicates an expected call of DeleteMachine.
func (mr *MockStateMockRecorder) DeleteMachine(arg0, arg1 any) *MockStateDeleteMachineCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMachine", reflect.TypeOf((*MockState)(nil).DeleteMachine), arg0, arg1)
	return &MockStateDeleteMachineCall{Call: call}
}

// MockStateDeleteMachineCall wrap *gomock.Call
type MockStateDeleteMachineCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateDeleteMachineCall) Return(arg0 error) *MockStateDeleteMachineCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateDeleteMachineCall) Do(f func(context.Context, machine.Name) error) *MockStateDeleteMachineCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateDeleteMachineCall) DoAndReturn(f func(context.Context, machine.Name) error) *MockStateDeleteMachineCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DeleteMachineCloudInstance mocks base method.
func (m *MockState) DeleteMachineCloudInstance(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMachineCloudInstance", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMachineCloudInstance indicates an expected call of DeleteMachineCloudInstance.
func (mr *MockStateMockRecorder) DeleteMachineCloudInstance(arg0, arg1 any) *MockStateDeleteMachineCloudInstanceCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMachineCloudInstance", reflect.TypeOf((*MockState)(nil).DeleteMachineCloudInstance), arg0, arg1)
	return &MockStateDeleteMachineCloudInstanceCall{Call: call}
}

// MockStateDeleteMachineCloudInstanceCall wrap *gomock.Call
type MockStateDeleteMachineCloudInstanceCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateDeleteMachineCloudInstanceCall) Return(arg0 error) *MockStateDeleteMachineCloudInstanceCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateDeleteMachineCloudInstanceCall) Do(f func(context.Context, string) error) *MockStateDeleteMachineCloudInstanceCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateDeleteMachineCloudInstanceCall) DoAndReturn(f func(context.Context, string) error) *MockStateDeleteMachineCloudInstanceCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetMachineLife mocks base method.
func (m *MockState) GetMachineLife(arg0 context.Context, arg1 machine.Name) (*life.Life, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMachineLife", arg0, arg1)
	ret0, _ := ret[0].(*life.Life)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMachineLife indicates an expected call of GetMachineLife.
func (mr *MockStateMockRecorder) GetMachineLife(arg0, arg1 any) *MockStateGetMachineLifeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMachineLife", reflect.TypeOf((*MockState)(nil).GetMachineLife), arg0, arg1)
	return &MockStateGetMachineLifeCall{Call: call}
}

// MockStateGetMachineLifeCall wrap *gomock.Call
type MockStateGetMachineLifeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateGetMachineLifeCall) Return(arg0 *life.Life, arg1 error) *MockStateGetMachineLifeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateGetMachineLifeCall) Do(f func(context.Context, machine.Name) (*life.Life, error)) *MockStateGetMachineLifeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateGetMachineLifeCall) DoAndReturn(f func(context.Context, machine.Name) (*life.Life, error)) *MockStateGetMachineLifeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// HardwareCharacteristics mocks base method.
func (m *MockState) HardwareCharacteristics(arg0 context.Context, arg1 string) (*instance.HardwareCharacteristics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HardwareCharacteristics", arg0, arg1)
	ret0, _ := ret[0].(*instance.HardwareCharacteristics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HardwareCharacteristics indicates an expected call of HardwareCharacteristics.
func (mr *MockStateMockRecorder) HardwareCharacteristics(arg0, arg1 any) *MockStateHardwareCharacteristicsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HardwareCharacteristics", reflect.TypeOf((*MockState)(nil).HardwareCharacteristics), arg0, arg1)
	return &MockStateHardwareCharacteristicsCall{Call: call}
}

// MockStateHardwareCharacteristicsCall wrap *gomock.Call
type MockStateHardwareCharacteristicsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateHardwareCharacteristicsCall) Return(arg0 *instance.HardwareCharacteristics, arg1 error) *MockStateHardwareCharacteristicsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateHardwareCharacteristicsCall) Do(f func(context.Context, string) (*instance.HardwareCharacteristics, error)) *MockStateHardwareCharacteristicsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateHardwareCharacteristicsCall) DoAndReturn(f func(context.Context, string) (*instance.HardwareCharacteristics, error)) *MockStateHardwareCharacteristicsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// InitialWatchInstanceStatement mocks base method.
func (m *MockState) InitialWatchInstanceStatement() (string, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitialWatchInstanceStatement")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// InitialWatchInstanceStatement indicates an expected call of InitialWatchInstanceStatement.
func (mr *MockStateMockRecorder) InitialWatchInstanceStatement() *MockStateInitialWatchInstanceStatementCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitialWatchInstanceStatement", reflect.TypeOf((*MockState)(nil).InitialWatchInstanceStatement))
	return &MockStateInitialWatchInstanceStatementCall{Call: call}
}

// MockStateInitialWatchInstanceStatementCall wrap *gomock.Call
type MockStateInitialWatchInstanceStatementCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateInitialWatchInstanceStatementCall) Return(arg0, arg1 string) *MockStateInitialWatchInstanceStatementCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateInitialWatchInstanceStatementCall) Do(f func() (string, string)) *MockStateInitialWatchInstanceStatementCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateInitialWatchInstanceStatementCall) DoAndReturn(f func() (string, string)) *MockStateInitialWatchInstanceStatementCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// InitialWatchStatement mocks base method.
func (m *MockState) InitialWatchStatement() (string, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitialWatchStatement")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// InitialWatchStatement indicates an expected call of InitialWatchStatement.
func (mr *MockStateMockRecorder) InitialWatchStatement() *MockStateInitialWatchStatementCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitialWatchStatement", reflect.TypeOf((*MockState)(nil).InitialWatchStatement))
	return &MockStateInitialWatchStatementCall{Call: call}
}

// MockStateInitialWatchStatementCall wrap *gomock.Call
type MockStateInitialWatchStatementCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateInitialWatchStatementCall) Return(arg0, arg1 string) *MockStateInitialWatchStatementCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateInitialWatchStatementCall) Do(f func() (string, string)) *MockStateInitialWatchStatementCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateInitialWatchStatementCall) DoAndReturn(f func() (string, string)) *MockStateInitialWatchStatementCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// InstanceId mocks base method.
func (m *MockState) InstanceId(arg0 context.Context, arg1 machine.Name) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstanceId", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InstanceId indicates an expected call of InstanceId.
func (mr *MockStateMockRecorder) InstanceId(arg0, arg1 any) *MockStateInstanceIdCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstanceId", reflect.TypeOf((*MockState)(nil).InstanceId), arg0, arg1)
	return &MockStateInstanceIdCall{Call: call}
}

// MockStateInstanceIdCall wrap *gomock.Call
type MockStateInstanceIdCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateInstanceIdCall) Return(arg0 string, arg1 error) *MockStateInstanceIdCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateInstanceIdCall) Do(f func(context.Context, machine.Name) (string, error)) *MockStateInstanceIdCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateInstanceIdCall) DoAndReturn(f func(context.Context, machine.Name) (string, error)) *MockStateInstanceIdCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// InstanceStatus mocks base method.
func (m *MockState) InstanceStatus(arg0 context.Context, arg1 machine.Name) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstanceStatus", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InstanceStatus indicates an expected call of InstanceStatus.
func (mr *MockStateMockRecorder) InstanceStatus(arg0, arg1 any) *MockStateInstanceStatusCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstanceStatus", reflect.TypeOf((*MockState)(nil).InstanceStatus), arg0, arg1)
	return &MockStateInstanceStatusCall{Call: call}
}

// MockStateInstanceStatusCall wrap *gomock.Call
type MockStateInstanceStatusCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateInstanceStatusCall) Return(arg0 string, arg1 error) *MockStateInstanceStatusCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateInstanceStatusCall) Do(f func(context.Context, machine.Name) (string, error)) *MockStateInstanceStatusCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateInstanceStatusCall) DoAndReturn(f func(context.Context, machine.Name) (string, error)) *MockStateInstanceStatusCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// IsMachineRebootRequired mocks base method.
func (m *MockState) IsMachineRebootRequired(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsMachineRebootRequired", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsMachineRebootRequired indicates an expected call of IsMachineRebootRequired.
func (mr *MockStateMockRecorder) IsMachineRebootRequired(arg0, arg1 any) *MockStateIsMachineRebootRequiredCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsMachineRebootRequired", reflect.TypeOf((*MockState)(nil).IsMachineRebootRequired), arg0, arg1)
	return &MockStateIsMachineRebootRequiredCall{Call: call}
}

// MockStateIsMachineRebootRequiredCall wrap *gomock.Call
type MockStateIsMachineRebootRequiredCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateIsMachineRebootRequiredCall) Return(arg0 bool, arg1 error) *MockStateIsMachineRebootRequiredCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateIsMachineRebootRequiredCall) Do(f func(context.Context, string) (bool, error)) *MockStateIsMachineRebootRequiredCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateIsMachineRebootRequiredCall) DoAndReturn(f func(context.Context, string) (bool, error)) *MockStateIsMachineRebootRequiredCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RequireMachineReboot mocks base method.
func (m *MockState) RequireMachineReboot(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequireMachineReboot", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RequireMachineReboot indicates an expected call of RequireMachineReboot.
func (mr *MockStateMockRecorder) RequireMachineReboot(arg0, arg1 any) *MockStateRequireMachineRebootCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequireMachineReboot", reflect.TypeOf((*MockState)(nil).RequireMachineReboot), arg0, arg1)
	return &MockStateRequireMachineRebootCall{Call: call}
}

// MockStateRequireMachineRebootCall wrap *gomock.Call
type MockStateRequireMachineRebootCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateRequireMachineRebootCall) Return(arg0 error) *MockStateRequireMachineRebootCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateRequireMachineRebootCall) Do(f func(context.Context, string) error) *MockStateRequireMachineRebootCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateRequireMachineRebootCall) DoAndReturn(f func(context.Context, string) error) *MockStateRequireMachineRebootCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetMachineCloudInstance mocks base method.
func (m *MockState) SetMachineCloudInstance(arg0 context.Context, arg1 string, arg2 instance.Id, arg3 instance.HardwareCharacteristics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMachineCloudInstance", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetMachineCloudInstance indicates an expected call of SetMachineCloudInstance.
func (mr *MockStateMockRecorder) SetMachineCloudInstance(arg0, arg1, arg2, arg3 any) *MockStateSetMachineCloudInstanceCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMachineCloudInstance", reflect.TypeOf((*MockState)(nil).SetMachineCloudInstance), arg0, arg1, arg2, arg3)
	return &MockStateSetMachineCloudInstanceCall{Call: call}
}

// MockStateSetMachineCloudInstanceCall wrap *gomock.Call
type MockStateSetMachineCloudInstanceCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateSetMachineCloudInstanceCall) Return(arg0 error) *MockStateSetMachineCloudInstanceCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateSetMachineCloudInstanceCall) Do(f func(context.Context, string, instance.Id, instance.HardwareCharacteristics) error) *MockStateSetMachineCloudInstanceCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateSetMachineCloudInstanceCall) DoAndReturn(f func(context.Context, string, instance.Id, instance.HardwareCharacteristics) error) *MockStateSetMachineCloudInstanceCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
