package database

import (
	"testing"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newChannelRepo(db.conn)

	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}

	err := repo.Create(channel)
	require.NoError(t, err, "Failed to create channel")

	assert.NotZero(t, channel.ID, "Expected channel ID to be set after creation")
}

func TestChannelRepository_GetBySlackID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newChannelRepo(db.conn)

	// Create a test channel
	original := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}

	err := repo.Create(original)
	require.NoError(t, err, "Failed to create test channel")

	// Test successful retrieval
	found, err := repo.GetBySlackID("C123456789")
	require.NoError(t, err, "Failed to get channel by Slack ID")
	require.NotNil(t, found, "Expected to find channel")

	assert.Equal(t, original.SlackChannelID, found.SlackChannelID)
	assert.Equal(t, original.SlackChannelName, found.SlackChannelName)
	assert.Equal(t, original.SlackTeamID, found.SlackTeamID)

	// Test not found
	notFound, err := repo.GetBySlackID("NONEXISTENT")
	require.NoError(t, err, "Unexpected error when channel not found")
	assert.Nil(t, notFound, "Expected nil when channel not found")
}

func TestChannelRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newChannelRepo(db.conn)

	// Create a test channel
	original := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}

	err := repo.Create(original)
	require.NoError(t, err, "Failed to create test channel")

	// Test successful retrieval
	found, err := repo.GetByID(original.ID)
	require.NoError(t, err, "Failed to get channel by ID")
	require.NotNil(t, found, "Expected to find channel")

	assert.Equal(t, original.ID, found.ID)
	assert.Equal(t, original.SlackChannelID, found.SlackChannelID)

	// Test not found
	notFound, err := repo.GetByID(99999)
	require.NoError(t, err, "Unexpected error when channel not found")
	assert.Nil(t, notFound, "Expected nil when channel not found")
}

func TestChannelRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newChannelRepo(db.conn)

	// Create a test channel
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}

	err := repo.Create(channel)
	require.NoError(t, err, "Failed to create test channel")

	// Update the channel
	channel.SlackChannelName = "updated-channel"
	channel.IsActive = false

	err = repo.Update(channel)
	require.NoError(t, err, "Failed to update channel")

	// Verify the update
	updated, err := repo.GetByID(channel.ID)
	require.NoError(t, err, "Failed to retrieve updated channel")
	require.NotNil(t, updated, "Expected to find updated channel")

	assert.Equal(t, "updated-channel", updated.SlackChannelName)
	assert.False(t, updated.IsActive)
}

func TestChannelRepository_GetActiveChannels(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := newChannelRepo(db.conn)

	// Create test channels
	channels := []*entity.Channel{
		{
			SlackChannelID:   "C123456789",
			SlackChannelName: "active-channel-1",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		},
		{
			SlackChannelID:   "C987654321",
			SlackChannelName: "active-channel-2",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		},
		{
			SlackChannelID:   "C555555555",
			SlackChannelName: "inactive-channel",
			SlackTeamID:      "T123456789",
			IsActive:         false,
		},
	}

	for _, ch := range channels {
		err := repo.Create(ch)
		require.NoError(t, err, "Failed to create test channel")
	}

	// Get active channels
	activeChannels, err := repo.GetActiveChannels()
	require.NoError(t, err, "Failed to get active channels")

	// Should return only the 2 active channels
	assert.Len(t, activeChannels, 2, "Expected 2 active channels")

	// Verify all returned channels are active
	for _, ch := range activeChannels {
		assert.True(t, ch.IsActive, "Expected all returned channels to be active")
	}
}
