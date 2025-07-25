package database

import (
	"testing"

	"github.com/diegoclair/slack-rotation-bot/internal/domain"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulerRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)

	// Create a channel first
	channelRepo := newChannelRepo(db.conn)
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	scheduler := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "09:00",
		ActiveDays:       domain.DefaultActiveDays,
		IsEnabled:        true,
		Role:             "presenter",
	}

	err = repo.Create(scheduler)
	require.NoError(t, err, "Failed to create scheduler")

	assert.NotZero(t, scheduler.ID, "Expected scheduler ID to be set after creation")
}

func TestSchedulerRepository_GetByChannelID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)

	// Create a channel first
	channelRepo := newChannelRepo(db.conn)
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create a test scheduler
	original := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "14:30",
		ActiveDays:       []int{1, 3, 5}, // Mon, Wed, Fri
		IsEnabled:        true,
		Role:             "reviewer",
	}

	err = repo.Create(original)
	require.NoError(t, err, "Failed to create test scheduler")

	// Test successful retrieval
	found, err := repo.GetByChannelID(channel.ID)
	require.NoError(t, err, "Failed to get scheduler by channel ID")
	require.NotNil(t, found, "Expected to find scheduler")

	assert.Equal(t, original.ChannelID, found.ChannelID)
	assert.Equal(t, original.NotificationTime, found.NotificationTime)
	assert.Equal(t, original.ActiveDays, found.ActiveDays)
	assert.Equal(t, original.IsEnabled, found.IsEnabled)
	assert.Equal(t, original.Role, found.Role)

	// Test not found
	notFound, err := repo.GetByChannelID(99999)
	require.NoError(t, err, "Unexpected error when scheduler not found")
	assert.Nil(t, notFound, "Expected nil when scheduler not found")
}

func TestSchedulerRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)

	// Create a channel first
	channelRepo := newChannelRepo(db.conn)
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create a test scheduler
	scheduler := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "09:00",
		ActiveDays:       domain.DefaultActiveDays,
		IsEnabled:        true,
		Role:             "presenter",
	}

	err = repo.Create(scheduler)
	require.NoError(t, err, "Failed to create test scheduler")

	// Update the scheduler
	scheduler.NotificationTime = "15:45"
	scheduler.ActiveDays = []int{2, 4} // Tue, Thu
	scheduler.IsEnabled = false
	scheduler.Role = "facilitator"

	err = repo.Update(scheduler)
	require.NoError(t, err, "Failed to update scheduler")

	// Verify the update
	updated, err := repo.GetByChannelID(channel.ID)
	require.NoError(t, err, "Failed to retrieve updated scheduler")
	require.NotNil(t, updated, "Expected to find updated scheduler")

	assert.Equal(t, "15:45", updated.NotificationTime)
	assert.Equal(t, []int{2, 4}, updated.ActiveDays)
	assert.False(t, updated.IsEnabled)
	assert.Equal(t, "facilitator", updated.Role)
}

func TestSchedulerRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)

	// Create a channel first
	channelRepo := newChannelRepo(db.conn)
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create a test scheduler
	scheduler := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "09:00",
		ActiveDays:       domain.DefaultActiveDays,
		IsEnabled:        true,
		Role:             "presenter",
	}

	err = repo.Create(scheduler)
	require.NoError(t, err, "Failed to create test scheduler")

	// Delete the scheduler
	err = repo.Delete(channel.ID)
	require.NoError(t, err, "Failed to delete scheduler")

	// Verify deletion
	deleted, err := repo.GetByChannelID(channel.ID)
	require.NoError(t, err, "Unexpected error when checking deleted scheduler")
	assert.Nil(t, deleted, "Expected scheduler to be deleted")
}

func TestSchedulerRepository_GetEnabled(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)
	channelRepo := newChannelRepo(db.conn)

	// Create test channels
	channels := []*entity.Channel{
		{
			SlackChannelID:   "C123456789",
			SlackChannelName: "channel-1",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		},
		{
			SlackChannelID:   "C987654321",
			SlackChannelName: "channel-2",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		},
		{
			SlackChannelID:   "C555555555",
			SlackChannelName: "channel-3",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		},
	}

	for _, ch := range channels {
		err := channelRepo.Create(ch)
		require.NoError(t, err, "Failed to create test channel")
	}

	// Create test schedulers
	schedulers := []*entity.Scheduler{
		{
			ChannelID:        channels[0].ID,
			NotificationTime: "09:00",
			ActiveDays:       domain.DefaultActiveDays,
			IsEnabled:        true, // Enabled
			Role:             "presenter",
		},
		{
			ChannelID:        channels[1].ID,
			NotificationTime: "10:00",
			ActiveDays:       domain.DefaultActiveDays,
			IsEnabled:        true, // Enabled
			Role:             "reviewer",
		},
		{
			ChannelID:        channels[2].ID,
			NotificationTime: "11:00",
			ActiveDays:       domain.DefaultActiveDays,
			IsEnabled:        false, // Disabled
			Role:             "facilitator",
		},
	}

	for _, s := range schedulers {
		err := repo.Create(s)
		require.NoError(t, err, "Failed to create test scheduler")
	}

	// Get enabled schedulers
	enabledSchedulers, err := repo.GetEnabled()
	require.NoError(t, err, "Failed to get enabled schedulers")

	// Should return only the 2 enabled schedulers
	assert.Len(t, enabledSchedulers, 2, "Expected 2 enabled schedulers")

	// Verify all returned schedulers are enabled
	for _, s := range enabledSchedulers {
		assert.True(t, s.IsEnabled, "Expected all returned schedulers to be enabled")
	}
}

func TestSchedulerRepository_SetEnabled(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newSchedulerRepo(db.conn)

	// Create a channel first
	channelRepo := newChannelRepo(db.conn)
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create a test scheduler
	scheduler := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "09:00",
		ActiveDays:       domain.DefaultActiveDays,
		IsEnabled:        true,
		Role:             "presenter",
	}

	err = repo.Create(scheduler)
	require.NoError(t, err, "Failed to create test scheduler")

	// Disable the scheduler
	err = repo.SetEnabled(channel.ID, false)
	require.NoError(t, err, "Failed to set scheduler disabled")

	// Verify the change
	updated, err := repo.GetByChannelID(channel.ID)
	require.NoError(t, err, "Failed to retrieve updated scheduler")
	require.NotNil(t, updated, "Expected to find updated scheduler")
	assert.False(t, updated.IsEnabled, "Expected scheduler to be disabled")

	// Enable it again
	err = repo.SetEnabled(channel.ID, true)
	require.NoError(t, err, "Failed to set scheduler enabled")

	// Verify the change
	updated, err = repo.GetByChannelID(channel.ID)
	require.NoError(t, err, "Failed to retrieve updated scheduler")
	require.NotNil(t, updated, "Expected to find updated scheduler")
	assert.True(t, updated.IsEnabled, "Expected scheduler to be enabled")
}
