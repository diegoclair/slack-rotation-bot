package database

import (
	"database/sql"
	"fmt"

	"github.com/diegoclair/slack-rotation-bot/pkg/models"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (channel_id, slack_user_id, slack_user_name, display_name, is_active)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.conn.Exec(query,
		user.ChannelID,
		user.SlackUserID,
		user.SlackUserName,
		user.DisplayName,
		user.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)
	return nil
}

func (r *UserRepository) GetByChannelAndSlackID(channelID int, slackUserID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, joined_at
		FROM users
		WHERE channel_id = ? AND slack_user_id = ?
	`

	err := r.db.conn.QueryRow(query, channelID, slackUserID).Scan(
		&user.ID,
		&user.ChannelID,
		&user.SlackUserID,
		&user.SlackUserName,
		&user.DisplayName,
		&user.IsActive,
		&user.JoinedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetActiveUsersByChannel(channelID int) ([]*models.User, error) {
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, joined_at
		FROM users
		WHERE channel_id = ? AND is_active = 1
		ORDER BY display_name ASC
	`

	rows, err := r.db.conn.Query(query, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.ChannelID,
			&user.SlackUserID,
			&user.SlackUserName,
			&user.DisplayName,
			&user.IsActive,
			&user.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) UpdateActiveStatus(userID int, isActive bool) error {
	query := `UPDATE users SET is_active = ? WHERE id = ?`
	
	_, err := r.db.conn.Exec(query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = ?`
	
	_, err := r.db.conn.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}