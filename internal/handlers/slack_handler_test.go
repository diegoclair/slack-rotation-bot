package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diegoclair/slack-rotation-bot/internal/domain"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/diegoclair/slack-rotation-bot/internal/handlers/test"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSlackHandler_HandleSlashCommand_AddUser(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should add user successfully",
			args: args{
				command:     "/rotation",
				text:        "add <@U123456789|testuser>",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock AddUser call
				m.RotationServiceMock.EXPECT().
					AddUser(int64(1), "U123456789").
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeInChannel, response.ResponseType)
				assert.Contains(t, response.Text, "✅ <@U123456789> has been added to the rotation!")
			},
		},
		{
			name: "Should add multiple users successfully",
			args: args{
				command:     "/rotation",
				text:        "add <@U123456789|testuser> <@U987654321|testuser2>",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock AddUser calls for both users
				m.RotationServiceMock.EXPECT().
					AddUser(int64(1), "U123456789").
					Return(nil).Times(1)
				m.RotationServiceMock.EXPECT().
					AddUser(int64(1), "U987654321").
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeInChannel, response.ResponseType)
				assert.Contains(t, response.Text, "✅ 2 users added to the rotation: <@U123456789>, <@U987654321>")
			},
		},
		{
			name: "Should return error when no user mentioned",
			args: args{
				command:     "/rotation",
				text:        "add",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "❌ Please mention at least one user: `/rotation add @user1 @user2`")
			},
		},
		{
			name: "Should show user display name in error message when add fails",
			args: args{
				command:     "/rotation",
				text:        "add <@U123456789|testuser>",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock AddUser call to fail
				m.RotationServiceMock.EXPECT().
					AddUser(int64(1), "U123456789").
					Return(errors.New("user already exists")).Times(1)

				// Mock GetUserInfo call for error case
				userInfo := &slack.User{
					ID:   "U123456789",
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName:    "João Silva",
						DisplayName: "joao.silva",
					},
				}
				m.SlackClientMock.EXPECT().
					GetUserInfo("U123456789").
					Return(userInfo, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "❌ Failed to add: João Silva")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_ListUsers(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should list users successfully",
			args: args{
				command:     "/rotation",
				text:        "list",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				users := []*entity.User{
					{
						ID:          1,
						ChannelID:   1,
						SlackUserID: "U123456789",
						DisplayName: "Test User 1",
						IsActive:    true,
					},
					{
						ID:          2,
						ChannelID:   1,
						SlackUserID: "U987654321",
						DisplayName: "Test User 2",
						IsActive:    true,
					},
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock ListUsers call
				m.RotationServiceMock.EXPECT().
					ListUsers(int64(1)).
					Return(users, nil).Times(1)
				
				// Mock GetCurrentPresenter call
				m.RotationServiceMock.EXPECT().
					GetCurrentPresenter(int64(1)).
					Return(users[0], nil).Times(1)
				
				// Mock GetSchedulerConfig call
				m.RotationServiceMock.EXPECT().
					GetSchedulerConfig(int64(1)).
					Return(&entity.Scheduler{
						Role: "presenter",
					}, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "*Members in rotation:*")
				assert.Contains(t, response.Text, "👉 1. Test User 1 *(presenter today)*")
				assert.Contains(t, response.Text, "2. Test User 2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Help(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should show help message",
			args: args{
				command:     "/rotation",
				text:        "help",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "*🔄 People Rotation Bot - Commands*")
				assert.Contains(t, response.Text, "*⚙️ Configuration:*")
				assert.Contains(t, response.Text, "*👥 Member Management:*")
				assert.Contains(t, response.Text, "*⏸️ Notification Control:*")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Next(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should advance to next presenter successfully",
			args: args{
				command:     "/rotation",
				text:        "next",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				nextUser := &entity.User{
					ID:            2,
					ChannelID:     1,
					SlackUserID:   "U234567890",
					DisplayName:   "Next User",
					IsActive:      true,
					LastPresenter: false,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock GetNextPresenter call
				m.RotationServiceMock.EXPECT().
					GetNextPresenter(int64(1)).
					Return(nextUser, nil).Times(1)

				// Mock RecordPresentation call
				m.RotationServiceMock.EXPECT().
					RecordPresentation(gomock.Any(), int64(1), int64(2)).
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeInChannel, response.ResponseType)
				assert.Contains(t, response.Text, "⏭️ Skipping to next presenter: <@U234567890>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_RemoveUser(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should remove multiple users successfully",
			args: args{
				command:     "/rotation",
				text:        "remove <@U123456789|testuser> <@U987654321|testuser2>",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U555555555",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock RemoveUser calls for both users
				m.RotationServiceMock.EXPECT().
					RemoveUser(int64(1), "U123456789").
					Return(nil).Times(1)
				m.RotationServiceMock.EXPECT().
					RemoveUser(int64(1), "U987654321").
					Return(nil).Times(1)

				// Mock GetUserInfo calls
				userInfo1 := &slack.User{
					ID:      "U123456789",
					Name:    "testuser",
					Profile: slack.UserProfile{RealName: "Test User 1"},
				}
				userInfo2 := &slack.User{
					ID:      "U987654321",
					Name:    "testuser2",
					Profile: slack.UserProfile{RealName: "Test User 2"},
				}
				m.SlackClientMock.EXPECT().
					GetUserInfo("U123456789").
					Return(userInfo1, nil).Times(1)
				m.SlackClientMock.EXPECT().
					GetUserInfo("U987654321").
					Return(userInfo2, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeInChannel, response.ResponseType)
				assert.Contains(t, response.Text, "✅ 2 users removed from the rotation: Test User 1, Test User 2")
			},
		},
		{
			name: "Should remove user successfully",
			args: args{
				command:     "/rotation",
				text:        "remove <@U123456789|testuser>",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock RemoveUser call
				m.RotationServiceMock.EXPECT().
					RemoveUser(int64(1), "U123456789").
					Return(nil).Times(1)

				// Mock GetUserInfo call
				userInfo := &slack.User{
					ID:   "U123456789",
					Name: "testuser",
					Profile: slack.UserProfile{
						RealName:    "Test User",
						DisplayName: "Test Display",
					},
				}
				m.SlackClientMock.EXPECT().
					GetUserInfo("U123456789").
					Return(userInfo, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeInChannel, response.ResponseType)
				assert.Contains(t, response.Text, "✅ Test User has been removed from the rotation.")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Config(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should show current configuration",
			args: args{
				command:     "/rotation",
				text:        "config show",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: "09:30",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "presenter",
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock GetChannelConfig call
				m.RotationServiceMock.EXPECT().
					GetChannelConfig(int64(1)).
					Return(channel, nil).Times(1)

				// Mock GetSchedulerConfig call
				m.RotationServiceMock.EXPECT().
					GetSchedulerConfig(int64(1)).
					Return(scheduler, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "📋 *Current Configuration for #test-channel*")
				assert.Contains(t, response.Text, "⏰ *Notification Time:* 09:30")
				assert.Contains(t, response.Text, "🔔 *Channel Status:* Active")
				assert.Contains(t, response.Text, "📅 *Scheduler Status:* Enabled")
			},
		},
		{
			name: "Should update configuration",
			args: args{
				command:     "/rotation",
				text:        "config time 10:00",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock UpdateChannelConfig call
				m.RotationServiceMock.EXPECT().
					UpdateChannelConfig(int64(1), "time", "10:00").
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "✅ Configuration updated: time = 10:00")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Pause(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should pause scheduler successfully",
			args: args{
				command:     "/rotation",
				text:        "pause",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				scheduler := &entity.Scheduler{
					ID:        1,
					ChannelID: 1,
					IsEnabled: true,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock GetSchedulerConfig call
				m.RotationServiceMock.EXPECT().
					GetSchedulerConfig(int64(1)).
					Return(scheduler, nil).Times(1)

				// Mock PauseScheduler call
				m.RotationServiceMock.EXPECT().
					PauseScheduler(int64(1)).
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "⏸️ Daily rotation notifications have been paused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Resume(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should resume scheduler successfully",
			args: args{
				command:     "/rotation",
				text:        "resume",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				scheduler := &entity.Scheduler{
					ID:        1,
					ChannelID: 1,
					IsEnabled: false,
				}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock GetSchedulerConfig call
				m.RotationServiceMock.EXPECT().
					GetSchedulerConfig(int64(1)).
					Return(scheduler, nil).Times(1)

				// Mock ResumeScheduler call
				m.RotationServiceMock.EXPECT().
					ResumeScheduler(int64(1)).
					Return(nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "▶️ Daily rotation notifications have been resumed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}

func TestSlackHandler_HandleSlashCommand_Status(t *testing.T) {
	type args struct {
		command     string
		text        string
		channelID   string
		channelName string
		userID      string
		teamID      string
	}

	tests := []struct {
		name          string
		args          args
		buildMocks    func(ctx context.Context, m test.ServiceMocks, args args)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Should show status successfully",
			args: args{
				command:     "/rotation",
				text:        "status",
				channelID:   "C123456789",
				channelName: "test-channel",
				userID:      "U987654321",
				teamID:      "T123456789",
			},
			buildMocks: func(ctx context.Context, m test.ServiceMocks, args args) {
				channel := &entity.Channel{
					ID:               1,
					SlackChannelID:   args.channelID,
					SlackChannelName: args.channelName,
					SlackTeamID:      args.teamID,
					IsActive:         true,
				}

				scheduler := &entity.Scheduler{
					ID:               1,
					ChannelID:        1,
					NotificationTime: "09:00",
					ActiveDays:       domain.DefaultActiveDays,
					IsEnabled:        true,
					Role:             "On duty",
				}

				currentUser := &entity.User{
					ID:          1,
					SlackUserID: "U123456789",
				}

				nextUser := &entity.User{
					ID:          2,
					SlackUserID: "U987654321",
				}

				users := []*entity.User{currentUser, nextUser}

				// Mock SetupChannel call
				m.RotationServiceMock.EXPECT().
					SetupChannel(args.channelID, args.channelName, args.teamID).
					Return(channel, false, nil).Times(1)

				// Mock GetChannelConfig call
				m.RotationServiceMock.EXPECT().
					GetChannelConfig(int64(1)).
					Return(channel, nil).Times(1)

				// Mock GetSchedulerConfig call
				m.RotationServiceMock.EXPECT().
					GetSchedulerConfig(int64(1)).
					Return(scheduler, nil).Times(1)

				// Mock GetCurrentPresenter call
				m.RotationServiceMock.EXPECT().
					GetCurrentPresenter(int64(1)).
					Return(currentUser, nil).Times(1)

				// Mock GetNextPresenter call
				m.RotationServiceMock.EXPECT().
					GetNextPresenter(int64(1)).
					Return(nextUser, nil).Times(1)

				// Mock ListUsers call
				m.RotationServiceMock.EXPECT().
					ListUsers(int64(1)).
					Return(users, nil).Times(1)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)

				var response slack.Msg
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, slack.ResponseTypeEphemeral, response.ResponseType)
				assert.Contains(t, response.Text, "📊 *Rotation Status for #test-channel*")
				assert.Contains(t, response.Text, "🔔 *Channel Status:* Active")
				assert.Contains(t, response.Text, "📅 *Scheduler Status:* Enabled ✅")
				assert.Contains(t, response.Text, "👥 *Total Members:* 2")
				assert.Contains(t, response.Text, "🎯 *Current On duty:* <@U123456789>")
				assert.Contains(t, response.Text, "⏭️ *Next On duty:* <@U987654321>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, handler, ctrl := test.GetHandlerTest(t)
			defer ctrl.Finish()

			if tt.buildMocks != nil {
				tt.buildMocks(context.Background(), m, tt.args)
			}

			recorder := test.CreateTestRecorder()
			req := test.CreateSlackRequest(t, tt.args.command, tt.args.text, tt.args.channelID, tt.args.channelName, tt.args.userID, tt.args.teamID, "test-signing-secret")

			handler.HandleSlashCommand(recorder, req)

			if tt.checkResponse != nil {
				tt.checkResponse(t, recorder)
			}
		})
	}
}
