package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
)

type schedulerRepository struct {
	db dbConn
}

func newSchedulerRepository(db dbConn) contract.SchedulerRepo {
	return &schedulerRepository{db: db}
}

func (r *schedulerRepository) Create(scheduler *entity.Scheduler) error {
	query := `
		INSERT INTO scheduler_configs (channel_id, notification_time, active_days, is_enabled, role)
		VALUES (?, ?, ?, ?, ?)
	`

	// Convert ActiveDays to JSON for storage
	activeDaysJSON, err := json.Marshal(scheduler.ActiveDays)
	if err != nil {
		return fmt.Errorf("failed to marshal active days: %w", err)
	}

	result, err := r.db.Exec(query,
		scheduler.ChannelID,
		scheduler.NotificationTime,
		string(activeDaysJSON),
		scheduler.IsEnabled,
		scheduler.Role,
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	scheduler.ID = id
	return nil
}

func (r *schedulerRepository) GetByChannelID(channelID int64) (*entity.Scheduler, error) {
	scheduler := &entity.Scheduler{}
	query := `
		SELECT id, channel_id, notification_time, active_days, is_enabled, role, created_at, updated_at
		FROM scheduler_configs
		WHERE channel_id = ?
	`

	var activeDaysJSON string
	err := r.db.QueryRow(query, channelID).Scan(
		&scheduler.ID,
		&scheduler.ChannelID,
		&scheduler.NotificationTime,
		&activeDaysJSON,
		&scheduler.IsEnabled,
		&scheduler.Role,
		&scheduler.CreatedAt,
		&scheduler.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduler: %w", err)
	}

	// Convert JSON to ActiveDays slice
	if err := json.Unmarshal([]byte(activeDaysJSON), &scheduler.ActiveDays); err != nil {
		return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
	}

	return scheduler, nil
}

func (r *schedulerRepository) Update(scheduler *entity.Scheduler) error {
	query := `
		UPDATE scheduler_configs SET
			notification_time = ?,
			active_days = ?,
			is_enabled = ?,
			role = ?,
			updated_at = ?
		WHERE channel_id = ?
	`

	// Convert ActiveDays to JSON for storage
	activeDaysJSON, err := json.Marshal(scheduler.ActiveDays)
	if err != nil {
		return fmt.Errorf("failed to marshal active days: %w", err)
	}

	_, err = r.db.Exec(query,
		scheduler.NotificationTime,
		string(activeDaysJSON),
		scheduler.IsEnabled,
		scheduler.Role,
		time.Now(),
		scheduler.ChannelID,
	)
	if err != nil {
		return fmt.Errorf("failed to update scheduler: %w", err)
	}

	return nil
}

func (r *schedulerRepository) Delete(channelID int64) error {
	query := `DELETE FROM scheduler_configs WHERE channel_id = ?`

	_, err := r.db.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduler: %w", err)
	}

	return nil
}

func (r *schedulerRepository) GetEnabled() ([]*entity.Scheduler, error) {
	query := `
		SELECT id, channel_id, notification_time, active_days, is_enabled, role, created_at, updated_at
		FROM scheduler_configs
		WHERE is_enabled = 1
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled schedulers: %w", err)
	}
	defer rows.Close()

	var schedulers []*entity.Scheduler
	for rows.Next() {
		scheduler := &entity.Scheduler{}
		var activeDaysJSON string
		err := rows.Scan(
			&scheduler.ID,
			&scheduler.ChannelID,
			&scheduler.NotificationTime,
			&activeDaysJSON,
			&scheduler.IsEnabled,
			&scheduler.Role,
			&scheduler.CreatedAt,
			&scheduler.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduler: %w", err)
		}

		// Convert JSON to ActiveDays slice
		if err := json.Unmarshal([]byte(activeDaysJSON), &scheduler.ActiveDays); err != nil {
			return nil, fmt.Errorf("failed to unmarshal active days: %w", err)
		}
		schedulers = append(schedulers, scheduler)
	}

	return schedulers, nil
}

func (r *schedulerRepository) SetEnabled(channelID int64, enabled bool) error {
	query := `
		UPDATE scheduler_configs SET
			is_enabled = ?,
			updated_at = ?
		WHERE channel_id = ?
	`

	_, err := r.db.Exec(query, enabled, time.Now(), channelID)
	if err != nil {
		return fmt.Errorf("failed to set scheduler enabled status: %w", err)
	}

	return nil
}
