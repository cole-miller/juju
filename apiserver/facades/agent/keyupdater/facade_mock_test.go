// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facade (interfaces: ModelContext)
//
// Generated by this command:
//
//	mockgen -typed -package keyupdater -destination facade_mock_test.go github.com/juju/juju/apiserver/facade ModelContext
//

// Package keyupdater is a generated GoMock package.
package keyupdater

import (
	context "context"
	reflect "reflect"

	facade "github.com/juju/juju/apiserver/facade"
	leadership "github.com/juju/juju/core/leadership"
	lease "github.com/juju/juju/core/lease"
	logger "github.com/juju/juju/core/logger"
	objectstore "github.com/juju/juju/core/objectstore"
	providertracker "github.com/juju/juju/core/providertracker"
	servicefactory "github.com/juju/juju/internal/servicefactory"
	state "github.com/juju/juju/state"
	names "github.com/juju/names/v5"
	gomock "go.uber.org/mock/gomock"
)

// MockModelContext is a mock of ModelContext interface.
type MockModelContext struct {
	ctrl     *gomock.Controller
	recorder *MockModelContextMockRecorder
}

// MockModelContextMockRecorder is the mock recorder for MockModelContext.
type MockModelContextMockRecorder struct {
	mock *MockModelContext
}

// NewMockModelContext creates a new mock instance.
func NewMockModelContext(ctrl *gomock.Controller) *MockModelContext {
	mock := &MockModelContext{ctrl: ctrl}
	mock.recorder = &MockModelContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModelContext) EXPECT() *MockModelContextMockRecorder {
	return m.recorder
}

// Auth mocks base method.
func (m *MockModelContext) Auth() facade.Authorizer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Auth")
	ret0, _ := ret[0].(facade.Authorizer)
	return ret0
}

// Auth indicates an expected call of Auth.
func (mr *MockModelContextMockRecorder) Auth() *MockModelContextAuthCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Auth", reflect.TypeOf((*MockModelContext)(nil).Auth))
	return &MockModelContextAuthCall{Call: call}
}

// MockModelContextAuthCall wrap *gomock.Call
type MockModelContextAuthCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextAuthCall) Return(arg0 facade.Authorizer) *MockModelContextAuthCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextAuthCall) Do(f func() facade.Authorizer) *MockModelContextAuthCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextAuthCall) DoAndReturn(f func() facade.Authorizer) *MockModelContextAuthCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ControllerObjectStore mocks base method.
func (m *MockModelContext) ControllerObjectStore() objectstore.ObjectStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ControllerObjectStore")
	ret0, _ := ret[0].(objectstore.ObjectStore)
	return ret0
}

// ControllerObjectStore indicates an expected call of ControllerObjectStore.
func (mr *MockModelContextMockRecorder) ControllerObjectStore() *MockModelContextControllerObjectStoreCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ControllerObjectStore", reflect.TypeOf((*MockModelContext)(nil).ControllerObjectStore))
	return &MockModelContextControllerObjectStoreCall{Call: call}
}

// MockModelContextControllerObjectStoreCall wrap *gomock.Call
type MockModelContextControllerObjectStoreCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextControllerObjectStoreCall) Return(arg0 objectstore.ObjectStore) *MockModelContextControllerObjectStoreCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextControllerObjectStoreCall) Do(f func() objectstore.ObjectStore) *MockModelContextControllerObjectStoreCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextControllerObjectStoreCall) DoAndReturn(f func() objectstore.ObjectStore) *MockModelContextControllerObjectStoreCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ControllerUUID mocks base method.
func (m *MockModelContext) ControllerUUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ControllerUUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ControllerUUID indicates an expected call of ControllerUUID.
func (mr *MockModelContextMockRecorder) ControllerUUID() *MockModelContextControllerUUIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ControllerUUID", reflect.TypeOf((*MockModelContext)(nil).ControllerUUID))
	return &MockModelContextControllerUUIDCall{Call: call}
}

// MockModelContextControllerUUIDCall wrap *gomock.Call
type MockModelContextControllerUUIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextControllerUUIDCall) Return(arg0 string) *MockModelContextControllerUUIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextControllerUUIDCall) Do(f func() string) *MockModelContextControllerUUIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextControllerUUIDCall) DoAndReturn(f func() string) *MockModelContextControllerUUIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DataDir mocks base method.
func (m *MockModelContext) DataDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// DataDir indicates an expected call of DataDir.
func (mr *MockModelContextMockRecorder) DataDir() *MockModelContextDataDirCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataDir", reflect.TypeOf((*MockModelContext)(nil).DataDir))
	return &MockModelContextDataDirCall{Call: call}
}

// MockModelContextDataDirCall wrap *gomock.Call
type MockModelContextDataDirCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextDataDirCall) Return(arg0 string) *MockModelContextDataDirCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextDataDirCall) Do(f func() string) *MockModelContextDataDirCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextDataDirCall) DoAndReturn(f func() string) *MockModelContextDataDirCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Dispose mocks base method.
func (m *MockModelContext) Dispose() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Dispose")
}

// Dispose indicates an expected call of Dispose.
func (mr *MockModelContextMockRecorder) Dispose() *MockModelContextDisposeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispose", reflect.TypeOf((*MockModelContext)(nil).Dispose))
	return &MockModelContextDisposeCall{Call: call}
}

// MockModelContextDisposeCall wrap *gomock.Call
type MockModelContextDisposeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextDisposeCall) Return() *MockModelContextDisposeCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextDisposeCall) Do(f func()) *MockModelContextDisposeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextDisposeCall) DoAndReturn(f func()) *MockModelContextDisposeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetProvider mocks base method.
func (m *MockModelContext) GetProvider(arg0 context.Context) (providertracker.Provider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProvider", arg0)
	ret0, _ := ret[0].(providertracker.Provider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProvider indicates an expected call of GetProvider.
func (mr *MockModelContextMockRecorder) GetProvider(arg0 any) *MockModelContextGetProviderCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProvider", reflect.TypeOf((*MockModelContext)(nil).GetProvider), arg0)
	return &MockModelContextGetProviderCall{Call: call}
}

// MockModelContextGetProviderCall wrap *gomock.Call
type MockModelContextGetProviderCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextGetProviderCall) Return(arg0 providertracker.Provider, arg1 error) *MockModelContextGetProviderCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextGetProviderCall) Do(f func(context.Context) (providertracker.Provider, error)) *MockModelContextGetProviderCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextGetProviderCall) DoAndReturn(f func(context.Context) (providertracker.Provider, error)) *MockModelContextGetProviderCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// HTTPClient mocks base method.
func (m *MockModelContext) HTTPClient(arg0 facade.HTTPClientPurpose) (facade.HTTPClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HTTPClient", arg0)
	ret0, _ := ret[0].(facade.HTTPClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HTTPClient indicates an expected call of HTTPClient.
func (mr *MockModelContextMockRecorder) HTTPClient(arg0 any) *MockModelContextHTTPClientCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HTTPClient", reflect.TypeOf((*MockModelContext)(nil).HTTPClient), arg0)
	return &MockModelContextHTTPClientCall{Call: call}
}

// MockModelContextHTTPClientCall wrap *gomock.Call
type MockModelContextHTTPClientCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextHTTPClientCall) Return(arg0 facade.HTTPClient, arg1 error) *MockModelContextHTTPClientCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextHTTPClientCall) Do(f func(facade.HTTPClientPurpose) (facade.HTTPClient, error)) *MockModelContextHTTPClientCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextHTTPClientCall) DoAndReturn(f func(facade.HTTPClientPurpose) (facade.HTTPClient, error)) *MockModelContextHTTPClientCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Hub mocks base method.
func (m *MockModelContext) Hub() facade.Hub {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Hub")
	ret0, _ := ret[0].(facade.Hub)
	return ret0
}

// Hub indicates an expected call of Hub.
func (mr *MockModelContextMockRecorder) Hub() *MockModelContextHubCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hub", reflect.TypeOf((*MockModelContext)(nil).Hub))
	return &MockModelContextHubCall{Call: call}
}

// MockModelContextHubCall wrap *gomock.Call
type MockModelContextHubCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextHubCall) Return(arg0 facade.Hub) *MockModelContextHubCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextHubCall) Do(f func() facade.Hub) *MockModelContextHubCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextHubCall) DoAndReturn(f func() facade.Hub) *MockModelContextHubCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ID mocks base method.
func (m *MockModelContext) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockModelContextMockRecorder) ID() *MockModelContextIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockModelContext)(nil).ID))
	return &MockModelContextIDCall{Call: call}
}

// MockModelContextIDCall wrap *gomock.Call
type MockModelContextIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextIDCall) Return(arg0 string) *MockModelContextIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextIDCall) Do(f func() string) *MockModelContextIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextIDCall) DoAndReturn(f func() string) *MockModelContextIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LeadershipChecker mocks base method.
func (m *MockModelContext) LeadershipChecker() (leadership.Checker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeadershipChecker")
	ret0, _ := ret[0].(leadership.Checker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeadershipChecker indicates an expected call of LeadershipChecker.
func (mr *MockModelContextMockRecorder) LeadershipChecker() *MockModelContextLeadershipCheckerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeadershipChecker", reflect.TypeOf((*MockModelContext)(nil).LeadershipChecker))
	return &MockModelContextLeadershipCheckerCall{Call: call}
}

// MockModelContextLeadershipCheckerCall wrap *gomock.Call
type MockModelContextLeadershipCheckerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLeadershipCheckerCall) Return(arg0 leadership.Checker, arg1 error) *MockModelContextLeadershipCheckerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLeadershipCheckerCall) Do(f func() (leadership.Checker, error)) *MockModelContextLeadershipCheckerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLeadershipCheckerCall) DoAndReturn(f func() (leadership.Checker, error)) *MockModelContextLeadershipCheckerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LeadershipClaimer mocks base method.
func (m *MockModelContext) LeadershipClaimer() (leadership.Claimer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeadershipClaimer")
	ret0, _ := ret[0].(leadership.Claimer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeadershipClaimer indicates an expected call of LeadershipClaimer.
func (mr *MockModelContextMockRecorder) LeadershipClaimer() *MockModelContextLeadershipClaimerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeadershipClaimer", reflect.TypeOf((*MockModelContext)(nil).LeadershipClaimer))
	return &MockModelContextLeadershipClaimerCall{Call: call}
}

// MockModelContextLeadershipClaimerCall wrap *gomock.Call
type MockModelContextLeadershipClaimerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLeadershipClaimerCall) Return(arg0 leadership.Claimer, arg1 error) *MockModelContextLeadershipClaimerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLeadershipClaimerCall) Do(f func() (leadership.Claimer, error)) *MockModelContextLeadershipClaimerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLeadershipClaimerCall) DoAndReturn(f func() (leadership.Claimer, error)) *MockModelContextLeadershipClaimerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LeadershipPinner mocks base method.
func (m *MockModelContext) LeadershipPinner() (leadership.Pinner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeadershipPinner")
	ret0, _ := ret[0].(leadership.Pinner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeadershipPinner indicates an expected call of LeadershipPinner.
func (mr *MockModelContextMockRecorder) LeadershipPinner() *MockModelContextLeadershipPinnerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeadershipPinner", reflect.TypeOf((*MockModelContext)(nil).LeadershipPinner))
	return &MockModelContextLeadershipPinnerCall{Call: call}
}

// MockModelContextLeadershipPinnerCall wrap *gomock.Call
type MockModelContextLeadershipPinnerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLeadershipPinnerCall) Return(arg0 leadership.Pinner, arg1 error) *MockModelContextLeadershipPinnerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLeadershipPinnerCall) Do(f func() (leadership.Pinner, error)) *MockModelContextLeadershipPinnerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLeadershipPinnerCall) DoAndReturn(f func() (leadership.Pinner, error)) *MockModelContextLeadershipPinnerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LeadershipReader mocks base method.
func (m *MockModelContext) LeadershipReader() (leadership.Reader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeadershipReader")
	ret0, _ := ret[0].(leadership.Reader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeadershipReader indicates an expected call of LeadershipReader.
func (mr *MockModelContextMockRecorder) LeadershipReader() *MockModelContextLeadershipReaderCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeadershipReader", reflect.TypeOf((*MockModelContext)(nil).LeadershipReader))
	return &MockModelContextLeadershipReaderCall{Call: call}
}

// MockModelContextLeadershipReaderCall wrap *gomock.Call
type MockModelContextLeadershipReaderCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLeadershipReaderCall) Return(arg0 leadership.Reader, arg1 error) *MockModelContextLeadershipReaderCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLeadershipReaderCall) Do(f func() (leadership.Reader, error)) *MockModelContextLeadershipReaderCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLeadershipReaderCall) DoAndReturn(f func() (leadership.Reader, error)) *MockModelContextLeadershipReaderCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LeadershipRevoker mocks base method.
func (m *MockModelContext) LeadershipRevoker() (leadership.Revoker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeadershipRevoker")
	ret0, _ := ret[0].(leadership.Revoker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeadershipRevoker indicates an expected call of LeadershipRevoker.
func (mr *MockModelContextMockRecorder) LeadershipRevoker() *MockModelContextLeadershipRevokerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeadershipRevoker", reflect.TypeOf((*MockModelContext)(nil).LeadershipRevoker))
	return &MockModelContextLeadershipRevokerCall{Call: call}
}

// MockModelContextLeadershipRevokerCall wrap *gomock.Call
type MockModelContextLeadershipRevokerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLeadershipRevokerCall) Return(arg0 leadership.Revoker, arg1 error) *MockModelContextLeadershipRevokerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLeadershipRevokerCall) Do(f func() (leadership.Revoker, error)) *MockModelContextLeadershipRevokerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLeadershipRevokerCall) DoAndReturn(f func() (leadership.Revoker, error)) *MockModelContextLeadershipRevokerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LogDir mocks base method.
func (m *MockModelContext) LogDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// LogDir indicates an expected call of LogDir.
func (mr *MockModelContextMockRecorder) LogDir() *MockModelContextLogDirCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogDir", reflect.TypeOf((*MockModelContext)(nil).LogDir))
	return &MockModelContextLogDirCall{Call: call}
}

// MockModelContextLogDirCall wrap *gomock.Call
type MockModelContextLogDirCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLogDirCall) Return(arg0 string) *MockModelContextLogDirCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLogDirCall) Do(f func() string) *MockModelContextLogDirCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLogDirCall) DoAndReturn(f func() string) *MockModelContextLogDirCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Logger mocks base method.
func (m *MockModelContext) Logger() logger.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logger")
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// Logger indicates an expected call of Logger.
func (mr *MockModelContextMockRecorder) Logger() *MockModelContextLoggerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logger", reflect.TypeOf((*MockModelContext)(nil).Logger))
	return &MockModelContextLoggerCall{Call: call}
}

// MockModelContextLoggerCall wrap *gomock.Call
type MockModelContextLoggerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextLoggerCall) Return(arg0 logger.Logger) *MockModelContextLoggerCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextLoggerCall) Do(f func() logger.Logger) *MockModelContextLoggerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextLoggerCall) DoAndReturn(f func() logger.Logger) *MockModelContextLoggerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MachineTag mocks base method.
func (m *MockModelContext) MachineTag() names.Tag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineTag")
	ret0, _ := ret[0].(names.Tag)
	return ret0
}

// MachineTag indicates an expected call of MachineTag.
func (mr *MockModelContextMockRecorder) MachineTag() *MockModelContextMachineTagCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineTag", reflect.TypeOf((*MockModelContext)(nil).MachineTag))
	return &MockModelContextMachineTagCall{Call: call}
}

// MockModelContextMachineTagCall wrap *gomock.Call
type MockModelContextMachineTagCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextMachineTagCall) Return(arg0 names.Tag) *MockModelContextMachineTagCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextMachineTagCall) Do(f func() names.Tag) *MockModelContextMachineTagCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextMachineTagCall) DoAndReturn(f func() names.Tag) *MockModelContextMachineTagCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ModelExporter mocks base method.
func (m *MockModelContext) ModelExporter(arg0 facade.LegacyStateExporter) facade.ModelExporter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelExporter", arg0)
	ret0, _ := ret[0].(facade.ModelExporter)
	return ret0
}

// ModelExporter indicates an expected call of ModelExporter.
func (mr *MockModelContextMockRecorder) ModelExporter(arg0 any) *MockModelContextModelExporterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelExporter", reflect.TypeOf((*MockModelContext)(nil).ModelExporter), arg0)
	return &MockModelContextModelExporterCall{Call: call}
}

// MockModelContextModelExporterCall wrap *gomock.Call
type MockModelContextModelExporterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextModelExporterCall) Return(arg0 facade.ModelExporter) *MockModelContextModelExporterCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextModelExporterCall) Do(f func(facade.LegacyStateExporter) facade.ModelExporter) *MockModelContextModelExporterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextModelExporterCall) DoAndReturn(f func(facade.LegacyStateExporter) facade.ModelExporter) *MockModelContextModelExporterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ModelImporter mocks base method.
func (m *MockModelContext) ModelImporter() facade.ModelImporter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelImporter")
	ret0, _ := ret[0].(facade.ModelImporter)
	return ret0
}

// ModelImporter indicates an expected call of ModelImporter.
func (mr *MockModelContextMockRecorder) ModelImporter() *MockModelContextModelImporterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelImporter", reflect.TypeOf((*MockModelContext)(nil).ModelImporter))
	return &MockModelContextModelImporterCall{Call: call}
}

// MockModelContextModelImporterCall wrap *gomock.Call
type MockModelContextModelImporterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextModelImporterCall) Return(arg0 facade.ModelImporter) *MockModelContextModelImporterCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextModelImporterCall) Do(f func() facade.ModelImporter) *MockModelContextModelImporterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextModelImporterCall) DoAndReturn(f func() facade.ModelImporter) *MockModelContextModelImporterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ObjectStore mocks base method.
func (m *MockModelContext) ObjectStore() objectstore.ObjectStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObjectStore")
	ret0, _ := ret[0].(objectstore.ObjectStore)
	return ret0
}

// ObjectStore indicates an expected call of ObjectStore.
func (mr *MockModelContextMockRecorder) ObjectStore() *MockModelContextObjectStoreCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObjectStore", reflect.TypeOf((*MockModelContext)(nil).ObjectStore))
	return &MockModelContextObjectStoreCall{Call: call}
}

// MockModelContextObjectStoreCall wrap *gomock.Call
type MockModelContextObjectStoreCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextObjectStoreCall) Return(arg0 objectstore.ObjectStore) *MockModelContextObjectStoreCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextObjectStoreCall) Do(f func() objectstore.ObjectStore) *MockModelContextObjectStoreCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextObjectStoreCall) DoAndReturn(f func() objectstore.ObjectStore) *MockModelContextObjectStoreCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Presence mocks base method.
func (m *MockModelContext) Presence() facade.Presence {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Presence")
	ret0, _ := ret[0].(facade.Presence)
	return ret0
}

// Presence indicates an expected call of Presence.
func (mr *MockModelContextMockRecorder) Presence() *MockModelContextPresenceCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Presence", reflect.TypeOf((*MockModelContext)(nil).Presence))
	return &MockModelContextPresenceCall{Call: call}
}

// MockModelContextPresenceCall wrap *gomock.Call
type MockModelContextPresenceCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextPresenceCall) Return(arg0 facade.Presence) *MockModelContextPresenceCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextPresenceCall) Do(f func() facade.Presence) *MockModelContextPresenceCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextPresenceCall) DoAndReturn(f func() facade.Presence) *MockModelContextPresenceCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RequestRecorder mocks base method.
func (m *MockModelContext) RequestRecorder() facade.RequestRecorder {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestRecorder")
	ret0, _ := ret[0].(facade.RequestRecorder)
	return ret0
}

// RequestRecorder indicates an expected call of RequestRecorder.
func (mr *MockModelContextMockRecorder) RequestRecorder() *MockModelContextRequestRecorderCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestRecorder", reflect.TypeOf((*MockModelContext)(nil).RequestRecorder))
	return &MockModelContextRequestRecorderCall{Call: call}
}

// MockModelContextRequestRecorderCall wrap *gomock.Call
type MockModelContextRequestRecorderCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextRequestRecorderCall) Return(arg0 facade.RequestRecorder) *MockModelContextRequestRecorderCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextRequestRecorderCall) Do(f func() facade.RequestRecorder) *MockModelContextRequestRecorderCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextRequestRecorderCall) DoAndReturn(f func() facade.RequestRecorder) *MockModelContextRequestRecorderCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Resources mocks base method.
func (m *MockModelContext) Resources() facade.Resources {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resources")
	ret0, _ := ret[0].(facade.Resources)
	return ret0
}

// Resources indicates an expected call of Resources.
func (mr *MockModelContextMockRecorder) Resources() *MockModelContextResourcesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resources", reflect.TypeOf((*MockModelContext)(nil).Resources))
	return &MockModelContextResourcesCall{Call: call}
}

// MockModelContextResourcesCall wrap *gomock.Call
type MockModelContextResourcesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextResourcesCall) Return(arg0 facade.Resources) *MockModelContextResourcesCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextResourcesCall) Do(f func() facade.Resources) *MockModelContextResourcesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextResourcesCall) DoAndReturn(f func() facade.Resources) *MockModelContextResourcesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ServiceFactory mocks base method.
func (m *MockModelContext) ServiceFactory() servicefactory.ServiceFactory {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceFactory")
	ret0, _ := ret[0].(servicefactory.ServiceFactory)
	return ret0
}

// ServiceFactory indicates an expected call of ServiceFactory.
func (mr *MockModelContextMockRecorder) ServiceFactory() *MockModelContextServiceFactoryCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceFactory", reflect.TypeOf((*MockModelContext)(nil).ServiceFactory))
	return &MockModelContextServiceFactoryCall{Call: call}
}

// MockModelContextServiceFactoryCall wrap *gomock.Call
type MockModelContextServiceFactoryCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextServiceFactoryCall) Return(arg0 servicefactory.ServiceFactory) *MockModelContextServiceFactoryCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextServiceFactoryCall) Do(f func() servicefactory.ServiceFactory) *MockModelContextServiceFactoryCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextServiceFactoryCall) DoAndReturn(f func() servicefactory.ServiceFactory) *MockModelContextServiceFactoryCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SingularClaimer mocks base method.
func (m *MockModelContext) SingularClaimer() (lease.Claimer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SingularClaimer")
	ret0, _ := ret[0].(lease.Claimer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SingularClaimer indicates an expected call of SingularClaimer.
func (mr *MockModelContextMockRecorder) SingularClaimer() *MockModelContextSingularClaimerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SingularClaimer", reflect.TypeOf((*MockModelContext)(nil).SingularClaimer))
	return &MockModelContextSingularClaimerCall{Call: call}
}

// MockModelContextSingularClaimerCall wrap *gomock.Call
type MockModelContextSingularClaimerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextSingularClaimerCall) Return(arg0 lease.Claimer, arg1 error) *MockModelContextSingularClaimerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextSingularClaimerCall) Do(f func() (lease.Claimer, error)) *MockModelContextSingularClaimerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextSingularClaimerCall) DoAndReturn(f func() (lease.Claimer, error)) *MockModelContextSingularClaimerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// State mocks base method.
func (m *MockModelContext) State() *state.State {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "State")
	ret0, _ := ret[0].(*state.State)
	return ret0
}

// State indicates an expected call of State.
func (mr *MockModelContextMockRecorder) State() *MockModelContextStateCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockModelContext)(nil).State))
	return &MockModelContextStateCall{Call: call}
}

// MockModelContextStateCall wrap *gomock.Call
type MockModelContextStateCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextStateCall) Return(arg0 *state.State) *MockModelContextStateCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextStateCall) Do(f func() *state.State) *MockModelContextStateCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextStateCall) DoAndReturn(f func() *state.State) *MockModelContextStateCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// StatePool mocks base method.
func (m *MockModelContext) StatePool() *state.StatePool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StatePool")
	ret0, _ := ret[0].(*state.StatePool)
	return ret0
}

// StatePool indicates an expected call of StatePool.
func (mr *MockModelContextMockRecorder) StatePool() *MockModelContextStatePoolCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatePool", reflect.TypeOf((*MockModelContext)(nil).StatePool))
	return &MockModelContextStatePoolCall{Call: call}
}

// MockModelContextStatePoolCall wrap *gomock.Call
type MockModelContextStatePoolCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextStatePoolCall) Return(arg0 *state.StatePool) *MockModelContextStatePoolCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextStatePoolCall) Do(f func() *state.StatePool) *MockModelContextStatePoolCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextStatePoolCall) DoAndReturn(f func() *state.StatePool) *MockModelContextStatePoolCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WatcherRegistry mocks base method.
func (m *MockModelContext) WatcherRegistry() facade.WatcherRegistry {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatcherRegistry")
	ret0, _ := ret[0].(facade.WatcherRegistry)
	return ret0
}

// WatcherRegistry indicates an expected call of WatcherRegistry.
func (mr *MockModelContextMockRecorder) WatcherRegistry() *MockModelContextWatcherRegistryCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatcherRegistry", reflect.TypeOf((*MockModelContext)(nil).WatcherRegistry))
	return &MockModelContextWatcherRegistryCall{Call: call}
}

// MockModelContextWatcherRegistryCall wrap *gomock.Call
type MockModelContextWatcherRegistryCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelContextWatcherRegistryCall) Return(arg0 facade.WatcherRegistry) *MockModelContextWatcherRegistryCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelContextWatcherRegistryCall) Do(f func() facade.WatcherRegistry) *MockModelContextWatcherRegistryCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelContextWatcherRegistryCall) DoAndReturn(f func() facade.WatcherRegistry) *MockModelContextWatcherRegistryCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
