// Code generated by MockGen. DO NOT EDIT.
// Source: keptn-contrib/job-executor-service/pkg/config (interfaces: KeptnResourceService)

// Package fake is a generated GoMock package.
package fake

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockKeptnResourceService is a mock of KeptnResourceService interface.
type MockKeptnResourceService struct {
	ctrl     *gomock.Controller
	recorder *MockKeptnResourceServiceMockRecorder
}

// MockKeptnResourceServiceMockRecorder is the mock recorder for MockKeptnResourceService.
type MockKeptnResourceServiceMockRecorder struct {
	mock *MockKeptnResourceService
}

// NewMockKeptnResourceService creates a new mock instance.
func NewMockKeptnResourceService(ctrl *gomock.Controller) *MockKeptnResourceService {
	mock := &MockKeptnResourceService{ctrl: ctrl}
	mock.recorder = &MockKeptnResourceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeptnResourceService) EXPECT() *MockKeptnResourceServiceMockRecorder {
	return m.recorder
}

// GetAllKeptnResources mocks base method.
func (m *MockKeptnResourceService) GetAllKeptnResources(arg0 string) (map[string][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllKeptnResources", arg0)
	ret0, _ := ret[0].(map[string][]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllKeptnResources indicates an expected call of GetAllKeptnResources.
func (mr *MockKeptnResourceServiceMockRecorder) GetAllKeptnResources(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllKeptnResources", reflect.TypeOf((*MockKeptnResourceService)(nil).GetAllKeptnResources), arg0)
}

// GetProjectResource mocks base method.
func (m *MockKeptnResourceService) GetProjectResource(arg0, arg1 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProjectResource", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProjectResource indicates an expected call of GetProjectResource.
func (mr *MockKeptnResourceServiceMockRecorder) GetProjectResource(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProjectResource", reflect.TypeOf((*MockKeptnResourceService)(nil).GetProjectResource), arg0, arg1)
}

// GetServiceResource mocks base method.
func (m *MockKeptnResourceService) GetServiceResource(arg0, arg1 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceResource", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceResource indicates an expected call of GetServiceResource.
func (mr *MockKeptnResourceServiceMockRecorder) GetServiceResource(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceResource", reflect.TypeOf((*MockKeptnResourceService)(nil).GetServiceResource), arg0, arg1)
}

// GetStageResource mocks base method.
func (m *MockKeptnResourceService) GetStageResource(arg0, arg1 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStageResource", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStageResource indicates an expected call of GetStageResource.
func (mr *MockKeptnResourceServiceMockRecorder) GetStageResource(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStageResource", reflect.TypeOf((*MockKeptnResourceService)(nil).GetStageResource), arg0, arg1)
}
