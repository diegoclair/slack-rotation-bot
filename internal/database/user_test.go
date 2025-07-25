package database

import (
	"testing"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// First create a channel for the test
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	t.Run("should create user successfully", func(t *testing.T) {
		user := &entity.User{
			ChannelID:       channel.ID,
			SlackUserID:     "U123456789",
			SlackUserName:   "testuser",
			DisplayName:     "Test User",
			IsActive:        true,
			LastPresenter:   false,
		}

		err := userRepo.Create(user)

		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, channel.ID, user.ChannelID)
		assert.Equal(t, "U123456789", user.SlackUserID)
		assert.Equal(t, "testuser", user.SlackUserName)
		assert.Equal(t, "Test User", user.DisplayName)
		assert.True(t, user.IsActive)
		assert.False(t, user.LastPresenter)
	})

	t.Run("should create user with last presenter flag", func(t *testing.T) {
		user := &entity.User{
			ChannelID:       channel.ID,
			SlackUserID:     "U987654321",
			SlackUserName:   "presenter",
			DisplayName:     "Presenter User",
			IsActive:        true,
			LastPresenter:   true,
		}

		err := userRepo.Create(user)

		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.True(t, user.LastPresenter)
	})
}

func TestUserRepo_GetByChannelAndSlackID(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	testUser := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "testuser",
		DisplayName:     "Test User",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(testUser)
	require.NoError(t, err)

	t.Run("should return user when found", func(t *testing.T) {
		user, err := userRepo.GetByChannelAndSlackID(channel.ID, "U123456789")

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.SlackUserID, user.SlackUserID)
		assert.Equal(t, testUser.SlackUserName, user.SlackUserName)
		assert.Equal(t, testUser.DisplayName, user.DisplayName)
		assert.Equal(t, testUser.IsActive, user.IsActive)
		assert.Equal(t, testUser.LastPresenter, user.LastPresenter)
	})

	t.Run("should return nil when user not found", func(t *testing.T) {
		user, err := userRepo.GetByChannelAndSlackID(channel.ID, "U999999999")

		require.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("should return nil when channel not found", func(t *testing.T) {
		user, err := userRepo.GetByChannelAndSlackID(999, "U123456789")

		require.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepo_GetActiveUsersByChannel(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create active users
	activeUser1 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "user1",
		DisplayName:     "User One",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(activeUser1)
	require.NoError(t, err)

	time.Sleep(time.Millisecond) // Ensure different joined_at times

	activeUser2 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U987654321",
		SlackUserName:   "user2",
		DisplayName:     "User Two",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(activeUser2)
	require.NoError(t, err)

	// Create inactive user
	inactiveUser := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U555555555",
		SlackUserName:   "user3",
		DisplayName:     "User Three",
		IsActive:        false,
		LastPresenter:   false,
	}
	err = userRepo.Create(inactiveUser)
	require.NoError(t, err)

	t.Run("should return only active users ordered by joined_at", func(t *testing.T) {
		users, err := userRepo.GetActiveUsersByChannel(channel.ID)

		require.NoError(t, err)
		require.Len(t, users, 2)
		
		// Should be ordered by joined_at ASC
		assert.Equal(t, activeUser1.ID, users[0].ID)
		assert.Equal(t, activeUser2.ID, users[1].ID)
		
		// Verify all returned users are active
		for _, user := range users {
			assert.True(t, user.IsActive)
		}
	})

	t.Run("should return empty slice when no active users", func(t *testing.T) {
		// Create a new channel with no users
		emptyChannel := &entity.Channel{
			SlackChannelID:   "C000000000",
			SlackChannelName: "empty-channel",
			SlackTeamID:      "T123456789",
			IsActive:         true,
		}
		err := channelRepo.Create(emptyChannel)
		require.NoError(t, err)

		users, err := userRepo.GetActiveUsersByChannel(emptyChannel.ID)

		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepo_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	testUser := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "testuser",
		DisplayName:     "Test User",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(testUser)
	require.NoError(t, err)

	t.Run("should delete user successfully", func(t *testing.T) {
		err := userRepo.Delete(testUser.ID)

		require.NoError(t, err)

		// Verify user is deleted
		user, err := userRepo.GetByChannelAndSlackID(channel.ID, "U123456789")
		require.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("should handle deleting non-existent user", func(t *testing.T) {
		err := userRepo.Delete(99999)

		require.NoError(t, err) // SQLite doesn't error on deleting non-existent rows
	})
}

func TestUserRepo_ClearLastPresenter(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create users with last presenter flags
	user1 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "user1",
		DisplayName:     "User One",
		IsActive:        true,
		LastPresenter:   true,
	}
	err = userRepo.Create(user1)
	require.NoError(t, err)

	user2 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U987654321",
		SlackUserName:   "user2",
		DisplayName:     "User Two",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(user2)
	require.NoError(t, err)

	t.Run("should clear all last presenter flags for channel", func(t *testing.T) {
		err := userRepo.ClearLastPresenter(channel.ID)

		require.NoError(t, err)

		// Verify no users have last presenter flag set
		lastPresenter, err := userRepo.GetLastPresenter(channel.ID)
		require.NoError(t, err)
		assert.Nil(t, lastPresenter)

		// Verify user1 no longer has last presenter flag
		updatedUser1, err := userRepo.GetByChannelAndSlackID(channel.ID, "U123456789")
		require.NoError(t, err)
		assert.False(t, updatedUser1.LastPresenter)
	})
}

func TestUserRepo_SetLastPresenter(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	testUser := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "testuser",
		DisplayName:     "Test User",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(testUser)
	require.NoError(t, err)

	t.Run("should set last presenter flag", func(t *testing.T) {
		err := userRepo.SetLastPresenter(testUser.ID)

		require.NoError(t, err)

		// Verify user has last presenter flag set
		updatedUser, err := userRepo.GetByChannelAndSlackID(channel.ID, "U123456789")
		require.NoError(t, err)
		assert.True(t, updatedUser.LastPresenter)
	})

	t.Run("should handle setting flag on non-existent user", func(t *testing.T) {
		err := userRepo.SetLastPresenter(99999)

		require.NoError(t, err) // SQLite doesn't error on updating non-existent rows
	})
}

func TestUserRepo_GetLastPresenter(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	userRepo := newUserRepo(db.conn)

	// Setup test data
	channel := &entity.Channel{
		SlackChannelID:   "C123456789",
		SlackChannelName: "test-channel",
		SlackTeamID:      "T123456789",
		IsActive:         true,
	}
	channelRepo := newChannelRepo(db.conn)
	err := channelRepo.Create(channel)
	require.NoError(t, err)

	// Create users
	user1 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U123456789",
		SlackUserName:   "user1",
		DisplayName:     "User One",
		IsActive:        true,
		LastPresenter:   false,
	}
	err = userRepo.Create(user1)
	require.NoError(t, err)

	user2 := &entity.User{
		ChannelID:       channel.ID,
		SlackUserID:     "U987654321",
		SlackUserName:   "user2",
		DisplayName:     "User Two",
		IsActive:        true,
		LastPresenter:   true,
	}
	err = userRepo.Create(user2)
	require.NoError(t, err)

	t.Run("should return user with last presenter flag", func(t *testing.T) {
		lastPresenter, err := userRepo.GetLastPresenter(channel.ID)

		require.NoError(t, err)
		require.NotNil(t, lastPresenter)
		assert.Equal(t, user2.ID, lastPresenter.ID)
		assert.Equal(t, user2.SlackUserID, lastPresenter.SlackUserID)
		assert.True(t, lastPresenter.LastPresenter)
	})

	t.Run("should return nil when no last presenter", func(t *testing.T) {
		// Clear all last presenter flags
		err := userRepo.ClearLastPresenter(channel.ID)
		require.NoError(t, err)

		lastPresenter, err := userRepo.GetLastPresenter(channel.ID)

		require.NoError(t, err)
		assert.Nil(t, lastPresenter)
	})

	t.Run("should return nil for non-existent channel", func(t *testing.T) {
		lastPresenter, err := userRepo.GetLastPresenter(99999)

		require.NoError(t, err)
		assert.Nil(t, lastPresenter)
	})
}