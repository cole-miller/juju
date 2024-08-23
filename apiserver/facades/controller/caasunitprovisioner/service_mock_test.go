// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/controller/caasunitprovisioner (interfaces: ApplicationService)
//
// Generated by this command:
//
//	mockgen -package caasunitprovisioner_test -destination service_mock_test.go github.com/juju/juju/apiserver/facades/controller/caasunitprovisioner ApplicationService
//

// Package caasunitprovisioner_test is a generated GoMock package.
package caasunitprovisioner_test

import (
	context "context"
	reflect "reflect"

	network "github.com/juju/juju/core/network"
	watcher "github.com/juju/juju/core/watcher"
	gomock "go.uber.org/mock/gomock"
)

// MockApplicationService is a mock of ApplicationService interface.
type MockApplicationService struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationServiceMockRecorder
}

// MockApplicationServiceMockRecorder is the mock recorder for MockApplicationService.
type MockApplicationServiceMockRecorder struct {
	mock *MockApplicationService
}

// NewMockApplicationService creates a new mock instance.
func NewMockApplicationService(ctrl *gomock.Controller) *MockApplicationService {
	mock := &MockApplicationService{ctrl: ctrl}
	mock.recorder = &MockApplicationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationService) EXPECT() *MockApplicationServiceMockRecorder {
	return m.recorder
}

// GetApplicationScale mocks base method.
func (m *MockApplicationService) GetApplicationScale(arg0 context.Context, arg1 string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationScale", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationScale indicates an expected call of GetApplicationScale.
func (mr *MockApplicationServiceMockRecorder) GetApplicationScale(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationScale", reflect.TypeOf((*MockApplicationService)(nil).GetApplicationScale), arg0, arg1)
}

// SetApplicationScale mocks base method.
func (m *MockApplicationService) SetApplicationScale(arg0 context.Context, arg1 string, arg2 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetApplicationScale", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetApplicationScale indicates an expected call of SetApplicationScale.
func (mr *MockApplicationServiceMockRecorder) SetApplicationScale(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetApplicationScale", reflect.TypeOf((*MockApplicationService)(nil).SetApplicationScale), arg0, arg1, arg2)
}

// UpdateCloudService mocks base method.
func (m *MockApplicationService) UpdateCloudService(arg0 context.Context, arg1, arg2 string, arg3 network.SpaceAddresses) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCloudService", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCloudService indicates an expected call of UpdateCloudService.
func (mr *MockApplicationServiceMockRecorder) UpdateCloudService(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCloudService", reflect.TypeOf((*MockApplicationService)(nil).UpdateCloudService), arg0, arg1, arg2, arg3)
}

// WatchApplicationScale mocks base method.
func (m *MockApplicationService) WatchApplicationScale(arg0 context.Context, arg1 string) (watcher.Watcher[struct{}], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchApplicationScale", arg0, arg1)
	ret0, _ := ret[0].(watcher.Watcher[struct{}])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchApplicationScale indicates an expected call of WatchApplicationScale.
func (mr *MockApplicationServiceMockRecorder) WatchApplicationScale(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchApplicationScale", reflect.TypeOf((*MockApplicationService)(nil).WatchApplicationScale), arg0, arg1)
}
