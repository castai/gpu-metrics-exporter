// Code generated by mockery v2.42.0. DO NOT EDIT.

package exporter

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockExporter is an autogenerated mock type for the Exporter type
type MockExporter struct {
	mock.Mock
}

type MockExporter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockExporter) EXPECT() *MockExporter_Expecter {
	return &MockExporter_Expecter{mock: &_m.Mock}
}

// Disable provides a mock function with given fields:
func (_m *MockExporter) Disable() {
	_m.Called()
}

// MockExporter_Disable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disable'
type MockExporter_Disable_Call struct {
	*mock.Call
}

// Disable is a helper method to define mock.On call
func (_e *MockExporter_Expecter) Disable() *MockExporter_Disable_Call {
	return &MockExporter_Disable_Call{Call: _e.mock.On("Disable")}
}

func (_c *MockExporter_Disable_Call) Run(run func()) *MockExporter_Disable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExporter_Disable_Call) Return() *MockExporter_Disable_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockExporter_Disable_Call) RunAndReturn(run func()) *MockExporter_Disable_Call {
	_c.Call.Return(run)
	return _c
}

// Enable provides a mock function with given fields:
func (_m *MockExporter) Enable() {
	_m.Called()
}

// MockExporter_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type MockExporter_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
func (_e *MockExporter_Expecter) Enable() *MockExporter_Enable_Call {
	return &MockExporter_Enable_Call{Call: _e.mock.On("Enable")}
}

func (_c *MockExporter_Enable_Call) Run(run func()) *MockExporter_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExporter_Enable_Call) Return() *MockExporter_Enable_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockExporter_Enable_Call) RunAndReturn(run func()) *MockExporter_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// Enabled provides a mock function with given fields:
func (_m *MockExporter) Enabled() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Enabled")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockExporter_Enabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enabled'
type MockExporter_Enabled_Call struct {
	*mock.Call
}

// Enabled is a helper method to define mock.On call
func (_e *MockExporter_Expecter) Enabled() *MockExporter_Enabled_Call {
	return &MockExporter_Enabled_Call{Call: _e.mock.On("Enabled")}
}

func (_c *MockExporter_Enabled_Call) Run(run func()) *MockExporter_Enabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExporter_Enabled_Call) Return(_a0 bool) *MockExporter_Enabled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExporter_Enabled_Call) RunAndReturn(run func() bool) *MockExporter_Enabled_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: ctx
func (_m *MockExporter) Start(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockExporter_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockExporter_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockExporter_Expecter) Start(ctx interface{}) *MockExporter_Start_Call {
	return &MockExporter_Start_Call{Call: _e.mock.On("Start", ctx)}
}

func (_c *MockExporter_Start_Call) Run(run func(ctx context.Context)) *MockExporter_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockExporter_Start_Call) Return(_a0 error) *MockExporter_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExporter_Start_Call) RunAndReturn(run func(context.Context) error) *MockExporter_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockExporter creates a new instance of MockExporter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockExporter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockExporter {
	mock := &MockExporter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
