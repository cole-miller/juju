// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/agent/secretsmanager (interfaces: ModelState)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	config "github.com/juju/juju/environs/config"
	state "github.com/juju/juju/state"
)

// MockModelState is a mock of ModelState interface.
type MockModelState struct {
	ctrl     *gomock.Controller
	recorder *MockModelStateMockRecorder
}

// MockModelStateMockRecorder is the mock recorder for MockModelState.
type MockModelStateMockRecorder struct {
	mock *MockModelState
}

// NewMockModelState creates a new mock instance.
func NewMockModelState(ctrl *gomock.Controller) *MockModelState {
	mock := &MockModelState{ctrl: ctrl}
	mock.recorder = &MockModelStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModelState) EXPECT() *MockModelStateMockRecorder {
	return m.recorder
}

// ControllerUUID mocks base method.
func (m *MockModelState) ControllerUUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ControllerUUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ControllerUUID indicates an expected call of ControllerUUID.
func (mr *MockModelStateMockRecorder) ControllerUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ControllerUUID", reflect.TypeOf((*MockModelState)(nil).ControllerUUID))
}

// ModelConfig mocks base method.
func (m *MockModelState) ModelConfig() (*config.Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelConfig")
	ret0, _ := ret[0].(*config.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ModelConfig indicates an expected call of ModelConfig.
func (mr *MockModelStateMockRecorder) ModelConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelConfig", reflect.TypeOf((*MockModelState)(nil).ModelConfig))
}

// Type mocks base method.
func (m *MockModelState) Type() state.ModelType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Type")
	ret0, _ := ret[0].(state.ModelType)
	return ret0
}

// Type indicates an expected call of Type.
func (mr *MockModelStateMockRecorder) Type() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Type", reflect.TypeOf((*MockModelState)(nil).Type))
}

// UUID mocks base method.
func (m *MockModelState) UUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// UUID indicates an expected call of UUID.
func (mr *MockModelStateMockRecorder) UUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UUID", reflect.TypeOf((*MockModelState)(nil).UUID))
}

// WatchForModelConfigChanges mocks base method.
func (m *MockModelState) WatchForModelConfigChanges() state.NotifyWatcher {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchForModelConfigChanges")
	ret0, _ := ret[0].(state.NotifyWatcher)
	return ret0
}

// WatchForModelConfigChanges indicates an expected call of WatchForModelConfigChanges.
func (mr *MockModelStateMockRecorder) WatchForModelConfigChanges() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchForModelConfigChanges", reflect.TypeOf((*MockModelState)(nil).WatchForModelConfigChanges))
}
