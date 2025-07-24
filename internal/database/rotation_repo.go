package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/diegoclair/slack-rotation-bot/pkg/models"
)

type RotationRepository struct {
	db *DB
}

func NewRotationRepository(db *DB) *RotationRepository {
	return &RotationRepository{db: db}
}

func (r *RotationRepository) Create(rotation *models.Rotation) error {
	query := `
		INSERT INTO rotations (channel_id, user_id, presented_at, was_presenter, skipped_reason)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.conn.Exec(query,
		rotation.ChannelID,
		rotation.UserID,
		rotation.PresentedAt,
		rotation.WasPresenter,
		rotation.SkippedReason,
	)
	if err != nil {
		return fmt.Errorf("failed to create rotation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	rotation.ID = int(id)
	return nil
}

func (r *RotationRepository) GetLastPresenterByChannel(channelID int) (*models.Rotation, error) {
	rotation := &models.Rotation{}
	query := `
		SELECT id, channel_id, user_id, presented_at, was_presenter, skipped_reason
		FROM rotations
		WHERE channel_id = ? AND was_presenter = 1
		ORDER BY presented_at DESC, id DESC
		LIMIT 1
	`

	err := r.db.conn.QueryRow(query, channelID).Scan(
		&rotation.ID,
		&rotation.ChannelID,
		&rotation.UserID,
		&rotation.PresentedAt,
		&rotation.WasPresenter,
		&rotation.SkippedReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last presenter: %w", err)
	}

	return rotation, nil
}

func (r *RotationRepository) GetTodaysPresenter(channelID int, date time.Time) (*models.Rotation, error) {
	rotation := &models.Rotation{}
	dateStr := date.Format("2006-01-02")
	
	query := `
		SELECT id, channel_id, user_id, presented_at, was_presenter, skipped_reason
		FROM rotations
		WHERE channel_id = ? AND DATE(presented_at) = DATE(?)
		ORDER BY id DESC
		LIMIT 1
	`

	err := r.db.conn.QueryRow(query, channelID, dateStr).Scan(
		&rotation.ID,
		&rotation.ChannelID,
		&rotation.UserID,
		&rotation.PresentedAt,
		&rotation.WasPresenter,
		&rotation.SkippedReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get today's presenter: %w", err)
	}

	return rotation, nil
}


func (r *RotationRepository) GetUserLastPresentation(userID int) (*models.Rotation, error) {
	rotation := &models.Rotation{}
	query := `
		SELECT id, channel_id, user_id, presented_at, was_presenter, skipped_reason
		FROM rotations
		WHERE user_id = ? AND was_presenter = 1
		ORDER BY presented_at DESC
		LIMIT 1
	`

	err := r.db.conn.QueryRow(query, userID).Scan(
		&rotation.ID,
		&rotation.ChannelID,
		&rotation.UserID,
		&rotation.PresentedAt,
		&rotation.WasPresenter,
		&rotation.SkippedReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user last presentation: %w", err)
	}

	return rotation, nil
}

