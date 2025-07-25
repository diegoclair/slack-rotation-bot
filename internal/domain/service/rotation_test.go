package service

import (
	"context"
	"testing"

	"github.com/diegoclair/slack-rotation-bot/internal/domain"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_rotationService_SetupChannel(t *testing.T) {
	type args struct {
		slackChannelID   string
		slackChannelName string
		slackTeamID      string
	}
	tests := []struct {
		name         string
		buildMock    func(mocks allMocks, args args)
		args         args
		wantChannel  *entity.Channel
		wantCreated  bool
		wantErr      bool
	}{
		{
			name: "Should create new channel successfully",
			args: args{
				slackChannelID:   "C123456789",
				slackChannelName: "test-channel",
				slackTeamID:      "T123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				// Channel doesn't exist
				mocks.mockChannelRepo.EXPECT().
					GetBySlackID(args.slackChannelID).
					Return(nil, nil).Times(1)

				// Create channel
				mocks.mockChannelRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(channel *entity.Channel) error {
						channel.ID = 1
						require.Equal(t, args.slackChannelID, channel.SlackChannelID)
						require.Equal(t, args.slackChannelName, channel.SlackChannelName)
						require.Equal(t, args.slackTeamID, channel.SlackTeamID)
						require.True(t, channel.IsActive)
						return nil
					}).Times(1)

				// Create scheduler config
				mocks.mockSchedulerRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(scheduler *entity.Scheduler) error {
						scheduler.ID = 1
						require.Equal(t, int64(1), scheduler.ChannelID)
						require.Equal(t, "09:00", scheduler.NotificationTime)
						require.Equal(t, domain.DefaultActiveDays, scheduler.ActiveDays)
						require.True(t, scheduler.IsEnabled)
						require.Equal(t, "On duty", scheduler.Role)
						return nil
					}).Times(1)
			},
			wantChannel: &entity.Channel{
				ID:               1,
				SlackChannelID:   "C123456789",
				SlackChannelName: "test-channel",
				SlackTeamID:      "T123456789",
				IsActive:         true,
			},
			wantCreated: true,
			wantErr:     false,
		},
		{
			name: "Should return existing channel",
			args: args{
				slackChannelID:   "C123456789",
				slackChannelName: "test-channel",
				slackTeamID:      "T123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				existingChannel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.slackChannelID,
					SlackChannelName: "existing-channel",
					SlackTeamID:      args.slackTeamID,
					IsActive:         true,
				}

				mocks.mockChannelRepo.EXPECT().
					GetBySlackID(args.slackChannelID).
					Return(existingChannel, nil).Times(1)
			},
			wantChannel: &entity.Channel{
				ID:               1,
				SlackChannelID:   "C123456789",
				SlackChannelName: "existing-channel",
				SlackTeamID:      "T123456789",
				IsActive:         true,
			},
			wantCreated: false,
			wantErr:     false,
		},
		{
			name: "Should return error when channel check fails",
			args: args{
				slackChannelID:   "C123456789",
				slackChannelName: "test-channel",
				slackTeamID:      "T123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetBySlackID(args.slackChannelID).
					Return(nil, assert.AnError).Times(1)
			},
			wantChannel: nil,
			wantCreated: false,
			wantErr:     true,
		},
		{
			name: "Should return error when channel creation fails",
			args: args{
				slackChannelID:   "C123456789",
				slackChannelName: "test-channel",
				slackTeamID:      "T123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetBySlackID(args.slackChannelID).
					Return(nil, nil).Times(1)

				mocks.mockChannelRepo.EXPECT().
					Create(gomock.Any()).
					Return(assert.AnError).Times(1)
			},
			wantChannel: nil,
			wantCreated: false,
			wantErr:     true,
		},
		{
			name: "Should return error when scheduler creation fails",
			args: args{
				slackChannelID:   "C123456789",
				slackChannelName: "test-channel",
				slackTeamID:      "T123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetBySlackID(args.slackChannelID).
					Return(nil, nil).Times(1)

				mocks.mockChannelRepo.EXPECT().
					Create(gomock.Any()).
					DoAndReturn(func(c *entity.Channel) error {
						c.ID = 1
						return nil
					}).Times(1)

				mocks.mockSchedulerRepo.EXPECT().
					Create(gomock.Any()).
					Return(assert.AnError).Times(1)
			},
			wantChannel: nil,
			wantCreated: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			gotChannel, gotCreated, err := s.SetupChannel(tt.args.slackChannelID, tt.args.slackChannelName, tt.args.slackTeamID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCreated, gotCreated)
			assert.Equal(t, tt.wantChannel, gotChannel)
		})
	}
}

func Test_rotationService_AddUser(t *testing.T) {
	type args struct {
		channelID   int64
		slackUserID string
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should add new user successfully",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				slackUser := &slack.User{
					ID:   args.slackUserID,
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName:    "Test User",
						DisplayName: "Test Display",
					},
				}

				gomock.InOrder(
					mocks.mockSlackClient.EXPECT().
						GetUserInfo(args.slackUserID).
						Return(slackUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(nil, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						Create(gomock.Any()).
						DoAndReturn(func(user *entity.User) error {
							require.Equal(t, args.channelID, user.ChannelID)
							require.Equal(t, args.slackUserID, user.SlackUserID)
							require.Equal(t, slackUser.Name, user.SlackUserName)
							require.Equal(t, slackUser.Profile.RealName, user.DisplayName)
							require.True(t, user.IsActive)
							user.ID = 1
							return nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when user already exists",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				slackUser := &slack.User{
					ID:   args.slackUserID,
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName: "Test User",
					},
				}

				existingUser := &entity.User{
					ID:            1,
					ChannelID:     args.channelID,
					SlackUserID:   args.slackUserID,
					SlackUserName: "testuser",
					DisplayName:   "Test User",
					IsActive:      true,
				}

				gomock.InOrder(
					mocks.mockSlackClient.EXPECT().
						GetUserInfo(args.slackUserID).
						Return(slackUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(existingUser, nil).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when Slack API fails",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSlackClient.EXPECT().
					GetUserInfo(args.slackUserID).
					Return(nil, assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when user repository GetByChannelAndSlackID fails",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				slackUser := &slack.User{
					ID:   args.slackUserID,
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName: "Test User",
					},
				}

				gomock.InOrder(
					mocks.mockSlackClient.EXPECT().
						GetUserInfo(args.slackUserID).
						Return(slackUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(nil, assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when user repository Create fails",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				slackUser := &slack.User{
					ID:   args.slackUserID,
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName: "Test User",
					},
				}

				gomock.InOrder(
					mocks.mockSlackClient.EXPECT().
						GetUserInfo(args.slackUserID).
						Return(slackUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(nil, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						Create(gomock.Any()).
						Return(assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.AddUser(tt.args.channelID, tt.args.slackUserID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_RemoveUser(t *testing.T) {
	type args struct {
		channelID   int64
		slackUserID string
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should remove user successfully",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				existingUser := &entity.User{
					ID:            1,
					ChannelID:     args.channelID,
					SlackUserID:   args.slackUserID,
					SlackUserName: "testuser",
					DisplayName:   "Test User",
					IsActive:      true,
				}

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(existingUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						Delete(existingUser.ID).
						Return(nil).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when user not found",
			args: args{
				channelID:   1,
				slackUserID: "U999999999",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetByChannelAndSlackID(args.channelID, args.slackUserID).
					Return(nil, nil).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.RemoveUser(tt.args.channelID, tt.args.slackUserID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_GetNextPresenter(t *testing.T) {
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
			name: "Should return first user when no last presenter",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
					{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
				}

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetLastPresenter(args.channelID).
						Return(nil, nil).Times(1),
				)
			},
			want: &entity.User{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
			wantErr: false,
		},
		{
			name: "Should return next user in rotation",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
					{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
					{ID: 3, SlackUserID: "U555555555", SlackUserName: "user3"},
				}

				lastPresenter := users[0] // First user was last

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetLastPresenter(args.channelID).
						Return(lastPresenter, nil).Times(1),
				)
			},
			want: &entity.User{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
			wantErr: false,
		},
		{
			name: "Should return error when no users",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return([]*entity.User{}, nil).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.GetNextPresenter(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_RecordPresentation(t *testing.T) {
	type args struct {
		ctx       context.Context
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
				ctx:       context.Background(),
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				// Mock transaction
				mocks.mockDataManager.EXPECT().
					WithTransaction(args.ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						// Call the function with mocks that simulate transaction behavior
						return fn(mocks.mockDataManager)
					}).Times(1)

				mocks.mockUserRepo.EXPECT().ClearLastPresenter(args.channelID).Return(nil).Times(1)
				mocks.mockUserRepo.EXPECT().SetLastPresenter(args.userID).Return(nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.RecordPresentation(tt.args.ctx, tt.args.channelID, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_UpdateChannelConfig(t *testing.T) {
	type args struct {
		channelID  int64
		configType string
		value      string
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should update notification time",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "14:30",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(scheduler, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Update(gomock.Any()).
						DoAndReturn(func(s *entity.Scheduler) error {
							require.Equal(t, args.value, s.NotificationTime)
							return nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should update role",
			args: args{
				channelID:  1,
				configType: "role",
				value:      "presenter",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(scheduler, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Update(gomock.Any()).
						DoAndReturn(func(s *entity.Scheduler) error {
							require.Equal(t, args.value, s.Role)
							return nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error for invalid time format",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "25:99",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(scheduler, nil).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.UpdateChannelConfig(tt.args.channelID, tt.args.configType, tt.args.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_ListUsers(t *testing.T) {
	type args struct {
		channelID int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		want      []*entity.User
		wantErr   bool
	}{
		{
			name: "Should return list of users successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				expectedUsers := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
					{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
				}

				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(expectedUsers, nil).Times(1)
			},
			want: []*entity.User{
				{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
				{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
			},
			wantErr: false,
		},
		{
			name: "Should return empty list when no users",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return([]*entity.User{}, nil).Times(1)
			},
			want:    []*entity.User{},
			wantErr: false,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.ListUsers(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_GetCurrentPresenter(t *testing.T) {
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
			name: "Should return current presenter successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				currentPresenter := &entity.User{
					ID:            1,
					SlackUserID:   "U123456789",
					SlackUserName: "user1",
					LastPresenter: true,
				}

				mocks.mockUserRepo.EXPECT().
					GetLastPresenter(args.channelID).
					Return(currentPresenter, nil).Times(1)
			},
			want: &entity.User{
				ID:            1,
				SlackUserID:   "U123456789",
				SlackUserName: "user1",
				LastPresenter: true,
			},
			wantErr: false,
		},
		{
			name: "Should return nil when no current presenter",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetLastPresenter(args.channelID).
					Return(nil, nil).Times(1)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetLastPresenter(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.GetCurrentPresenter(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_PauseScheduler(t *testing.T) {
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
			name: "Should pause scheduler successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					SetEnabled(args.channelID, false).
					Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					SetEnabled(args.channelID, false).
					Return(assert.AnError).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.PauseScheduler(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_ResumeScheduler(t *testing.T) {
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
			name: "Should resume scheduler successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					SetEnabled(args.channelID, true).
					Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					SetEnabled(args.channelID, true).
					Return(assert.AnError).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.ResumeScheduler(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_rotationService_GetChannelConfig(t *testing.T) {
	type args struct {
		channelID int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		want      *entity.Channel
		wantErr   bool
	}{
		{
			name: "Should return channel config successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   "C123456789",
					SlackChannelName: "test-channel",
					SlackTeamID:      "T123456789",
					IsActive:         true,
				}

				mocks.mockChannelRepo.EXPECT().
					GetByID(args.channelID).
					Return(channel, nil).Times(1)
			},
			want: &entity.Channel{
				ID:               1,
				SlackChannelID:   "C123456789",
				SlackChannelName: "test-channel",
				SlackTeamID:      "T123456789",
				IsActive:         true,
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
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockChannelRepo.EXPECT().
					GetByID(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.GetChannelConfig(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_GetSchedulerConfig(t *testing.T) {
	type args struct {
		channelID int64
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		want      *entity.Scheduler
		wantErr   bool
	}{
		{
			name: "Should return scheduler config successfully",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				schedulerConfig := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: "09:00",
					ActiveDays:       []int{1, 2, 3, 4, 5},
					IsEnabled:        true,
					Role:             "On duty",
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(schedulerConfig, nil).Times(1)
			},
			want: &entity.Scheduler{
				ID:               1,
				ChannelID:        1,
				NotificationTime: "09:00",
				ActiveDays:       []int{1, 2, 3, 4, 5},
				IsEnabled:        true,
				Role:             "On duty",
			},
			wantErr: false,
		},
		{
			name: "Should return nil when scheduler config not found",
			args: args{channelID: 999},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(nil, nil).Times(1)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Should return error when repository fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.GetSchedulerConfig(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_GetChannelStatus(t *testing.T) {
	type args struct {
		channelID int
	}
	tests := []struct {
		name    string
		args    args
		want    *entity.Channel
		wantErr bool
	}{
		{
			name:    "Should return not implemented error",
			args:    args{channelID: 1},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			got, err := s.GetChannelStatus(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not implemented")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_rotationService_UpdateChannelConfig_MoreScenarios(t *testing.T) {
	type args struct {
		channelID  int64
		configType string
		value      string
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should update active days successfully",
			args: args{
				channelID:  1,
				configType: "days",
				value:      "1,3,5",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(scheduler, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Update(gomock.Any()).
						DoAndReturn(func(s *entity.Scheduler) error {
							expected := []int{1, 3, 5}
							require.Equal(t, expected, s.ActiveDays)
							return nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error for invalid config type",
			args: args{
				channelID:  1,
				configType: "invalid",
				value:      "test",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(scheduler, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error for empty role",
			args: args{
				channelID:  1,
				configType: "role",
				value:      "   ",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(scheduler, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error for invalid days",
			args: args{
				channelID:  1,
				configType: "days",
				value:      "invalid,8,9",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(scheduler, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should create scheduler config when not exists",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "10:30",
			},
			buildMock: func(mocks allMocks, args args) {
				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(nil, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Create(gomock.Any()).
						DoAndReturn(func(scheduler *entity.Scheduler) error {
							require.Equal(t, args.channelID, scheduler.ChannelID)
							require.Equal(t, "09:00", scheduler.NotificationTime)
							require.Equal(t, domain.DefaultActiveDays, scheduler.ActiveDays)
							require.True(t, scheduler.IsEnabled)
							require.Equal(t, "On duty", scheduler.Role)
							scheduler.ID = 1
							return nil
						}).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Update(gomock.Any()).
						DoAndReturn(func(s *entity.Scheduler) error {
							require.Equal(t, args.value, s.NotificationTime)
							return nil
						}).Times(1),
				)
			},
			wantErr: false,
		},
		{
			name: "Should return error when get scheduler fails",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "10:30",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockSchedulerRepo.EXPECT().
					GetByChannelID(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when create scheduler fails",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "10:30",
			},
			buildMock: func(mocks allMocks, args args) {
				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(nil, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Create(gomock.Any()).
						Return(assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "Should return error when update scheduler fails",
			args: args{
				channelID:  1,
				configType: "time",
				value:      "10:30",
			},
			buildMock: func(mocks allMocks, args args) {
				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        args.channelID,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				gomock.InOrder(
					mocks.mockSchedulerRepo.EXPECT().
						GetByChannelID(args.channelID).
						Return(scheduler, nil).Times(1),

					mocks.mockSchedulerRepo.EXPECT().
						Update(gomock.Any()).
						Return(assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.UpdateChannelConfig(tt.args.channelID, tt.args.configType, tt.args.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_parseDays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "Should parse valid weekday numbers",
			input:    "1,3,5",
			expected: []int{1, 3, 5},
		},
		{
			name:     "Should parse and sort weekday numbers",
			input:    "5,1,3",
			expected: []int{1, 3, 5},
		},
		{
			name:     "Should handle spaces around numbers",
			input:    " 1 , 3 , 5 ",
			expected: []int{1, 3, 5},
		},
		{
			name:     "Should ignore invalid numbers",
			input:    "1,8,3,9,5",
			expected: []int{1, 3, 5},
		},
		{
			name:     "Should ignore non-numeric strings",
			input:    "1,abc,3,def,5",
			expected: []int{1, 3, 5},
		},
		{
			name:     "Should return empty for all invalid input",
			input:    "abc,def,xyz",
			expected: nil,
		},
		{
			name:     "Should handle single valid number",
			input:    "1",
			expected: []int{1},
		},
		{
			name:     "Should handle Sunday (7) correctly",
			input:    "7,1,6",
			expected: []int{1, 6, 7},
		},
		{
			name:     "Should handle empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "Should handle only spaces",
			input:    "   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDays(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_cleanRoleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple role",
			input:    "presenter",
			expected: "presenter",
		},
		{
			name:     "role with quotes",
			input:    `"presenter"`,
			expected: "presenter",
		},
		{
			name:     "role with single quotes",
			input:    `'presenter'`,
			expected: "presenter",
		},
		{
			name:     "role with equals pattern",
			input:    `= "presenter"`,
			expected: "presenter",
		},
		{
			name:     "role with nested quotes",
			input:    `""presenter""`,
			expected: "presenter",
		},
		{
			name:     "role with spaces",
			input:    `  presenter  `,
			expected: "presenter",
		},
		{
			name:     "multi-word role with quotes",
			input:    `"Daily  presenter"`,
			expected: "Daily presenter",
		},
		{
			name:     "multi-word role without quotes",
			input:    `Code reviewer`,
			expected: "Code reviewer",
		},
		{
			name:     "role with multiple spaces",
			input:    `"Code   reviewer"`,
			expected: "Code reviewer",
		},
		{
			name:     "role with Unicode quotes",
			input:    "\u201cpresenter\u201d",
			expected: "presenter",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "only quotes",
			input:    `""`,
			expected: "",
		},
		{
			name:     "complex pattern with equals and nested quotes",
			input:    "= \"\u201cDaily presenter\u201d\"",
			expected: "Daily presenter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanRoleName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Additional error scenarios for RemoveUser
func Test_rotationService_RemoveUser_ErrorScenarios(t *testing.T) {
	type args struct {
		channelID   int64
		slackUserID string
	}
	tests := []struct {
		name      string
		buildMock func(mocks allMocks, args args)
		args      args
		wantErr   bool
	}{
		{
			name: "Should return error when user repository GetByChannelAndSlackID fails",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetByChannelAndSlackID(args.channelID, args.slackUserID).
					Return(nil, assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when user repository Delete fails",
			args: args{
				channelID:   1,
				slackUserID: "U123456789",
			},
			buildMock: func(mocks allMocks, args args) {
				existingUser := &entity.User{
					ID:            1,
					ChannelID:     args.channelID,
					SlackUserID:   args.slackUserID,
					SlackUserName: "testuser",
					DisplayName:   "Test User",
					IsActive:      true,
				}

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().
						GetByChannelAndSlackID(args.channelID, args.slackUserID).
						Return(existingUser, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						Delete(existingUser.ID).
						Return(assert.AnError).Times(1),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.RemoveUser(tt.args.channelID, tt.args.slackUserID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Additional error scenarios for GetNextPresenter
func Test_rotationService_GetNextPresenter_ErrorScenarios(t *testing.T) {
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
			name: "Should return error when GetActiveUsersByChannel fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockUserRepo.EXPECT().
					GetActiveUsersByChannel(args.channelID).
					Return(nil, assert.AnError).Times(1)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return error when GetLastPresenter fails",
			args: args{channelID: 1},
			buildMock: func(mocks allMocks, args args) {
				users := []*entity.User{
					{ID: 1, SlackUserID: "U123456789", SlackUserName: "user1"},
					{ID: 2, SlackUserID: "U987654321", SlackUserName: "user2"},
				}

				gomock.InOrder(
					mocks.mockUserRepo.EXPECT().
						GetActiveUsersByChannel(args.channelID).
						Return(users, nil).Times(1),

					mocks.mockUserRepo.EXPECT().
						GetLastPresenter(args.channelID).
						Return(nil, assert.AnError).Times(1),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ctrl := newServiceTestMock(t)
			defer ctrl.Finish()

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			got, err := s.GetNextPresenter(tt.args.channelID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Improved RecordPresentation tests following go_boilerplate patterns
func Test_rotationService_RecordPresentation_Enhanced(t *testing.T) {
	type args struct {
		ctx       context.Context
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
				ctx:       context.Background(),
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(args.ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(contract.DataManager) error) error {
						return fn(mocks.mockDataManager)
					}).Times(1)

				mocks.mockUserRepo.EXPECT().ClearLastPresenter(args.channelID).Return(nil).Times(1)
				mocks.mockUserRepo.EXPECT().SetLastPresenter(args.userID).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Should return error when transaction fails",
			args: args{
				ctx:       context.Background(),
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(args.ctx, gomock.Any()).
					Return(assert.AnError).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Should return error when ClearLastPresenter fails",
			args: args{
				ctx:       context.Background(),
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(args.ctx, gomock.Any()).
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
				ctx:       context.Background(),
				channelID: 1,
				userID:    2,
			},
			buildMock: func(mocks allMocks, args args) {
				mocks.mockDataManager.EXPECT().
					WithTransaction(args.ctx, gomock.Any()).
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

			s := newRotation(m.mockDataManager, m.mockSlackClient)

			if tt.buildMock != nil {
				tt.buildMock(m, tt.args)
			}

			err := s.RecordPresentation(tt.args.ctx, tt.args.channelID, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}