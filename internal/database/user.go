package database

import (
	"database/sql"
	"fmt"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

type userRepo struct {
	db dbConn
}

func newUserRepo(db dbConn) contract.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(user *entity.User) error {
	query := `
		INSERT INTO users (channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
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

	user.ID = id
	return nil
}

func (r *userRepo) GetByChannelAndSlackID(channelID int64, slackUserID string) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
		FROM users
		WHERE channel_id = ? AND slack_user_id = ?
	`

	err := r.db.QueryRow(query, channelID, slackUserID).Scan(
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

func (r *userRepo) GetActiveUsersByChannel(channelID int64) ([]*entity.User, error) {
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
		FROM users
		WHERE channel_id = ? AND is_active = 1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.Query(query, channelID)
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

func (r *userRepo) Delete(userID int64) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *userRepo) ClearLastPresenter(channelID int64) error {
	query := `UPDATE users SET last_presenter = 0 WHERE channel_id = ?`
	_, err := r.db.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("failed to clear last presenter: %w", err)
	}
	return nil
}

func (r *userRepo) SetLastPresenter(userID int64) error {
	query := `UPDATE users SET last_presenter = 1 WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to set last presenter: %w", err)
	}
	return nil
}

func (r *userRepo) GetLastPresenter(channelID int64) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, channel_id, slack_user_id, slack_user_name, display_name, is_active, last_presenter, joined_at
		FROM users
		WHERE channel_id = ? AND last_presenter = 1
		LIMIT 1
	`

	err := r.db.QueryRow(query, channelID).Scan(
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
