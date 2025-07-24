package database

import (
	"database/sql"
	"fmt"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *entity.User) error {
	query := `
		INSERT INTO users (channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.conn.Exec(query,
		user.ChannelID,
		user.SlackUserID,
		user.SlackUserName,
		user.DisplayName,
		user.IsActive,
		user.LastPresenter,
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

func (r *UserRepository) GetByChannelAndSlackID(channelID int, slackUserID string) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
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
		&user.LastPresenter,
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

func (r *UserRepository) GetActiveUsersByChannel(channelID int) ([]*entity.User, error) {
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
		FROM users
		WHERE channel_id = ? AND is_active = 1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.conn.Query(query, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.ChannelID,
			&user.SlackUserID,
			&user.SlackUserName,
			&user.DisplayName,
			&user.IsActive,
			&user.LastPresenter,
			&user.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}


func (r *UserRepository) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = ?`
	
	_, err := r.db.conn.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *UserRepository) SetLastPresenter(channelID, userID int) error {
	tx, err := r.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear previous last_presenter
	_, err = tx.Exec(`UPDATE users SET last_presenter = 0 WHERE channel_id = ?`, channelID)
	if err != nil {
		return fmt.Errorf("failed to clear last presenter: %w", err)
	}

	// Set new last_presenter
	_, err = tx.Exec(`UPDATE users SET last_presenter = 1 WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to set last presenter: %w", err)
	}

	return tx.Commit()
}

func (r *UserRepository) GetLastPresenter(channelID int) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
		FROM users
		WHERE channel_id = ? AND last_presenter = 1
		LIMIT 1
	`

	err := r.db.conn.QueryRow(query, channelID).Scan(
		&user.ID,
		&user.ChannelID,
		&user.SlackUserID,
		&user.SlackUserName,
		&user.DisplayName,
		&user.IsActive,
		&user.LastPresenter,
		&user.JoinedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last presenter: %w", err)
	}

	return user, nil
}


