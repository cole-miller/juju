// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/cmd/juju/application (interfaces: ApplicationAPI,RemoveApplicationAPI)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	application "github.com/juju/juju/api/client/application"
	params "github.com/juju/juju/rpc/params"
)

// MockApplicationAPI is a mock of ApplicationAPI interface.
type MockApplicationAPI struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationAPIMockRecorder
}

// MockApplicationAPIMockRecorder is the mock recorder for MockApplicationAPI.
type MockApplicationAPIMockRecorder struct {
	mock *MockApplicationAPI
}

// NewMockApplicationAPI creates a new mock instance.
func NewMockApplicationAPI(ctrl *gomock.Controller) *MockApplicationAPI {
	mock := &MockApplicationAPI{ctrl: ctrl}
	mock.recorder = &MockApplicationAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationAPI) EXPECT() *MockApplicationAPIMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockApplicationAPI) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockApplicationAPIMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockApplicationAPI)(nil).Close))
}

// Get mocks base method.
func (m *MockApplicationAPI) Get(arg0, arg1 string) (*params.ApplicationGetResults, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*params.ApplicationGetResults)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockApplicationAPIMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockApplicationAPI)(nil).Get), arg0, arg1)
}

// SetConfig mocks base method.
func (m *MockApplicationAPI) SetConfig(arg0, arg1, arg2 string, arg3 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetConfig", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetConfig indicates an expected call of SetConfig.
func (mr *MockApplicationAPIMockRecorder) SetConfig(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConfig", reflect.TypeOf((*MockApplicationAPI)(nil).SetConfig), arg0, arg1, arg2, arg3)
}

// UnsetApplicationConfig mocks base method.
func (m *MockApplicationAPI) UnsetApplicationConfig(arg0, arg1 string, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnsetApplicationConfig", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnsetApplicationConfig indicates an expected call of UnsetApplicationConfig.
func (mr *MockApplicationAPIMockRecorder) UnsetApplicationConfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnsetApplicationConfig", reflect.TypeOf((*MockApplicationAPI)(nil).UnsetApplicationConfig), arg0, arg1, arg2)
}

// MockRemoveApplicationAPI is a mock of RemoveApplicationAPI interface.
type MockRemoveApplicationAPI struct {
	ctrl     *gomock.Controller
	recorder *MockRemoveApplicationAPIMockRecorder
}

// MockRemoveApplicationAPIMockRecorder is the mock recorder for MockRemoveApplicationAPI.
type MockRemoveApplicationAPIMockRecorder struct {
	mock *MockRemoveApplicationAPI
}

// NewMockRemoveApplicationAPI creates a new mock instance.
func NewMockRemoveApplicationAPI(ctrl *gomock.Controller) *MockRemoveApplicationAPI {
	mock := &MockRemoveApplicationAPI{ctrl: ctrl}
	mock.recorder = &MockRemoveApplicationAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRemoveApplicationAPI) EXPECT() *MockRemoveApplicationAPIMockRecorder {
	return m.recorder
}

// BestAPIVersion mocks base method.
func (m *MockRemoveApplicationAPI) BestAPIVersion() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BestAPIVersion")
	ret0, _ := ret[0].(int)
	return ret0
}

// BestAPIVersion indicates an expected call of BestAPIVersion.
func (mr *MockRemoveApplicationAPIMockRecorder) BestAPIVersion() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BestAPIVersion", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).BestAPIVersion))
}

// Close mocks base method.
func (m *MockRemoveApplicationAPI) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRemoveApplicationAPIMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).Close))
}

// DestroyApplications mocks base method.
func (m *MockRemoveApplicationAPI) DestroyApplications(arg0 application.DestroyApplicationsParams) ([]params.DestroyApplicationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroyApplications", arg0)
	ret0, _ := ret[0].([]params.DestroyApplicationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DestroyApplications indicates an expected call of DestroyApplications.
func (mr *MockRemoveApplicationAPIMockRecorder) DestroyApplications(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroyApplications", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).DestroyApplications), arg0)
}

// DestroyUnits mocks base method.
func (m *MockRemoveApplicationAPI) DestroyUnits(arg0 application.DestroyUnitsParams) ([]params.DestroyUnitResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroyUnits", arg0)
	ret0, _ := ret[0].([]params.DestroyUnitResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DestroyUnits indicates an expected call of DestroyUnits.
func (mr *MockRemoveApplicationAPIMockRecorder) DestroyUnits(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroyUnits", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).DestroyUnits), arg0)
}

// ModelUUID mocks base method.
func (m *MockRemoveApplicationAPI) ModelUUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelUUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ModelUUID indicates an expected call of ModelUUID.
func (mr *MockRemoveApplicationAPIMockRecorder) ModelUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelUUID", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).ModelUUID))
}

// ScaleApplication mocks base method.
func (m *MockRemoveApplicationAPI) ScaleApplication(arg0 application.ScaleApplicationParams) (params.ScaleApplicationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScaleApplication", arg0)
	ret0, _ := ret[0].(params.ScaleApplicationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ScaleApplication indicates an expected call of ScaleApplication.
func (mr *MockRemoveApplicationAPIMockRecorder) ScaleApplication(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScaleApplication", reflect.TypeOf((*MockRemoveApplicationAPI)(nil).ScaleApplication), arg0)
}
