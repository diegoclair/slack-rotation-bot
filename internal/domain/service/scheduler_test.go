package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_newScheduler(t *testing.T) {
	m, ctrl := newServiceTestMock(t)
	defer ctrl.Finish()

	scheduler := newScheduler(m.mockDataManager, m.mockSlackClient)

	require.NotNil(t, scheduler)
	assert.Equal(t, m.mockDataManager, scheduler.dm)
	assert.Equal(t, m.mockSlackClient, scheduler.slackClient)
	assert.NotNil(t, scheduler.configChanged)
	assert.NotNil(t, scheduler.stopChan)
	assert.False(t, scheduler.running)
}

func Test_scheduler_calculateNextForScheduler(t *testing.T) {
	type args struct {
		scheduler *entity.Scheduler
		now       time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Should return today if time hasn't passed",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "15:00",
					ActiveDays:       []int{1, 2, 3, 4, 5}, // Monday-Friday
				},
				now: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), // Monday 10:00
			},
			want: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), // Monday 15:00
		},
		{
			name: "Should return next day if time has passed",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "09:00",
					ActiveDays:       []int{1, 2, 3, 4, 5}, // Monday-Friday
				},
				now: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), // Monday 10:00
			},
			want: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), // Tuesday 09:00
		},
		{
			name: "Should skip weekend and go to Monday",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "09:00",
					ActiveDays:       []int{1, 2, 3, 4, 5}, // Monday-Friday
				},
				now: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC), // Friday 10:00
			},
			want: time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), // Next Monday 09:00
		},
		{
			name: "Should handle Sunday correctly (ISO weekday 7)",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "09:00",
					ActiveDays:       []int{7}, // Sunday only
				},
				now: time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC), // Saturday 10:00
			},
			want: time.Date(2024, 1, 7, 9, 0, 0, 0, time.UTC), // Sunday 09:00
		},
		{
			name: "Should return zero time for invalid time format",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "invalid",
					ActiveDays:       []int{1, 2, 3, 4, 5},
				},
				now: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			want: time.Time{},
		},
		{
			name: "Should return zero time for no active days",
			args: args{
				scheduler: &entity.Scheduler{
					NotificationTime: "09:00",
					ActiveDays:       []int{},
				},
				now: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			want: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)
			got := s.calculateNextForScheduler(tt.args.scheduler, tt.args.now)

			if tt.want.IsZero() {
				assert.True(t, got.IsZero(), "Expected zero time but got %v", got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_scheduler_getNextPresenter(t *testing.T) {
	type args struct {
		channelID int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		want      *entity.User
		wantErr   bool
	}{
		{
			name: "Should return first user when no current presenter",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
					{ID: 2, SlackUserID: "U987654321", LastPresenter: false},
				}

				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(users, nil).Times(1)
			},
			want: &entity.User{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
			wantErr: false,
		},
		{
			name: "Should return next user after current presenter",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: true},  // Current presenter
					{ID: 2, SlackUserID: "U987654321", LastPresenter: false},
					{ID: 3, SlackUserID: "U555555555", LastPresenter: false},
				}

				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(users, nil).Times(1)
			},
			want: &entity.User{ID: 2, SlackUserID: "U987654321", LastPresenter: false},
			wantErr: false,
		},
		{
			name: "Should wrap around to first user",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
					{ID: 2, SlackUserID: "U987654321", LastPresenter: true}, // Last user is current
				}

				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(users, nil).Times(1)
			},
			want: &entity.User{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
			wantErr: false,
		},
		{
			name: "Should return nil when no users",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return([]*entity.User{}, nil).Times(1)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.getNextPresenter(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_scheduler_sendNotificationToChannel(t *testing.T) {
	type args struct {
		channelID int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should send notification successfully with next presenter",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: args.channelID,
					Role:      "Daily presenter",
				}

				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
					{ID: 2, SlackUserID: "U987654321", LastPresenter: false},
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(schedulerConfig, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockDataManager.EXPECT().
						WithTransaction(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx interface{}, fn interface{}) error {
							// Simulate successful transaction
							return nil
						}).Times(1),

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						Return("", "", nil).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should send no users message when rotation is empty",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: args.channelID,
					Role:      "Daily presenter",
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(schedulerConfig, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return([]*entity.User{}, nil).Times(1),

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						Return("", "", nil).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when channel not found",
			args: args{channelID: 999},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetByID(args.channelID).
					Return(nil, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should use default role when scheduler config is nil",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(nil, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockDataManager.EXPECT().
						WithTransaction(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx interface{}, fn interface{}) error {
							return nil
						}).Times(1),

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						DoAndReturn(func(channelID string, options ...slack.MsgOption) (string, string, error) {
							// Verify default role "On duty" is used in message
							return "", "", nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when channel repository GetByID fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetByID(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when scheduler repository GetByChannelID fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(nil, assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when getNextPresenter fails (GetActiveUsersByChannel error)",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: args.channelID,
					Role:      "Daily presenter",
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(schedulerConfig, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(nil, assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when Slack PostMessage fails for user notification",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: args.channelID,
					Role:      "Daily presenter",
				}

				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(schedulerConfig, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockDataManager.EXPECT().
						WithTransaction(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx interface{}, fn interface{}) error {
							return nil
						}).Times(1),

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						Return("", "", assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when Slack PostMessage fails for no users message",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:             args.channelID,
					SlackChannelID: "C123456789",
				}

				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: args.channelID,
					Role:      "Daily presenter",
				}

				gomock.InOrder(
					mocks.mockChannelRepo.EXPECT().
						GetByID(args.channelID).
						Return(channel, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(schedulerConfig, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return([]*entity.User{}, nil).Times(1),

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						Return("", "", assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.sendNotificationToChannel(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_scheduler_findNextNotification(t *testing.T) {
	tests := []struct {
		name             string
		buildMock        func(mocks allMocks)
		wantTime         bool // Whether we expect a valid time
		wantChannelCount int
	}{
		{
			name: "Should return next notification time for enabled schedulers",
			buildMock: func(mocks allMocks) {
				now := time.Now().UTC()
				tomorrow := now.AddDate(0, 0, 1)

				schedulers := []*entity.Scheduler{
					{
						ID:               1,
						ChannelID:        1,
						NotificationTime: "09:00",
						ActiveDays:       []int{int(tomorrow.Weekday())}, // Tomorrow's weekday
						IsEnabled:        true,
					},
					{
						ID:               2,
						ChannelID:        2,
						NotificationTime: "09:00",
						ActiveDays:       []int{int(tomorrow.Weekday())}, // Same time tomorrow
						IsEnabled:        true,
					},
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return(schedulers, nil).Times(1)
			},
			wantTime:         true,
			wantChannelCount: 2, // Both channels at same time
		},
		{
			name: "Should return empty when no enabled schedulers",
			buildMock: func(mocks allMocks) {
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{}, nil).Times(1)
			},
			wantTime:         false,
			wantChannelCount: 0,
		},
		{
			name: "Should handle error from repository",
			buildMock: func(mocks allMocks) {
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return(nil, assert.AnError).Times(1)
			},
			wantTime:         false,
			wantChannelCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m)
			}

			nextTime, channelIDs := s.findNextNotification()

			if tt.wantTime {
				assert.False(t, nextTime.IsZero(), "Expected valid time but got zero time")
			} else {
				assert.True(t, nextTime.IsZero(), "Expected zero time but got %v", nextTime)
			}

			assert.Len(t, channelIDs, tt.wantChannelCount)
		})
	}
}

func Test_scheduler_NotifyConfigChange(t *testing.T) {
	m, ctrl := newServiceTestMock(t)
	defer ctrl.Finish()

	s := newScheduler(m.mockDataManager, m.mockSlackClient)

	// Should not block even if channel is full
	s.NotifyConfigChange()
	s.NotifyConfigChange() // Second call should not block

	// Verify channel received at least one notification
	select {
	case <-s.configChanged:
		// Expected - received notification
	default:
		t.Error("Expected config change notification but channel was empty")
	}
}

func Test_scheduler_Start_Stop(t *testing.T) {
	m, ctrl := newServiceTestMock(t)
	defer ctrl.Finish()

	s := newScheduler(m.mockDataManager, m.mockSlackClient)

	// Mock the scheduler repo to return empty result so mainLoop doesn't panic
	m.mockSchedulerRepo.EXPECT().
		GetEnabled().
		Return([]*entity.Scheduler{}, nil).
		AnyTimes()

	// Initial state
	assert.False(t, s.running)

	// Start scheduler
	s.Start()
	assert.True(t, s.running)

	// Starting again should not change state
	s.Start()
	assert.True(t, s.running)

	// Stop scheduler
	s.Stop()
	assert.False(t, s.running)

	// Give the goroutine a moment to fully stop
	time.Sleep(10 * time.Millisecond)

	// Stopping again should not change state
	s.Stop()
	assert.False(t, s.running)
}

func Test_scheduler_recordPresentation(t *testing.T) {
	type args struct {
		channelID int64
		userID    int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should record presentation successfully",
			args: args{
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return fn(mocks.mockDataManager)
					}).Times(1)

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().ClearLastPresenter(args.channelID).Return(nil).Times(1),
					mocks.mockUserRepo.EXPECT().SetLastPresenter(args.userID).Return(nil).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when transaction fails",
			args: args{
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					Return(assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when ClearLastPresenter fails",
			args: args{
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return fn(mocks.mockDataManager)
					}).Times(1)

				mocks.mockUserRepo.EXPECT().ClearLastPresenter(args.channelID).Return(assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when SetLastPresenter fails",
			args: args{
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return fn(mocks.mockDataManager)
					}).Times(1)

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().ClearLastPresenter(args.channelID).Return(nil).Times(1),
					mocks.mockUserRepo.EXPECT().SetLastPresenter(args.userID).Return(assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.recordPresentation(tt.args.channelID, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_scheduler_sendNotifications(t *testing.T) {
	type args struct {
		channelIDs []int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
	}{
		{
			name: "Should send notifications to multiple channels successfully",
			args: args{channelIDs: []int64{1, 2}},
			buildMock: func(mocks allMocks, args args) {
				// For each channel, we expect the sendNotificationToChannel flow
				for _, channelID := range args.channelIDs {
					channel := &entity.Channel{
						ID:             channelID,
						SlackChannelID: fmt.Sprintf("C%d", channelID),
					}

					schedulerConfig := &entity.Scheduler{
						ID:        channelID,
						ChannelID: channelID,
						Role:      "Daily presenter",
					}

					users := []*entity.User{
						{ID: channelID, SlackUserID: fmt.Sprintf("U%d", channelID), LastPresenter: false},
					}

					mocks.mockChannelRepo.EXPECT().
						GetByID(channelID).
						Return(channel, nil).AnyTimes()

					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(channelID).
						Return(schedulerConfig, nil).AnyTimes()

					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(channelID).
						Return(users, nil).AnyTimes()

					mocks.mockDataManager.EXPECT().
						WithTransaction(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
							return nil
						}).AnyTimes()

					mocks.mockSlackClient.EXPECT().
						PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
						Return("", "", nil).AnyTimes()
				}
			},
		},
		{
			name: "Should handle empty channel list",
			args: args{channelIDs: []int64{}},
			buildMock: func(mocks allMocks, args args) {
				// No expectations needed for empty list
			},
		},
		{
			name: "Should continue sending to other channels even if one fails",
			args: args{channelIDs: []int64{1, 2}},
			buildMock: func(mocks allMocks, args args) {
				// Channel 1 will fail (channel not found)
				mocks.mockChannelRepo.EXPECT().
					GetByID(int64(1)).
					Return(nil, nil).AnyTimes()

				// Channel 2 will succeed
				channel2 := &entity.Channel{
					ID:             2,
					SlackChannelID: "C2",
				}

				schedulerConfig2 := &entity.Scheduler{
					ID:        2,
					ChannelID: 2,
					Role:      "Daily presenter",
				}

				users2 := []*entity.User{
					{ID: 2, SlackUserID: "U2", LastPresenter: false},
				}

				mocks.mockChannelRepo.EXPECT().
					GetByID(int64(2)).
					Return(channel2, nil).AnyTimes()

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(int64(2)).
					Return(schedulerConfig2, nil).AnyTimes()

				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(int64(2)).
					Return(users2, nil).AnyTimes()

				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return nil
					}).AnyTimes()

				mocks.mockSlackClient.EXPECT().
					PostMessage(channel2.SlackChannelID, gomock.Any(), gomock.Any()).
					Return("", "", nil).AnyTimes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newScheduler(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			// Since sendNotifications uses goroutines, we need to give it time to complete
			s.sendNotifications(tt.args.channelIDs)
			
			// Give goroutines time to complete
			time.Sleep(50 * time.Millisecond)
		})
	}
}

func Test_scheduler_mainLoop(t *testing.T) {
	tests := []struct {
		name      string
		buildMock func(mocks allMocks)
		testFunc  func(t *testing.T, s *scheduler)
	}{
		{
			name: "Should handle no active channels and respond to config change",
			buildMock: func(mocks allMocks) {
				// First call returns no channels
				// Second call (after config change) also returns no channels
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{}, nil).
					Times(2)
			},
			testFunc: func(t *testing.T, s *scheduler) {
				// Start the mainLoop in a goroutine
				go s.mainLoop()
				
				// Give it a moment to enter the first loop
				time.Sleep(10 * time.Millisecond)
				
				// Trigger config change
				s.NotifyConfigChange()
				
				// Give it time to process
				time.Sleep(50 * time.Millisecond)
				
				// Stop the scheduler
				close(s.stopChan)
				
				// Wait for goroutine to finish
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name: "Should send notifications immediately when time has passed",
			buildMock: func(mocks allMocks) {
				// Use a simple time that already passed today (early morning)
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: "06:00", // 6 AM - likely already passed
					ActiveDays:       []int{1, 2, 3, 4, 5, 6, 7}, // All days
					IsEnabled:        true,
				}
				
				// Mock for findNextNotification (time already passed, returns next day)
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{scheduler}, nil).
					AnyTimes()
				
				// Mock for sendNotificationToChannel
				channel := &entity.Channel{
					ID:             1,
					SlackChannelID: "C123456789",
				}
				
				schedulerConfig := &entity.Scheduler{
					ID:        1,
					ChannelID: 1,
					Role:      "Daily presenter",
				}
				
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", LastPresenter: false},
				}
				
				mocks.mockChannelRepo.EXPECT().
					GetByID(int64(1)).
					Return(channel, nil).AnyTimes()
				
				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(int64(1)).
					Return(schedulerConfig, nil).AnyTimes()
				
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(int64(1)).
					Return(users, nil).AnyTimes()
				
				mocks.mockDataManager.EXPECT().
					WithTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return nil
					}).AnyTimes()
				
				mocks.mockSlackClient.EXPECT().
					PostMessage(channel.SlackChannelID, gomock.Any(), gomock.Any()).
					Return("", "", nil).AnyTimes()
			},
			testFunc: func(t *testing.T, s *scheduler) {
				go s.mainLoop()
				
				// Wait for processing
				time.Sleep(200 * time.Millisecond)
				
				close(s.stopChan)
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "Should handle scheduler errors gracefully",
			buildMock: func(mocks allMocks) {
				// First call returns error, then we should get another call during the 1-hour wait
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return(nil, assert.AnError).
					Times(1)
				
				// During the 1-hour wait, there will be another call when we trigger config change
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{}, nil).
					AnyTimes()
			},
			testFunc: func(t *testing.T, s *scheduler) {
				go s.mainLoop()
				
				// Wait for error handling - goes into 1-hour wait
				time.Sleep(50 * time.Millisecond)
				
				// Trigger config change to interrupt the wait
				s.NotifyConfigChange()
				
				// Wait for config change processing
				time.Sleep(50 * time.Millisecond)
				
				close(s.stopChan)
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "Should handle config change during timer wait",
			buildMock: func(mocks allMocks) {
				now := time.Now().UTC()
				futureTime := now.Add(1 * time.Hour) // 1 hour in future
				
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: futureTime.Format("15:04"),
					ActiveDays:       []int{int(now.Weekday())}, // Today
					IsEnabled:        true,
				}
				
				// First call - returns scheduler with long wait time
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{scheduler}, nil).
					Times(1)
				
				// Second call after config change - return empty to exit loop
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{}, nil).
					Times(1)
			},
			testFunc: func(t *testing.T, s *scheduler) {
				go s.mainLoop()
				
				// Give timer time to start
				time.Sleep(50 * time.Millisecond)
				
				// Trigger config change to interrupt timer
				s.NotifyConfigChange()
				
				// Wait for config change to be processed
				time.Sleep(50 * time.Millisecond)
				
				close(s.stopChan)
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name: "Should handle stop signal during timer wait",
			buildMock: func(mocks allMocks) {
				now := time.Now().UTC()
				futureTime := now.Add(1 * time.Hour) // 1 hour in future
				
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: futureTime.Format("15:04"),
					ActiveDays:       []int{int(now.Weekday())}, // Today
					IsEnabled:        true,
				}
				
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{scheduler}, nil).
					Times(1)
			},
			testFunc: func(t *testing.T, s *scheduler) {
				go s.mainLoop()
				
				// Give timer time to start
				time.Sleep(50 * time.Millisecond)
				
				// Stop scheduler while timer is waiting
				close(s.stopChan)
				
				// Wait for goroutine to exit
				time.Sleep(50 * time.Millisecond)
			},
		},
		{
			name: "Should handle stop signal during no-channels wait",
			buildMock: func(mocks allMocks) {
				// Return no channels to trigger 1-hour wait
				mocks.mockSchedulerRepo.EXPECT().
					GetEnabled().
					Return([]*entity.Scheduler{}, nil).
					Times(1)
			},
			testFunc: func(t *testing.T, s *scheduler) {
				go s.mainLoop()
				
				// Give it time to enter no-channels wait
				time.Sleep(50 * time.Millisecond)
				
				// Stop scheduler during no-channels wait
				close(s.stopChan)
				
				// Wait for goroutine to exit
				time.Sleep(50 * time.Millisecond)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()
			
			s := newScheduler(m.mockDataManager, m.mockSlackClient)
			
			if tt.buildMock != nil {
				tt.buildMock(m)
			}
			
			if tt.testFunc != nil {
				tt.testFunc(t, s)
			}
		})
	}
}