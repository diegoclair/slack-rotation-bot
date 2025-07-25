package contract

import (
	"context"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

type RotationService interface {
	SetupChannel(slackChannelID, channelName, teamID string) (*entity.Channel, bool, error)
	AddUser(channelID int64, slackUserID string) error
	RemoveUser(channelID int64, slackUserID string) error
	GetNextPresenter(channelID int64) (*entity.User, error)
	RecordPresentation(ctx context.Context, channelID, userID int64) error
	UpdateChannelConfig(channelID int64, configType, configValue string) error
	ListUsers(channelID int64) ([]*entity.User, error)
	GetCurrentPresenter(channelID int64) (*entity.User, error)
	PauseScheduler(channelID int64) error
	ResumeScheduler(channelID int64) error
	GetChannelConfig(channelID int64) (*entity.Channel, error)
	GetSchedulerConfig(channelID int64) (*entity.Scheduler, error)
	GetChannelStatus(channelID int) (*entity.Channel, error)
}