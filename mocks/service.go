// Code generated by MockGen. DO NOT EDIT.
// Source: internal/domain/contract/service.go
//
// Generated by this command:
//
//	mockgen -package mocks -source=internal/domain/contract/service.go -destination=mocks/service.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entity "github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	gomock "go.uber.org/mock/gomock"
)

// MockRotationService is a mock of RotationService interface.
type MockRotationService struct {
	ctrl     *gomock.Controller
	recorder *MockRotationServiceMockRecorder
	isgomock struct{}
}

// MockRotationServiceMockRecorder is the mock recorder for MockRotationService.
type MockRotationServiceMockRecorder struct {
	mock *MockRotationService
}

// NewMockRotationService creates a new mock instance.
func NewMockRotationService(ctrl *gomock.Controller) *MockRotationService {
	mock := &MockRotationService{ctrl: ctrl}
	mock.recorder = &MockRotationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRotationService) EXPECT() *MockRotationServiceMockRecorder {
	return m.recorder
}

// AddUser mocks base method.
func (m *MockRotationService) AddUser(channelID int64, slackUserID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddUser", channelID, slackUserID)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddUser indicates an expected call of AddUser.
func (mr *MockRotationServiceMockRecorder) AddUser(channelID, slackUserID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUser", reflect.TypeOf((*MockRotationService)(nil).AddUser), channelID, slackUserID)
}

// GetChannelConfig mocks base method.
func (m *MockRotationService) GetChannelConfig(channelID int64) (*entity.Channel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannelConfig", channelID)
	ret0, _ := ret[0].(*entity.Channel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChannelConfig indicates an expected call of GetChannelConfig.
func (mr *MockRotationServiceMockRecorder) GetChannelConfig(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannelConfig", reflect.TypeOf((*MockRotationService)(nil).GetChannelConfig), channelID)
}

// GetChannelStatus mocks base method.
func (m *MockRotationService) GetChannelStatus(channelID int) (*entity.Channel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChannelStatus", channelID)
	ret0, _ := ret[0].(*entity.Channel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChannelStatus indicates an expected call of GetChannelStatus.
func (mr *MockRotationServiceMockRecorder) GetChannelStatus(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChannelStatus", reflect.TypeOf((*MockRotationService)(nil).GetChannelStatus), channelID)
}

// GetCurrentPresenter mocks base method.
func (m *MockRotationService) GetCurrentPresenter(channelID int64) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentPresenter", channelID)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentPresenter indicates an expected call of GetCurrentPresenter.
func (mr *MockRotationServiceMockRecorder) GetCurrentPresenter(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentPresenter", reflect.TypeOf((*MockRotationService)(nil).GetCurrentPresenter), channelID)
}

// GetNextPresenter mocks base method.
func (m *MockRotationService) GetNextPresenter(channelID int64) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNextPresenter", channelID)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNextPresenter indicates an expected call of GetNextPresenter.
func (mr *MockRotationServiceMockRecorder) GetNextPresenter(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNextPresenter", reflect.TypeOf((*MockRotationService)(nil).GetNextPresenter), channelID)
}

// GetSchedulerConfig mocks base method.
func (m *MockRotationService) GetSchedulerConfig(channelID int64) (*entity.Scheduler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSchedulerConfig", channelID)
	ret0, _ := ret[0].(*entity.Scheduler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSchedulerConfig indicates an expected call of GetSchedulerConfig.
func (mr *MockRotationServiceMockRecorder) GetSchedulerConfig(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSchedulerConfig", reflect.TypeOf((*MockRotationService)(nil).GetSchedulerConfig), channelID)
}

// ListUsers mocks base method.
func (m *MockRotationService) ListUsers(channelID int64) ([]*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", channelID)
	ret0, _ := ret[0].([]*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockRotationServiceMockRecorder) ListUsers(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockRotationService)(nil).ListUsers), channelID)
}

// PauseScheduler mocks base method.
func (m *MockRotationService) PauseScheduler(channelID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PauseScheduler", channelID)
	ret0, _ := ret[0].(error)
	return ret0
}

// PauseScheduler indicates an expected call of PauseScheduler.
func (mr *MockRotationServiceMockRecorder) PauseScheduler(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PauseScheduler", reflect.TypeOf((*MockRotationService)(nil).PauseScheduler), channelID)
}

// RecordPresentation mocks base method.
func (m *MockRotationService) RecordPresentation(ctx context.Context, channelID, userID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordPresentation", ctx, channelID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordPresentation indicates an expected call of RecordPresentation.
func (mr *MockRotationServiceMockRecorder) RecordPresentation(ctx, channelID, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPresentation", reflect.TypeOf((*MockRotationService)(nil).RecordPresentation), ctx, channelID, userID)
}

// RemoveUser mocks base method.
func (m *MockRotationService) RemoveUser(channelID int64, slackUserID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveUser", channelID, slackUserID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveUser indicates an expected call of RemoveUser.
func (mr *MockRotationServiceMockRecorder) RemoveUser(channelID, slackUserID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUser", reflect.TypeOf((*MockRotationService)(nil).RemoveUser), channelID, slackUserID)
}

// ResumeScheduler mocks base method.
func (m *MockRotationService) ResumeScheduler(channelID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResumeScheduler", channelID)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResumeScheduler indicates an expected call of ResumeScheduler.
func (mr *MockRotationServiceMockRecorder) ResumeScheduler(channelID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResumeScheduler", reflect.TypeOf((*MockRotationService)(nil).ResumeScheduler), channelID)
}

// SetupChannel mocks base method.
func (m *MockRotationService) SetupChannel(slackChannelID, channelName, teamID string) (*entity.Channel, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupChannel", slackChannelID, channelName, teamID)
	ret0, _ := ret[0].(*entity.Channel)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SetupChannel indicates an expected call of SetupChannel.
func (mr *MockRotationServiceMockRecorder) SetupChannel(slackChannelID, channelName, teamID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupChannel", reflect.TypeOf((*MockRotationService)(nil).SetupChannel), slackChannelID, channelName, teamID)
}

// UpdateChannelConfig mocks base method.
func (m *MockRotationService) UpdateChannelConfig(channelID int64, configType, configValue string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateChannelConfig", channelID, configType, configValue)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateChannelConfig indicates an expected call of UpdateChannelConfig.
func (mr *MockRotationServiceMockRecorder) UpdateChannelConfig(channelID, configType, configValue any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateChannelConfig", reflect.TypeOf((*MockRotationService)(nil).UpdateChannelConfig), channelID, configType, configValue)
}
