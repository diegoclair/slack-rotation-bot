package contract

import (
	"context"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

// DataManager aggregates all repository interfaces
type DataManager interface {
	WithTransaction(ctx context.Context, fn func(dm DataManager) error) error
	Channel() ChannelRepo
	User() UserRepo
}

// ChannelRepo defines the contract for channel repository
type ChannelRepo interface {
	Create(channel *entity.Channel) error
	GetBySlackID(slackChannelID string) (*entity.Channel, error)
	GetByID(id int64) (*entity.Channel, error)
	Update(channel *entity.Channel) error
	GetActiveChannels() ([]*entity.Channel, error)
}

// UserRepo defines the contract for user repository
type UserRepo interface {
	Create(user *entity.User) error
	GetByChannelAndSlackID(channelID int64, slackUserID string) (*entity.User, error)
	GetActiveUsersByChannel(channelID int64) ([]*entity.User, error)
	Delete(userID int64) error
	ClearLastPresenter(channelID int64) error
	SetLastPresenter(userID int64) error
	GetLastPresenter(channelID int64) (*entity.User, error)
}
