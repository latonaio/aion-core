// Code generated by MockGen. DO NOT EDIT.
// Source: ./config/yaml.go

// Package mock_pconfig is a generated GoMock package.
package mock_config

import (
	pconfig "bitbucket.org/latonaio/aion-core/config"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockServiceConfigContainer is a mock of ServiceConfigContainer interface
type MockServiceConfigContainer struct {
	ctrl     *gomock.Controller
	recorder *MockServiceConfigContainerMockRecorder
}

// MockServiceConfigContainerMockRecorder is the mock recorder for MockServiceConfigContainer
type MockServiceConfigContainerMockRecorder struct {
	mock *MockServiceConfigContainer
}

// NewMockServiceConfigContainer creates a new mock instance
func NewMockServiceConfigContainer(ctrl *gomock.Controller) *MockServiceConfigContainer {
	mock := &MockServiceConfigContainer{ctrl: ctrl}
	mock.recorder = &MockServiceConfigContainerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockServiceConfigContainer) EXPECT() *MockServiceConfigContainerMockRecorder {
	return m.recorder
}

// GetMicroserviceList mocks base method
func (m *MockServiceConfigContainer) GetMicroserviceList() map[string]*pconfig.Microservice {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMicroserviceList")
	ret0, _ := ret[0].(map[string]*pconfig.Microservice)
	return ret0
}

// GetMicroserviceList indicates an expected call of GetMicroserviceList
func (mr *MockServiceConfigContainerMockRecorder) GetMicroserviceList() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMicroserviceList", reflect.TypeOf((*MockServiceConfigContainer)(nil).GetMicroserviceList))
}

// GetMicroserviceByName mocks base method
func (m *MockServiceConfigContainer) GetMicroserviceByName(name string) (*pconfig.Microservice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMicroserviceByName", name)
	ret0, _ := ret[0].(*pconfig.Microservice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMicroserviceByName indicates an expected call of GetMicroserviceByName
func (mr *MockServiceConfigContainerMockRecorder) GetMicroserviceByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMicroserviceByName", reflect.TypeOf((*MockServiceConfigContainer)(nil).GetMicroserviceByName), name)
}

// GetNextServiceList mocks base method
func (m *MockServiceConfigContainer) GetNextServiceList(name, connectionKey string) ([]*pconfig.NextServiceSetting, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNextServiceList", name, connectionKey)
	ret0, _ := ret[0].([]*pconfig.NextServiceSetting)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNextServiceList indicates an expected call of GetNextServiceList
func (mr *MockServiceConfigContainerMockRecorder) GetNextServiceList(name, connectionKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNextServiceList", reflect.TypeOf((*MockServiceConfigContainer)(nil).GetNextServiceList), name, connectionKey)
}

// GetDeviceList mocks base method
func (m *MockServiceConfigContainer) GetDeviceList() map[string]*pconfig.Device {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceList")
	ret0, _ := ret[0].(map[string]*pconfig.Device)
	return ret0
}

// GetDeviceList indicates an expected call of GetDeviceList
func (mr *MockServiceConfigContainerMockRecorder) GetDeviceList() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceList", reflect.TypeOf((*MockServiceConfigContainer)(nil).GetDeviceList))
}

// LoadConfig mocks base method
func (m *MockServiceConfigContainer) LoadConfig(confPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadConfig", confPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadConfig indicates an expected call of LoadConfig
func (mr *MockServiceConfigContainerMockRecorder) LoadConfig(confPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadConfig", reflect.TypeOf((*MockServiceConfigContainer)(nil).LoadConfig), confPath)
}
