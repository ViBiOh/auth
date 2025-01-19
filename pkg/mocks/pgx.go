// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/jackc/pgx/v5 (interfaces: Row)
//
// Generated by this command:
//
//	mockgen -destination pkg/mocks/pgx.go -package mocks -mock_names Row=Row github.com/jackc/pgx/v5 Row
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// Row is a mock of Row interface.
type Row struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *RowMockRecorder
}

// RowMockRecorder is the mock recorder for Row.
type RowMockRecorder struct {
	mock *Row
}

// NewRow creates a new mock instance.
func NewRow(ctrl *gomock.Controller) *Row {
	mock := &Row{ctrl: ctrl}
	mock.recorder = &RowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Row) EXPECT() *RowMockRecorder {
	return m.recorder
}

// Scan mocks base method.
func (m *Row) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *RowMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*Row)(nil).Scan), dest...)
}
