package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

type channelRepository struct {
	db dbConn
}

func newChannelRepository(db dbConn) contract.ChannelRepo {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(channel *entity.Channel) error {
	query := `
		INSERT INTO channels (slack_channel_id, slack_channel_name, slack_team_id, is_active)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		channel.SlackChannelID,
		channel.SlackChannelName,
		channel.SlackTeamID,
		channel.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	channel.ID = id
	return nil
}

func (r *channelRepository) GetBySlackID(slackChannelID string) (*entity.Channel, error) {
	channel := &entity.Channel{}
	query := `
		SELECT id, slack_channel_id, slack_channel_name, slack_team_id,
			is_active, created_at, updated_at
		FROM channels
		WHERE slack_channel_id = ?
	`

	err := r.db.QueryRow(query, slackChannelID).Scan(
		&channel.ID,
		&channel.SlackChannelID,
		&channel.SlackChannelName,
		&channel.SlackTeamID,
		&channel.IsActive,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return channel, nil
}

func (r *channelRepository) GetByID(id int64) (*entity.Channel, error) {
	channel := &entity.Channel{}
	query := `
		SELECT id, slack_channel_id, slack_channel_name, slack_team_id,
			is_active, created_at, updated_at
		FROM channels
		WHERE id = ?
	`

	err := r.db.QueryRow(query, id).Scan(
		&channel.ID,
		&channel.SlackChannelID,
		&channel.SlackChannelName,
		&channel.SlackTeamID,
		&channel.IsActive,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return channel, nil
}

func (r *channelRepository) Update(channel *entity.Channel) error {
	query := `
		UPDATE channels SET
			slack_channel_name = ?,
			is_active = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		channel.SlackChannelName,
		channel.IsActive,
		time.Now(),
		channel.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

func (r *channelRepository) GetActiveChannels() ([]*entity.Channel, error) {
	query := `
		SELECT id, slack_channel_id, slack_channel_name, slack_team_id,
			is_active, created_at, updated_at
		FROM channels
		WHERE is_active = 1
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}
	defer rows.Close()

	var channels []*entity.Channel
	for rows.Next() {
		channel := &entity.Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.SlackChannelID,
			&channel.SlackChannelName,
			&channel.SlackTeamID,
			&channel.IsActive,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}
