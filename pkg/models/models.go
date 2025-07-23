package models

import "time"

type Channel struct {
	ID               int       `json:"id" db:"id"`
	SlackChannelID   string    `json:"slack_channel_id" db:"slack_channel_id"`
	SlackChannelName string    `json:"slack_channel_name" db:"slack_channel_name"`
	SlackTeamID      string    `json:"slack_team_id" db:"slack_team_id"`
	NotificationTime string    `json:"notification_time" db:"notification_time"` // HH:MM format
	ActiveDays       string    `json:"active_days" db:"active_days"`           // JSON array of weekdays
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID             int       `json:"id" db:"id"`
	ChannelID      int       `json:"channel_id" db:"channel_id"`
	SlackUserID    string    `json:"slack_user_id" db:"slack_user_id"`
	SlackUserName  string    `json:"slack_user_name" db:"slack_user_name"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
}

type Rotation struct {
	ID           int       `json:"id" db:"id"`
	ChannelID    int       `json:"channel_id" db:"channel_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	PresentedAt  time.Time `json:"presented_at" db:"presented_at"`
	WasPresenter bool      `json:"was_presenter" db:"was_presenter"`
	SkippedReason string   `json:"skipped_reason,omitempty" db:"skipped_reason"`
}