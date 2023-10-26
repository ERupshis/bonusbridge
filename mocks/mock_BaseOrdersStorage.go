// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erupshis/bonusbridge/internal/orders/storage (interfaces: BaseOrdersStorage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	data "github.com/erupshis/bonusbridge/internal/orders/data"
	gomock "github.com/golang/mock/gomock"
)

// MockBaseOrdersStorage is a mock of BaseOrdersStorage interface.
type MockBaseOrdersStorage struct {
	ctrl     *gomock.Controller
	recorder *MockBaseOrdersStorageMockRecorder
}

// MockBaseOrdersStorageMockRecorder is the mock recorder for MockBaseOrdersStorage.
type MockBaseOrdersStorageMockRecorder struct {
	mock *MockBaseOrdersStorage
}

// NewMockBaseOrdersStorage creates a new mock instance.
func NewMockBaseOrdersStorage(ctrl *gomock.Controller) *MockBaseOrdersStorage {
	mock := &MockBaseOrdersStorage{ctrl: ctrl}
	mock.recorder = &MockBaseOrdersStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBaseOrdersStorage) EXPECT() *MockBaseOrdersStorageMockRecorder {
	return m.recorder
}

// AddOrder mocks base method.
func (m *MockBaseOrdersStorage) AddOrder(arg0 context.Context, arg1 string, arg2 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOrder", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddOrder indicates an expected call of AddOrder.
func (mr *MockBaseOrdersStorageMockRecorder) AddOrder(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrder", reflect.TypeOf((*MockBaseOrdersStorage)(nil).AddOrder), arg0, arg1, arg2)
}

// GetOrders mocks base method.
func (m *MockBaseOrdersStorage) GetOrders(arg0 context.Context, arg1 map[string]interface{}) ([]data.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", arg0, arg1)
	ret0, _ := ret[0].([]data.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockBaseOrdersStorageMockRecorder) GetOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockBaseOrdersStorage)(nil).GetOrders), arg0, arg1)
}

// UpdateOrder mocks base method.
func (m *MockBaseOrdersStorage) UpdateOrder(arg0 context.Context, arg1 *data.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockBaseOrdersStorageMockRecorder) UpdateOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockBaseOrdersStorage)(nil).UpdateOrder), arg0, arg1)
}
