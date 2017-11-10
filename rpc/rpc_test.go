// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package rpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/juju/loggo"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/rpc"
	"github.com/juju/juju/rpc/jsoncodec"
	"github.com/juju/juju/rpc/rpcreflect"
	"github.com/juju/juju/testing"
)

var logger = loggo.GetLogger("juju.rpc")

type rpcSuite struct {
	testing.BaseSuite
}

var _ = gc.Suite(&rpcSuite{})

type callInfo struct {
	rcvr   interface{}
	method string
	arg    interface{}
}

type callError callInfo

func (e *callError) Error() string {
	return fmt.Sprintf("error calling %s", e.method)
}

type stringVal struct {
	Val string
}

type Root struct {
	mu          sync.Mutex
	conn        *rpc.Conn
	calls       []*callInfo
	returnErr   bool
	simple      map[string]*SimpleMethods
	delayed     map[string]*DelayedMethods
	errorInst   *ErrorMethods
	contextInst *ContextMethods
}

func (r *Root) callError(rcvr interface{}, name string, arg interface{}) error {
	if r.returnErr {
		return &callError{rcvr, name, arg}
	}
	return nil
}

func (r *Root) SimpleMethods(id string) (*SimpleMethods, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a := r.simple[id]; a != nil {
		return a, nil
	}
	return nil, fmt.Errorf("unknown SimpleMethods id")
}

func (r *Root) DelayedMethods(id string) (*DelayedMethods, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a := r.delayed[id]; a != nil {
		return a, nil
	}
	return nil, fmt.Errorf("unknown DelayedMethods id")
}

func (r *Root) ErrorMethods(id string) (*ErrorMethods, error) {
	if r.errorInst == nil {
		return nil, fmt.Errorf("no error methods")
	}
	return r.errorInst, nil
}

func (r *Root) ContextMethods(id string) (*ContextMethods, error) {
	if r.contextInst == nil {
		return nil, fmt.Errorf("no context methods")
	}
	return r.contextInst, nil
}

func (r *Root) Discard1() {}

func (r *Root) Discard2(id string) error { return nil }

func (r *Root) Discard3(id string) int { return 0 }

func (r *Root) CallbackMethods(string) (*CallbackMethods, error) {
	return &CallbackMethods{r}, nil
}

func (r *Root) InterfaceMethods(id string) (InterfaceMethods, error) {
	logger.Infof("interface methods called")
	m, err := r.SimpleMethods(id)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type InterfaceMethods interface {
	Call1r1e(s stringVal) (stringVal, error)
}

type ChangeAPIMethods struct {
	r *Root
}

func (r *Root) ChangeAPIMethods(string) (*ChangeAPIMethods, error) {
	return &ChangeAPIMethods{r}, nil
}

func (t *Root) called(rcvr interface{}, method string, arg interface{}) {
	t.mu.Lock()
	t.calls = append(t.calls, &callInfo{rcvr, method, arg})
	t.mu.Unlock()
}

type SimpleMethods struct {
	root *Root
	id   string
}

// Each Call method is named in this standard form:
//
//     Call<narg>r<nret><e>
//
// where narg is the number of arguments, nret is the number of returned
// values (not including the error) and e is the letter 'e' if the
// method returns an error.

func (a *SimpleMethods) Call0r0() {
	a.root.called(a, "Call0r0", nil)
}

func (a *SimpleMethods) Call0r1() stringVal {
	a.root.called(a, "Call0r1", nil)
	return stringVal{"Call0r1 ret"}
}

func (a *SimpleMethods) Call0r1e() (stringVal, error) {
	a.root.called(a, "Call0r1e", nil)
	return stringVal{"Call0r1e ret"}, a.root.callError(a, "Call0r1e", nil)
}

func (a *SimpleMethods) Call0r0e() error {
	a.root.called(a, "Call0r0e", nil)
	return a.root.callError(a, "Call0r0e", nil)
}

func (a *SimpleMethods) Call1r0(s stringVal) {
	a.root.called(a, "Call1r0", s)
}

func (a *SimpleMethods) Call1r1(s stringVal) stringVal {
	a.root.called(a, "Call1r1", s)
	return stringVal{"Call1r1 ret"}
}

func (a *SimpleMethods) Call1r1e(s stringVal) (stringVal, error) {
	a.root.called(a, "Call1r1e", s)
	return stringVal{"Call1r1e ret"}, a.root.callError(a, "Call1r1e", s)
}

func (a *SimpleMethods) Call1r0e(s stringVal) error {
	a.root.called(a, "Call1r0e", s)
	return a.root.callError(a, "Call1r0e", s)
}

func (a *SimpleMethods) SliceArg(struct{ X []string }) stringVal {
	return stringVal{"SliceArg ret"}
}

func (a *SimpleMethods) Discard1(int) {}

func (a *SimpleMethods) Discard2(struct{}, struct{}) {}

func (a *SimpleMethods) Discard3() int { return 0 }

func (a *SimpleMethods) Discard4() (_, _ struct{}) { return }

type ContextMethods struct {
	root        *Root
	callContext context.Context
	waiting     chan struct{}
}

func (c *ContextMethods) Call0(ctx context.Context) error {
	c.root.called(c, "Call0", nil)
	c.callContext = ctx
	return c.checkContext(ctx)
}

func (c *ContextMethods) Call1(ctx context.Context, s stringVal) error {
	c.root.called(c, "Call1", s)
	c.callContext = ctx
	return c.checkContext(ctx)
}

func (c *ContextMethods) Wait(ctx context.Context) error {
	c.root.called(c, "Wait", nil)
	close(c.waiting)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(testing.LongWait):
		return errors.New("expected context to be cancelled")
	}
}

func (c *ContextMethods) checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(testing.ShortWait):
	}
	return nil
}

type DelayedMethods struct {
	ready     chan struct{}
	done      chan string
	doneError chan error
}

func (a *DelayedMethods) Delay() (stringVal, error) {
	if a.ready != nil {
		a.ready <- struct{}{}
	}
	select {
	case s := <-a.done:
		return stringVal{s}, nil
	case err := <-a.doneError:
		return stringVal{}, err
	}
}

type ErrorMethods struct {
	err error
}

func (e *ErrorMethods) Call() error {
	return e.err
}

type CallbackMethods struct {
	root *Root
}

type int64val struct {
	I int64
}

func (a *CallbackMethods) Factorial(x int64val) (int64val, error) {
	if x.I <= 1 {
		return int64val{1}, nil
	}
	var r int64val
	err := a.root.conn.Call(rpc.Request{"CallbackMethods", 0, "", "Factorial"}, int64val{x.I - 1}, &r)
	if err != nil {
		return int64val{}, err
	}
	return int64val{x.I * r.I}, nil
}

func (a *ChangeAPIMethods) ChangeAPI() {
	a.r.conn.Serve(&changedAPIRoot{}, nil)
}

func (a *ChangeAPIMethods) RemoveAPI() {
	a.r.conn.Serve(nil, nil)
}

type changedAPIRoot struct{}

func (r *changedAPIRoot) NewlyAvailable(string) (newlyAvailableMethods, error) {
	return newlyAvailableMethods{}, nil
}

type newlyAvailableMethods struct{}

func (newlyAvailableMethods) NewMethod() stringVal {
	return stringVal{"new method result"}
}

type VariableMethods1 struct {
	sm *SimpleMethods
}

func (vm *VariableMethods1) Call0r1() stringVal {
	return vm.sm.Call0r1()
}

type VariableMethods2 struct {
	sm *SimpleMethods
}

func (vm *VariableMethods2) Call1r1(s stringVal) stringVal {
	return vm.sm.Call1r1(s)
}

type RestrictedMethods struct {
	InterfaceMethods
}

type CustomRoot struct {
	root *Root
}

type wrapper func(*SimpleMethods) reflect.Value

type customMethodCaller struct {
	wrap         wrapper
	root         *Root
	objMethod    rpcreflect.ObjMethod
	expectedType reflect.Type
}

func (c customMethodCaller) ParamsType() reflect.Type {
	return c.objMethod.Params
}

func (c customMethodCaller) ResultType() reflect.Type {
	return c.objMethod.Result
}

func (c customMethodCaller) Call(_ context.Context, objId string, arg reflect.Value) (reflect.Value, error) {
	sm, err := c.root.SimpleMethods(objId)
	if err != nil {
		return reflect.Value{}, err
	}
	obj := c.wrap(sm)
	if reflect.TypeOf(obj) != c.expectedType {
		logger.Errorf("got the wrong type back, expected %s got %T", c.expectedType, obj)
	}
	logger.Debugf("calling: %T %v %#v", obj, obj, c.objMethod)
	return c.objMethod.Call(context.TODO(), obj, arg)
}

func (cc *CustomRoot) Kill() {
}

func (cc *CustomRoot) FindMethod(
	rootMethodName string, version int, objMethodName string,
) (
	rpcreflect.MethodCaller, error,
) {
	logger.Debugf("got to FindMethod: %q %d %q", rootMethodName, version, objMethodName)
	if rootMethodName != "MultiVersion" {
		return nil, &rpcreflect.CallNotImplementedError{
			RootMethod: rootMethodName,
		}
	}
	var goType reflect.Type
	var wrap wrapper
	switch version {
	case 0:
		goType = reflect.TypeOf((*VariableMethods1)(nil))
		wrap = func(sm *SimpleMethods) reflect.Value {
			return reflect.ValueOf(&VariableMethods1{sm})
		}
	case 1:
		goType = reflect.TypeOf((*VariableMethods2)(nil))
		wrap = func(sm *SimpleMethods) reflect.Value {
			return reflect.ValueOf(&VariableMethods2{sm})
		}
	case 2:
		goType = reflect.TypeOf((*RestrictedMethods)(nil))
		wrap = func(sm *SimpleMethods) reflect.Value {
			methods := &RestrictedMethods{InterfaceMethods: sm}
			return reflect.ValueOf(methods)
		}
	default:
		return nil, &rpcreflect.CallNotImplementedError{
			RootMethod: rootMethodName,
			Version:    version,
		}
	}
	logger.Debugf("found type: %s", goType)
	objType := rpcreflect.ObjTypeOf(goType)
	objMethod, err := objType.Method(objMethodName)
	if err != nil {
		return nil, &rpcreflect.CallNotImplementedError{
			RootMethod: rootMethodName,
			Version:    version,
			Method:     objMethodName,
		}
	}
	return customMethodCaller{
		objMethod:    objMethod,
		root:         cc.root,
		wrap:         wrap,
		expectedType: goType,
	}, nil
}

func SimpleRoot() *Root {
	root := &Root{
		simple: make(map[string]*SimpleMethods),
	}
	root.simple["a99"] = &SimpleMethods{root: root, id: "a99"}
	return root
}

func (*rpcSuite) TestRPC(c *gc.C) {
	root := SimpleRoot()
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	for narg := 0; narg < 2; narg++ {
		for nret := 0; nret < 2; nret++ {
			for nerr := 0; nerr < 2; nerr++ {
				retErr := nerr != 0
				p := testCallParams{
					client:         client,
					serverNotifier: serverNotifier,
					entry:          "SimpleMethods",
					narg:           narg,
					nret:           nret,
					retErr:         retErr,
					testErr:        false,
				}
				root.testCall(c, p)
				if retErr {
					p.testErr = true
					root.testCall(c, p)
				}
			}
		}
	}
}

func callName(narg, nret int, retErr bool) string {
	e := ""
	if retErr {
		e = "e"
	}
	return fmt.Sprintf("Call%dr%d%s", narg, nret, e)
}

type testCallParams struct {
	// client holds the client-side of the rpc connection that
	// will be used to make the call.
	client *rpc.Conn

	// serverNotifier holds the notifier for the server side.
	serverNotifier *notifier

	// entry holds the top-level type that will be invoked
	// (e.g. "SimpleMethods").
	entry string

	// narg holds the number of arguments accepted by the
	// call (0 or 1).
	narg int

	// nret holds the number of values returned by the
	// call (0 or 1).
	nret int

	// retErr specifies whether the call returns an error.
	retErr bool

	// testErr specifies whether the call should be made to return an error.
	testErr bool

	// version specifies what version of the interface to call, defaults to 0.
	version int
}

// request returns the RPC request for the test call.
func (p testCallParams) request() rpc.Request {
	return rpc.Request{
		Type:    p.entry,
		Version: p.version,
		Id:      "a99",
		Action:  callName(p.narg, p.nret, p.retErr),
	}
}

// error message returns the error message that the test call
// should return if it returns an error.
func (p testCallParams) errorMessage() string {
	return fmt.Sprintf("error calling %s", p.request().Action)
}

func (root *Root) testCall(c *gc.C, args testCallParams) {
	args.serverNotifier.reset()
	root.calls = nil
	root.returnErr = args.testErr
	c.Logf("test call %s", args.request().Action)
	var response stringVal
	err := args.client.Call(args.request(), stringVal{"arg"}, &response)
	switch {
	case args.retErr && args.testErr:
		c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
			Message: args.errorMessage(),
		})
		c.Assert(response, gc.Equals, stringVal{})
	case args.nret > 0:
		c.Check(response, gc.Equals, stringVal{args.request().Action + " ret"})
	}
	if !args.testErr {
		c.Check(err, jc.ErrorIsNil)
	}

	// Check that the call was actually made, the right
	// parameters were received and the right result returned.
	root.mu.Lock()
	defer root.mu.Unlock()

	root.assertCallMade(c, args)
	root.assertServerNotified(c, args, args.client.ClientRequestID())
}

func (root *Root) assertCallMade(c *gc.C, p testCallParams) {
	expectCall := callInfo{
		rcvr:   root.simple["a99"],
		method: p.request().Action,
	}
	if p.narg > 0 {
		expectCall.arg = stringVal{"arg"}
	}
	c.Assert(root.calls, gc.HasLen, 1)
	c.Assert(*root.calls[0], gc.Equals, expectCall)
}

// assertServerNotified asserts that the right server notifications
// were made for the given test call parameters. The id of the request
// is held in requestId.
func (root *Root) assertServerNotified(c *gc.C, p testCallParams, requestId uint64) {
	// Test that there was a notification for the request.
	c.Assert(p.serverNotifier.serverRequests, gc.HasLen, 1)
	serverReq := p.serverNotifier.serverRequests[0]
	c.Assert(serverReq.hdr, gc.DeepEquals, rpc.Header{
		RequestId: requestId,
		Request:   p.request(),
		Version:   1,
	})
	if p.narg > 0 {
		c.Assert(serverReq.body, gc.Equals, stringVal{"arg"})
	} else {
		c.Assert(serverReq.body, gc.Equals, struct{}{})
	}

	// Test that there was a notification for the reply.
	c.Assert(p.serverNotifier.serverReplies, gc.HasLen, 1)
	serverReply := p.serverNotifier.serverReplies[0]
	c.Assert(serverReply.req, gc.Equals, p.request())
	if p.retErr && p.testErr || p.nret == 0 {
		c.Assert(serverReply.body, gc.Equals, struct{}{})
	} else {
		c.Assert(serverReply.body, gc.Equals, stringVal{p.request().Action + " ret"})
	}
	if p.retErr && p.testErr {
		c.Assert(serverReply.hdr, gc.Equals, rpc.Header{
			RequestId: requestId,
			Error:     p.errorMessage(),
			Version:   1,
		})
	} else {
		c.Assert(serverReply.hdr, gc.Equals, rpc.Header{
			RequestId: requestId,
			Version:   1,
		})
	}
}

func (*rpcSuite) TestInterfaceMethods(c *gc.C) {
	root := SimpleRoot()
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	p := testCallParams{
		client:         client,
		serverNotifier: serverNotifier,
		entry:          "InterfaceMethods",
		narg:           1,
		nret:           1,
		retErr:         true,
		testErr:        false,
	}

	root.testCall(c, p)
	p.testErr = true
	root.testCall(c, p)
	// Call0r0 is defined on the underlying SimpleMethods, but is not
	// exposed at the InterfaceMethods level, so this call should fail with
	// CodeNotImplemented.
	var r stringVal
	err := client.Call(rpc.Request{"InterfaceMethods", 0, "a99", "Call0r0"}, stringVal{"arg"}, &r)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "no such request - method InterfaceMethods.Call0r0 is not implemented",
		Code:    rpc.CodeNotImplemented,
	})
}

func (*rpcSuite) TestCustomRootV0(c *gc.C) {
	root := &CustomRoot{SimpleRoot()}
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	// V0 of MultiVersion implements only VariableMethods1.Call0r1.
	p := testCallParams{
		client:         client,
		serverNotifier: serverNotifier,
		entry:          "MultiVersion",
		version:        0,
		narg:           0,
		nret:           1,
		retErr:         false,
		testErr:        false,
	}

	root.root.testCall(c, p)
	// Call1r1 is exposed in version 1, but not in version 0.
	var r stringVal
	err := client.Call(rpc.Request{"MultiVersion", 0, "a99", "Call1r1"}, stringVal{"arg"}, &r)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "no such request - method MultiVersion.Call1r1 is not implemented",
		Code:    rpc.CodeNotImplemented,
	})
}

func (*rpcSuite) TestCustomRootV1(c *gc.C) {
	root := &CustomRoot{SimpleRoot()}
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	// V1 of MultiVersion implements only VariableMethods2.Call1r1.
	p := testCallParams{
		client:         client,
		serverNotifier: serverNotifier,
		entry:          "MultiVersion",
		version:        1,
		narg:           1,
		nret:           1,
		retErr:         false,
		testErr:        false,
	}

	root.root.testCall(c, p)
	// Call0r1 is exposed in version 0, but not in version 1.
	var r stringVal
	err := client.Call(rpc.Request{"MultiVersion", 1, "a99", "Call0r1"}, nil, &r)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "no such request - method MultiVersion(1).Call0r1 is not implemented",
		Code:    rpc.CodeNotImplemented,
	})
}

func (*rpcSuite) TestCustomRootV2(c *gc.C) {
	root := &CustomRoot{SimpleRoot()}
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	p := testCallParams{
		client:         client,
		serverNotifier: serverNotifier,
		entry:          "MultiVersion",
		version:        2,
		narg:           1,
		nret:           1,
		retErr:         true,
		testErr:        false,
	}

	root.root.testCall(c, p)
	// By embedding the InterfaceMethods inside a concrete
	// RestrictedMethods type, we actually only expose the methods defined
	// in InterfaceMethods.
	var r stringVal
	err := client.Call(rpc.Request{"MultiVersion", 2, "a99", "Call0r1e"}, nil, &r)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: `no such request - method MultiVersion(2).Call0r1e is not implemented`,
		Code:    rpc.CodeNotImplemented,
	})
}

func (*rpcSuite) TestCustomRootUnknownVersion(c *gc.C) {
	root := &CustomRoot{SimpleRoot()}
	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	var r stringVal
	// Unknown version 5
	err := client.Call(rpc.Request{"MultiVersion", 5, "a99", "Call0r1"}, nil, &r)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: `unknown version (5) of interface "MultiVersion"`,
		Code:    rpc.CodeNotImplemented,
	})
}

func (*rpcSuite) TestConcurrentCalls(c *gc.C) {
	start1 := make(chan string)
	start2 := make(chan string)
	ready1 := make(chan struct{})
	ready2 := make(chan struct{})

	root := &Root{
		delayed: map[string]*DelayedMethods{
			"1": {ready: ready1, done: start1},
			"2": {ready: ready2, done: start2},
		},
	}

	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	call := func(id string, done chan<- struct{}) {
		var r stringVal
		err := client.Call(rpc.Request{"DelayedMethods", 0, id, "Delay"}, nil, &r)
		c.Check(err, jc.ErrorIsNil)
		c.Check(r.Val, gc.Equals, "return "+id)
		done <- struct{}{}
	}
	done1 := make(chan struct{})
	done2 := make(chan struct{})
	go call("1", done1)
	go call("2", done2)

	// Check that both calls are running concurrently.
	chanRead(c, ready1, "method 1 ready")
	chanRead(c, ready2, "method 2 ready")

	// Let the requests complete.
	start1 <- "return 1"
	start2 <- "return 2"
	chanRead(c, done1, "method 1 done")
	chanRead(c, done2, "method 2 done")
}

type codedError struct {
	m    string
	code string
}

func (e *codedError) Error() string {
	return e.m
}

func (e *codedError) ErrorCode() string {
	return e.code
}

func (*rpcSuite) TestErrorCode(c *gc.C) {
	root := &Root{
		errorInst: &ErrorMethods{&codedError{"message", "code"}},
	}
	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	err := client.Call(rpc.Request{"ErrorMethods", 0, "", "Call"}, nil, nil)
	c.Assert(err, gc.ErrorMatches, `message \(code\)`)
	c.Assert(errors.Cause(err).(rpc.ErrorCoder).ErrorCode(), gc.Equals, "code")
}

func (*rpcSuite) TestTransformErrors(c *gc.C) {
	root := &Root{
		errorInst: &ErrorMethods{&codedError{"message", "code"}},
	}
	tfErr := func(err error) error {
		c.Check(err, gc.NotNil)
		if e, ok := err.(*codedError); ok {
			return &codedError{
				m:    "transformed: " + e.m,
				code: "transformed: " + e.code,
			}
		}
		return fmt.Errorf("transformed: %v", err)
	}
	client, _, srvDone, _ := newRPCClientServer(c, root, tfErr, false)
	defer closeClient(c, client, srvDone)
	// First, we don't transform methods we can't find.
	err := client.Call(rpc.Request{"foo", 0, "", "bar"}, nil, nil)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: `unknown object type "foo"`,
		Code:    rpc.CodeNotImplemented,
	})

	err = client.Call(rpc.Request{"ErrorMethods", 0, "", "NoMethod"}, nil, nil)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "no such request - method ErrorMethods.NoMethod is not implemented",
		Code:    rpc.CodeNotImplemented,
	})

	// We do transform any errors that happen from calling the RootMethod
	// and beyond.
	err = client.Call(rpc.Request{"ErrorMethods", 0, "", "Call"}, nil, nil)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "transformed: message",
		Code:    "transformed: code",
	})

	root.errorInst.err = nil
	err = client.Call(rpc.Request{"ErrorMethods", 0, "", "Call"}, nil, nil)
	c.Assert(err, jc.ErrorIsNil)

	root.errorInst = nil
	err = client.Call(rpc.Request{"ErrorMethods", 0, "", "Call"}, nil, nil)
	c.Assert(errors.Cause(err), gc.DeepEquals, &rpc.RequestError{
		Message: "transformed: no error methods",
	})

}

func (*rpcSuite) TestServerWaitsForOutstandingCalls(c *gc.C) {
	ready := make(chan struct{})
	start := make(chan string)
	root := &Root{
		delayed: map[string]*DelayedMethods{
			"1": {
				ready: ready,
				done:  start,
			},
		},
	}
	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	done := make(chan struct{})
	go func() {
		var r stringVal
		err := client.Call(rpc.Request{"DelayedMethods", 0, "1", "Delay"}, nil, &r)
		c.Check(errors.Cause(err), gc.Equals, rpc.ErrShutdown)
		done <- struct{}{}
	}()
	chanRead(c, ready, "DelayedMethods.Delay ready")
	client.Close()
	select {
	case err := <-srvDone:
		c.Fatalf("server returned while outstanding operation in progress: %v", err)
		<-done
	case <-time.After(25 * time.Millisecond):
	}
	start <- "xxx"
}

func chanRead(c *gc.C, ch <-chan struct{}, what string) {
	select {
	case <-ch:
		return
	case <-time.After(3 * time.Second):
		c.Fatalf("timeout on channel read %s", what)
	}
}

func (*rpcSuite) TestCompatibility(c *gc.C) {
	root := &Root{
		simple: make(map[string]*SimpleMethods),
	}
	a0 := &SimpleMethods{root: root, id: "a0"}
	root.simple["a0"] = a0

	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)
	call := func(method string, arg, ret interface{}) (passedArg interface{}) {
		root.calls = nil
		err := client.Call(rpc.Request{"SimpleMethods", 0, "a0", method}, arg, ret)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(root.calls, gc.HasLen, 1)
		info := root.calls[0]
		c.Assert(info.rcvr, gc.Equals, a0)
		c.Assert(info.method, gc.Equals, method)
		return info.arg
	}
	type extra struct {
		Val   string
		Extra string
	}
	// Extra fields in request and response.
	var r extra
	arg := call("Call1r1", extra{"x", "y"}, &r)
	c.Assert(arg, gc.Equals, stringVal{"x"})

	// Nil argument as request.
	r = extra{}
	arg = call("Call1r1", nil, &r)
	c.Assert(arg, gc.Equals, stringVal{})

	// Nil argument as response.
	arg = call("Call1r1", stringVal{"x"}, nil)
	c.Assert(arg, gc.Equals, stringVal{"x"})

	// Non-nil argument for no response.
	r = extra{}
	arg = call("Call1r0", stringVal{"x"}, &r)
	c.Assert(arg, gc.Equals, stringVal{"x"})
	c.Assert(r, gc.Equals, extra{})
}

func (*rpcSuite) TestBadCall(c *gc.C) {
	loggo.GetLogger("juju.rpc").SetLogLevel(loggo.TRACE)
	root := &Root{
		simple: make(map[string]*SimpleMethods),
	}
	a0 := &SimpleMethods{root: root, id: "a0"}
	root.simple["a0"] = a0
	client, _, srvDone, serverNotifier := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)

	testBadCall(c, client, serverNotifier,
		rpc.Request{"BadSomething", 0, "a0", "No"},
		`unknown object type "BadSomething"`,
		rpc.CodeNotImplemented,
		false,
	)
	testBadCall(c, client, serverNotifier,
		rpc.Request{"SimpleMethods", 0, "xx", "No"},
		"no such request - method SimpleMethods.No is not implemented",
		rpc.CodeNotImplemented,
		false,
	)
	testBadCall(c, client, serverNotifier,
		rpc.Request{"SimpleMethods", 0, "xx", "Call0r0"},
		`unknown SimpleMethods id`,
		"",
		true,
	)
}

func testBadCall(
	c *gc.C,
	client *rpc.Conn,
	serverNotifier *notifier,
	req rpc.Request,
	expectedErr string,
	expectedErrCode string,
	requestKnown bool,
) {
	serverNotifier.reset()
	err := client.Call(req, nil, nil)
	msg := expectedErr
	if expectedErrCode != "" {
		msg += " (" + expectedErrCode + ")"
	}
	c.Assert(err, gc.ErrorMatches, regexp.QuoteMeta(msg))

	// From docs on ServerRequest:
	// 	If the request was not recognized or there was
	//	an error reading the body, body will be nil.
	var expectBody interface{}
	if requestKnown {
		expectBody = struct{}{}
	}
	c.Assert(serverNotifier.serverRequests[0], gc.DeepEquals, requestEvent{
		hdr: rpc.Header{
			RequestId: client.ClientRequestID(),
			Request:   req,
			Version:   1,
		},
		body: expectBody,
	})

	// Test that there was a notification for the server reply.
	c.Assert(serverNotifier.serverReplies, gc.HasLen, 1)
	serverReply := serverNotifier.serverReplies[0]
	c.Assert(serverReply, gc.DeepEquals, replyEvent{
		hdr: rpc.Header{
			RequestId: client.ClientRequestID(),
			Error:     expectedErr,
			ErrorCode: expectedErrCode,
			Version:   1,
		},
		req:  req,
		body: struct{}{},
	})
}

func (*rpcSuite) TestContinueAfterReadBodyError(c *gc.C) {
	root := &Root{
		simple: make(map[string]*SimpleMethods),
	}
	a0 := &SimpleMethods{root: root, id: "a0"}
	root.simple["a0"] = a0
	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)

	var ret stringVal
	arg0 := struct {
		X map[string]int
	}{
		X: map[string]int{"hello": 65},
	}
	err := client.Call(rpc.Request{"SimpleMethods", 0, "a0", "SliceArg"}, arg0, &ret)
	c.Assert(err, gc.ErrorMatches, `json: cannot unmarshal object into Go (?:value)|(?:struct field \.X) of type \[\]string`)

	err = client.Call(rpc.Request{"SimpleMethods", 0, "a0", "SliceArg"}, arg0, &ret)
	c.Assert(err, gc.ErrorMatches, `json: cannot unmarshal object into Go (?:value)|(?:struct field \.X) of type \[\]string`)

	arg1 := struct {
		X []string
	}{
		X: []string{"one"},
	}
	err = client.Call(rpc.Request{"SimpleMethods", 0, "a0", "SliceArg"}, arg1, &ret)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(ret.Val, gc.Equals, "SliceArg ret")
}

func (*rpcSuite) TestErrorAfterClientClose(c *gc.C) {
	client, _, srvDone, _ := newRPCClientServer(c, &Root{}, nil, false)
	err := client.Close()
	c.Assert(err, jc.ErrorIsNil)
	err = client.Call(rpc.Request{"Foo", 0, "", "Bar"}, nil, nil)
	c.Assert(errors.Cause(err), gc.Equals, rpc.ErrShutdown)
	err = chanReadError(c, srvDone, "server done")
	c.Assert(err, jc.ErrorIsNil)
}

func (*rpcSuite) TestClientCloseIdempotent(c *gc.C) {
	client, _, _, _ := newRPCClientServer(c, &Root{}, nil, false)
	err := client.Close()
	c.Assert(err, jc.ErrorIsNil)
	err = client.Close()
	c.Assert(err, jc.ErrorIsNil)
	err = client.Close()
	c.Assert(err, jc.ErrorIsNil)
}

func (*rpcSuite) TestBidirectional(c *gc.C) {
	srvRoot := &Root{}
	client, _, srvDone, _ := newRPCClientServer(c, srvRoot, nil, true)
	defer closeClient(c, client, srvDone)
	clientRoot := &Root{conn: client}
	client.Serve(clientRoot, nil)
	var r int64val
	err := client.Call(rpc.Request{"CallbackMethods", 0, "", "Factorial"}, int64val{12}, &r)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(r.I, gc.Equals, int64(479001600))
}

func (*rpcSuite) TestServerRequestWhenNotServing(c *gc.C) {
	srvRoot := &Root{}
	client, _, srvDone, _ := newRPCClientServer(c, srvRoot, nil, true)
	defer closeClient(c, client, srvDone)
	var r int64val
	err := client.Call(rpc.Request{"CallbackMethods", 0, "", "Factorial"}, int64val{12}, &r)
	c.Assert(err, gc.ErrorMatches, "no service")
}

func (*rpcSuite) TestChangeAPI(c *gc.C) {
	srvRoot := &Root{}
	client, _, srvDone, _ := newRPCClientServer(c, srvRoot, nil, true)
	defer closeClient(c, client, srvDone)
	var s stringVal
	err := client.Call(rpc.Request{"NewlyAvailable", 0, "", "NewMethod"}, nil, &s)
	c.Assert(err, gc.ErrorMatches, `unknown object type "NewlyAvailable" \(not implemented\)`)
	err = client.Call(rpc.Request{"ChangeAPIMethods", 0, "", "ChangeAPI"}, nil, nil)
	c.Assert(err, jc.ErrorIsNil)
	err = client.Call(rpc.Request{"ChangeAPIMethods", 0, "", "ChangeAPI"}, nil, nil)
	c.Assert(err, gc.ErrorMatches, `unknown object type "ChangeAPIMethods" \(not implemented\)`)
	err = client.Call(rpc.Request{"NewlyAvailable", 0, "", "NewMethod"}, nil, &s)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(s, gc.Equals, stringVal{"new method result"})
}

func (*rpcSuite) TestChangeAPIToNil(c *gc.C) {
	srvRoot := &Root{}
	client, _, srvDone, _ := newRPCClientServer(c, srvRoot, nil, true)
	defer closeClient(c, client, srvDone)

	err := client.Call(rpc.Request{"ChangeAPIMethods", 0, "", "RemoveAPI"}, nil, nil)
	c.Assert(err, jc.ErrorIsNil)

	err = client.Call(rpc.Request{"ChangeAPIMethods", 0, "", "RemoveAPI"}, nil, nil)
	c.Assert(err, gc.ErrorMatches, "no service")
}

func (*rpcSuite) TestChangeAPIWhileServingRequest(c *gc.C) {
	ready := make(chan struct{})
	done := make(chan error)
	srvRoot := &Root{
		delayed: map[string]*DelayedMethods{
			"1": {ready: ready, doneError: done},
		},
	}
	transform := func(err error) error {
		return fmt.Errorf("transformed: %v", err)
	}
	client, _, srvDone, _ := newRPCClientServer(c, srvRoot, transform, true)
	defer closeClient(c, client, srvDone)

	result := make(chan error)
	go func() {
		result <- client.Call(rpc.Request{"DelayedMethods", 0, "1", "Delay"}, nil, nil)
	}()
	chanRead(c, ready, "method ready")

	err := client.Call(rpc.Request{"ChangeAPIMethods", 0, "", "ChangeAPI"}, nil, nil)
	c.Assert(err, jc.ErrorIsNil)

	// Ensure that not only does the request in progress complete,
	// but that the original transformErrors function is called.
	done <- fmt.Errorf("an error")
	select {
	case r := <-result:
		c.Assert(r, gc.ErrorMatches, "transformed: an error")
	case <-time.After(3 * time.Second):
		c.Fatalf("timeout on channel read")
	}
}

func (*rpcSuite) TestCodeNotImplementedMatchesAPIserverParams(c *gc.C) {
	c.Assert(rpc.CodeNotImplemented, gc.Equals, params.CodeNotImplemented)
}

func (*rpcSuite) TestRequestContext(c *gc.C) {
	root := &Root{}
	root.contextInst = &ContextMethods{root: root}

	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)

	call := func(method string, arg, ret interface{}) (passedArg interface{}) {
		root.calls = nil
		root.contextInst.callContext = nil
		err := client.Call(rpc.Request{"ContextMethods", 0, "", method}, arg, ret)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(root.calls, gc.HasLen, 1)
		info := root.calls[0]
		c.Assert(info.rcvr, gc.Equals, root.contextInst)
		c.Assert(info.method, gc.Equals, method)
		c.Assert(root.contextInst.callContext, gc.NotNil)
		// context is cancelled when the method returns
		c.Assert(root.contextInst.callContext.Err(), gc.Equals, context.Canceled)
		return info.arg
	}

	arg := call("Call0", nil, nil)
	c.Assert(arg, gc.IsNil)

	arg = call("Call1", stringVal{"foo"}, nil)
	c.Assert(arg, gc.Equals, stringVal{"foo"})
}

func (*rpcSuite) TestConnectionContextCloseClient(c *gc.C) {
	root := &Root{}
	root.contextInst = &ContextMethods{
		root:    root,
		waiting: make(chan struct{}),
	}

	client, _, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)

	errch := make(chan error, 1)
	go func() {
		errch <- client.Call(rpc.Request{"ContextMethods", 0, "", "Wait"}, nil, nil)
	}()

	<-root.contextInst.waiting
	err := client.Close()
	c.Assert(err, jc.ErrorIsNil)

	err = <-errch
	c.Assert(err, jc.Satisfies, rpc.IsShutdownErr)
}

func (*rpcSuite) TestConnectionContextCloseServer(c *gc.C) {
	root := &Root{}
	root.contextInst = &ContextMethods{
		root:    root,
		waiting: make(chan struct{}),
	}

	client, server, srvDone, _ := newRPCClientServer(c, root, nil, false)
	defer closeClient(c, client, srvDone)

	errch := make(chan error, 1)
	go func() {
		errch <- client.Call(rpc.Request{"ContextMethods", 0, "", "Wait"}, nil, nil)
	}()

	<-root.contextInst.waiting
	err := server.Close()
	c.Assert(err, jc.ErrorIsNil)

	err = <-errch
	c.Assert(err, gc.ErrorMatches, "context canceled")
}

func chanReadError(c *gc.C, ch <-chan error, what string) error {
	select {
	case e := <-ch:
		return e
	case <-time.After(3 * time.Second):
		c.Fatalf("timeout on channel read %s", what)
	}
	panic("unreachable")
}

// newRPCClientServer starts an RPC server serving a connection from a
// single client.  When the server has finished serving the connection,
// it sends a value on the returned channel.
// If bidir is true, requests can flow in both directions.
func newRPCClientServer(
	c *gc.C,
	root interface{},
	tfErr func(error) error,
	bidir bool,
) (client *rpc.Conn, server *rpc.Conn, srvDone chan error, serverNotifier *notifier) {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, jc.ErrorIsNil)

	srvDone = make(chan error, 1)
	serverNotifier = new(notifier)
	srvStarted := make(chan *rpc.Conn)
	go func() {
		defer close(srvDone)
		defer close(srvStarted)
		defer l.Close()

		conn, err := l.Accept()
		if err != nil {
			srvDone <- err
			return
		}

		role := roleServer
		if bidir {
			role = roleBoth
		}
		rpcConn := rpc.NewConn(NewJSONCodec(conn, role), serverNotifier)
		if custroot, ok := root.(*CustomRoot); ok {
			rpcConn.ServeRoot(custroot, tfErr)
			custroot.root.conn = rpcConn
		} else {
			rpcConn.Serve(root, tfErr)
		}
		if root, ok := root.(*Root); ok {
			root.conn = rpcConn
		}
		rpcConn.Start(context.Background())
		srvStarted <- rpcConn
		<-rpcConn.Dead()
		srvDone <- rpcConn.Close()
	}()
	conn, err := net.Dial("tcp", l.Addr().String())
	c.Assert(err, jc.ErrorIsNil)
	server = <-srvStarted
	if server == nil {
		conn.Close()
		c.Fatal(<-srvDone)
	}
	role := roleClient
	if bidir {
		role = roleBoth
	}
	client = rpc.NewConn(NewJSONCodec(conn, role), &notifier{})
	client.Start(context.Background())
	return client, server, srvDone, serverNotifier
}

func closeClient(c *gc.C, client *rpc.Conn, srvDone <-chan error) {
	err := client.Close()
	c.Assert(err, jc.ErrorIsNil)
	err = chanReadError(c, srvDone, "server done")
	c.Assert(err, jc.ErrorIsNil)
}

type encoder interface {
	Encode(e interface{}) error
}

type decoder interface {
	Decode(e interface{}) error
}

// testCodec wraps an rpc.Codec with extra error checking code.
type testCodec struct {
	role connRole
	rpc.Codec
}

func (c *testCodec) WriteMessage(hdr *rpc.Header, x interface{}) error {
	if reflect.ValueOf(x).Kind() != reflect.Struct {
		panic(fmt.Errorf("WriteRequest bad param; want struct got %T (%#v)", x, x))
	}
	if c.role != roleBoth && hdr.IsRequest() != (c.role == roleClient) {
		panic(fmt.Errorf("codec role %v; header wrong type %#v", c.role, hdr))
	}
	logger.Infof("send header: %#v; body: %#v", hdr, x)
	return c.Codec.WriteMessage(hdr, x)
}

func (c *testCodec) ReadHeader(hdr *rpc.Header) error {
	err := c.Codec.ReadHeader(hdr)
	if err != nil {
		return err
	}
	logger.Infof("got header %#v", hdr)
	if c.role != roleBoth && hdr.IsRequest() == (c.role == roleClient) {
		panic(fmt.Errorf("codec role %v; read wrong type %#v", c.role, hdr))
	}
	return nil
}

func (c *testCodec) ReadBody(r interface{}, isRequest bool) error {
	if v := reflect.ValueOf(r); v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("ReadResponseBody bad destination; want *struct got %T", r))
	}
	if c.role != roleBoth && isRequest == (c.role == roleClient) {
		panic(fmt.Errorf("codec role %v; read wrong body type %#v", c.role, r))
	}
	// Note: this will need to change if we want to test a non-JSON codec.
	var m json.RawMessage
	err := c.Codec.ReadBody(&m, isRequest)
	if err != nil {
		return err
	}
	logger.Infof("got response body: %q", m)
	err = json.Unmarshal(m, r)
	logger.Infof("unmarshalled into %#v", r)
	return err
}

type connRole string

const (
	roleBoth   connRole = "both"
	roleClient connRole = "client"
	roleServer connRole = "server"
)

func NewJSONCodec(c net.Conn, role connRole) rpc.Codec {
	return &testCodec{
		role:  role,
		Codec: jsoncodec.NewNet(c),
	}
}

type requestEvent struct {
	hdr  rpc.Header
	body interface{}
}

type replyEvent struct {
	req  rpc.Request
	hdr  rpc.Header
	body interface{}
}

type notifier struct {
	mu             sync.Mutex
	serverRequests []requestEvent
	serverReplies  []replyEvent
}

func (n *notifier) RPCObserver() rpc.Observer {
	// For testing, we usually won't want an actual copy of the
	// stub. To avoid confusing test failures (e.g. wondering why your
	// calls aren't showing up on your stub because the underlying
	// code has called DeepCopy) and immense complexity, just return
	// the same value.
	return n
}

func (n *notifier) reset() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.serverRequests = nil
	n.serverReplies = nil
}

func (n *notifier) ServerRequest(hdr *rpc.Header, body interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.serverRequests = append(n.serverRequests, requestEvent{
		hdr:  *hdr,
		body: body,
	})
}

func (n *notifier) ServerReply(req rpc.Request, hdr *rpc.Header, body interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.serverReplies = append(n.serverReplies, replyEvent{
		req:  req,
		hdr:  *hdr,
		body: body,
	})
}
