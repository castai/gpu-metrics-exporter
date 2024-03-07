// Code generated by mockery v2.42.0. DO NOT EDIT.

package exporter

import (
	exporter "github.com/castai/gpu-metrics-exporter/internal/exporter"
	mock "github.com/stretchr/testify/mock"

	pb "github.com/castai/gpu-metrics-exporter/pb"

	time "time"
)

// MockMetricMapper is an autogenerated mock type for the MetricMapper type
type MockMetricMapper struct {
	mock.Mock
}

type MockMetricMapper_Expecter struct {
	mock *mock.Mock
}

func (_m *MockMetricMapper) EXPECT() *MockMetricMapper_Expecter {
	return &MockMetricMapper_Expecter{mock: &_m.Mock}
}

// Map provides a mock function with given fields: metrics, ts
func (_m *MockMetricMapper) Map(metrics []exporter.MetricFamiliyMap, ts time.Time) *pb.MetricsBatch {
	ret := _m.Called(metrics, ts)

	if len(ret) == 0 {
		panic("no return value specified for Map")
	}

	var r0 *pb.MetricsBatch
	if rf, ok := ret.Get(0).(func([]exporter.MetricFamiliyMap, time.Time) *pb.MetricsBatch); ok {
		r0 = rf(metrics, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.MetricsBatch)
		}
	}

	return r0
}

// MockMetricMapper_Map_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Map'
type MockMetricMapper_Map_Call struct {
	*mock.Call
}

// Map is a helper method to define mock.On call
//   - metrics []exporter.MetricFamiliyMap
//   - ts time.Time
func (_e *MockMetricMapper_Expecter) Map(metrics interface{}, ts interface{}) *MockMetricMapper_Map_Call {
	return &MockMetricMapper_Map_Call{Call: _e.mock.On("Map", metrics, ts)}
}

func (_c *MockMetricMapper_Map_Call) Run(run func(metrics []exporter.MetricFamiliyMap, ts time.Time)) *MockMetricMapper_Map_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]exporter.MetricFamiliyMap), args[1].(time.Time))
	})
	return _c
}

func (_c *MockMetricMapper_Map_Call) Return(_a0 *pb.MetricsBatch) *MockMetricMapper_Map_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockMetricMapper_Map_Call) RunAndReturn(run func([]exporter.MetricFamiliyMap, time.Time) *pb.MetricsBatch) *MockMetricMapper_Map_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockMetricMapper creates a new instance of MockMetricMapper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMetricMapper(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMetricMapper {
	mock := &MockMetricMapper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
