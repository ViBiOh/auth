// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/auth/v2/pkg/auth (interfaces: Storage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/auth/v2/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// Storage is a mock of Storage interface.
type Storage struct {
	ctrl     *gomock.Controller
	recorder *StorageMockRecorder
}

// StorageMockRecorder is the mock recorder for Storage.
type StorageMockRecorder struct {
	mock *Storage
}

// NewStorage creates a new mock instance.
func NewStorage(ctrl *gomock.Controller) *Storage {
	mock := &Storage{ctrl: ctrl}
	mock.recorder = &StorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Storage) EXPECT() *StorageMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *Storage) Create(arg0 context.Context, arg1 model.User) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *StorageMockRecorder) Create(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*Storage)(nil).Create), arg0, arg1)
}

// Delete mocks base method.
func (m *Storage) Delete(arg0 context.Context, arg1 model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *StorageMockRecorder) Delete(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*Storage)(nil).Delete), arg0, arg1)
}

// DoAtomic mocks base method.
func (m *Storage) DoAtomic(arg0 context.Context, arg1 func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoAtomic", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DoAtomic indicates an expected call of DoAtomic.
func (mr *StorageMockRecorder) DoAtomic(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoAtomic", reflect.TypeOf((*Storage)(nil).DoAtomic), arg0, arg1)
}

// Get mocks base method.
func (m *Storage) Get(arg0 context.Context, arg1 uint64) (model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *StorageMockRecorder) Get(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*Storage)(nil).Get), arg0, arg1)
}

// Update mocks base method.
func (m *Storage) Update(arg0 context.Context, arg1 model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *StorageMockRecorder) Update(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*Storage)(nil).Update), arg0, arg1)
}
