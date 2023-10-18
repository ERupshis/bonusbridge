// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erupshis/bonusbridge/internal/orders/storage/managers (interfaces: BaseOrdersManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	data "github.com/erupshis/bonusbridge/internal/orders/data"
	gomock "github.com/golang/mock/gomock"
)

// MockBaseOrdersManager is a mock of BaseOrdersManager interface.
type MockBaseOrdersManager struct {
	ctrl     *gomock.Controller
	recorder *MockBaseOrdersManagerMockRecorder
}

// MockBaseOrdersManagerMockRecorder is the mock recorder for MockBaseOrdersManager.
type MockBaseOrdersManagerMockRecorder struct {
	mock *MockBaseOrdersManager
}

// NewMockBaseOrdersManager creates a new mock instance.
func NewMockBaseOrdersManager(ctrl *gomock.Controller) *MockBaseOrdersManager {
	mock := &MockBaseOrdersManager{ctrl: ctrl}
	mock.recorder = &MockBaseOrdersManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBaseOrdersManager) EXPECT() *MockBaseOrdersManagerMockRecorder {
	return m.recorder
}

// AddOrder mocks base method.
func (m *MockBaseOrdersManager) AddOrder(arg0 context.Context, arg1 string, arg2 int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOrder", arg0, arg1, arg2)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddOrder indicates an expected call of AddOrder.
func (mr *MockBaseOrdersManagerMockRecorder) AddOrder(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrder", reflect.TypeOf((*MockBaseOrdersManager)(nil).AddOrder), arg0, arg1, arg2)
}

// GetOrders mocks base method.
func (m *MockBaseOrdersManager) GetOrders(arg0 context.Context, arg1 map[string]interface{}) ([]data.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", arg0, arg1)
	ret0, _ := ret[0].([]data.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockBaseOrdersManagerMockRecorder) GetOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockBaseOrdersManager)(nil).GetOrders), arg0, arg1)
}

// UpdateOrder mocks base method.
func (m *MockBaseOrdersManager) UpdateOrder(arg0 context.Context, arg1 *data.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockBaseOrdersManagerMockRecorder) UpdateOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockBaseOrdersManager)(nil).UpdateOrder), arg0, arg1)
}