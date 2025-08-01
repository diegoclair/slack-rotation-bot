package entity

import "time"

type Channel struct {
	ID               int64     `json:"id" db:"id"`
	SlackChannelID   string    `json:"slack_channel_id" db:"slack_channel_id"`
	SlackChannelName string    `json:"slack_channel_name" db:"slack_channel_name"`
	SlackTeamID      string    `json:"slack_team_id" db:"slack_team_id"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Scheduler struct {
	ID               int64     `json:"id" db:"id"`
	ChannelID        int64     `json:"channel_id" db:"channel_id"`
	NotificationTime string    `json:"notification_time" db:"notification_time"` // HH:MM format in UTC
	ActiveDays       []int     `json:"active_days" db:"active_days"`             // ISO 8601 weekdays (1-7)
	IsEnabled        bool      `json:"is_enabled" db:"is_enabled"`               // Scheduler enabled/disabled
	Role             string    `json:"role" db:"role"`                           // Role name (e.g., "presenter", "reviewer", "On duty")
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID            int64     `json:"id" db:"id"`
	ChannelID     int64     `json:"channel_id" db:"channel_id"`
	SlackUserID   string    `json:"slack_user_id" db:"slack_user_id"`
	SlackUserName string    `json:"slack_user_name" db:"slack_user_name"`
	DisplayName   string    `json:"display_name" db:"display_name"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	LastPresenter bool      `json:"last_presenter" db:"last_presenter"`
	JoinedAt      time.Time `json:"joined_at" db:"joined_at"`
}

// GetDisplayName returns the best available name for display
func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.SlackUserName != "" {
		return u.SlackUserName
	}
	return "Unknown User"
}
