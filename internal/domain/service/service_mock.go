package service

import (
	"testing"

	"github.com/diegoclair/slack-rotation-bot/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type allMocks struct {
	mockDataManager   *mocks.MockDataManager
	mockChannelRepo   *mocks.MockChannelRepo
	mockUserRepo      *mocks.MockUserRepo
	mockSchedulerRepo *mocks.MockSchedulerRepo
	mockSlackClient   *mocks.MockSlackClient
}

func newServiceTestMock(t *testing.T) (m allMocks, ctrl *gomock.Controller) {
	t.Helper()

	ctrl = gomock.NewController(t)

	dm := mocks.NewMockDataManager(ctrl)

	channelRepo := mocks.NewMockChannelRepo(ctrl)
	dm.EXPECT().Channel().Return(channelRepo).AnyTimes()

	userRepo := mocks.NewMockUserRepo(ctrl)
	dm.EXPECT().User().Return(userRepo).AnyTimes()

	schedulerRepo := mocks.NewMockSchedulerRepo(ctrl)
	dm.EXPECT().Scheduler().Return(schedulerRepo).AnyTimes()

	slackClient := mocks.NewMockSlackClient(ctrl)

	m = allMocks{
		mockDataManager:   dm,
		mockChannelRepo:   channelRepo,
		mockUserRepo:      userRepo,
		mockSchedulerRepo: schedulerRepo,
		mockSlackClient:   slackClient,
	}

	// validate service creation
	rotationService := newRotation(dm, slackClient)
	require.NotNil(t, rotationService)

	return
}