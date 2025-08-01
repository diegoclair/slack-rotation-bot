package database

import (
	"context"
	"fmt"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
)

// instance implements DataManager interface
type instance struct {
	db            *DB
	channelRepo   contract.ChannelRepo
	userRepo      contract.UserRepo
	schedulerRepo contract.SchedulerRepo
}

// NewInstance creates a new database instance with all repositories
func NewInstance(db *DB) contract.DataManager {
	instance := &instance{
		db: db,
	}
	instance.repoInstances()
	return instance
}

// repoInstances initializes all repositories
func (i *instance) repoInstances() {
	i.channelRepo = newChannelRepo(i.db.conn)
	i.userRepo = newUserRepo(i.db.conn)
	i.schedulerRepo = newSchedulerRepo(i.db.conn)
}

// repoInstancesWithConn creates repository instances with custom dbConn
func repoInstancesWithConn(db dbConn) *instance {
	return &instance{
		channelRepo:   newChannelRepo(db),
		userRepo:      newUserRepo(db),
		schedulerRepo: newSchedulerRepo(db),
	}
}

// Channel returns the channel repository
func (i *instance) Channel() contract.ChannelRepo {
	return i.channelRepo
}

// User returns the user repository
func (i *instance) User() contract.UserRepo {
	return i.userRepo
}

// Scheduler returns the scheduler repository
func (i *instance) Scheduler() contract.SchedulerRepo {
	return i.schedulerRepo
}

// WithTransaction executes a function within a database transaction
func (i *instance) WithTransaction(ctx context.Context, fn func(dm contract.DataManager) error) error {
	tx, err := i.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txInstance := repoInstancesWithConn(tx)
	err = fn(txInstance)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v, original error: %w", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}
