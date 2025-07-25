package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/slack-go/slack"
)

type rotationService struct {
	dm          contract.DataManager
	slackClient *slack.Client
	scheduler   *scheduler
}

func newRotation(dm contract.DataManager, slackClient *slack.Client) *rotationService {
	return &rotationService{
		dm:          dm,
		slackClient: slackClient,
		scheduler:   nil, // Will be set later to avoid circular dependency
	}
}

func (s *rotationService) SetScheduler(scheduler *scheduler) {
	s.scheduler = scheduler
}

func (s *rotationService) SetupChannel(slackChannelID, slackChannelName, slackTeamID string) (*entity.Channel, bool, error) {
	// Check if channel already exists
	channel, err := s.dm.Channel().GetBySlackID(slackChannelID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check channel: %w", err)
	}

	if channel != nil {
		return channel, false, nil // Channel already existed
	}

	// Create new channel with default settings
	channel = &entity.Channel{
		SlackChannelID:   slackChannelID,
		SlackChannelName: slackChannelName,
		SlackTeamID:      slackTeamID,
		IsActive:         true,
	}

	if err := s.dm.Channel().Create(channel); err != nil {
		return nil, false, fmt.Errorf("failed to create channel: %w", err)
	}

	// Create default scheduler config
	scheduler := &entity.Scheduler{
		ChannelID:        channel.ID,
		NotificationTime: "09:00",
		ActiveDays:       domain.DefaultActiveDays, // Monday-Friday in ISO format
		IsEnabled:        true,
		Role:             "On duty", // Default role
	}

	if err := s.dm.Scheduler().Create(scheduler); err != nil {
		return nil, false, fmt.Errorf("failed to create scheduler config: %w", err)
	}

	// Notify scheduler of new channel
	if s.scheduler != nil {
		s.scheduler.NotifyConfigChange()
	}

	return channel, true, nil // Channel was auto-created
}

func (s *rotationService) AddUser(channelID int64, slackUserID string) error {
	log.Printf("DEBUG AddUser: channelID=%d, slackUserID=%s", channelID, slackUserID)

	// Get user info from Slack
	userInfo, err := s.slackClient.GetUserInfo(slackUserID)
	if err != nil {
		log.Printf("ERROR getting user info from Slack API for %s: %v", slackUserID, err)
		return fmt.Errorf("failed to get user info from Slack: %w", err)
	}

	log.Printf("DEBUG: Got user info - Name: %s, DisplayName: %s, RealName: %s",
		userInfo.Name, userInfo.Profile.DisplayName, userInfo.Profile.RealName)

	// Check if user already exists
	existingUser, err := s.dm.User().GetByChannelAndSlackID(channelID, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return fmt.Errorf("user is already in the rotation")
	}

	// Create new user
	displayName := userInfo.Profile.RealName
	if displayName == "" {
		displayName = userInfo.Profile.DisplayName
	}
	if displayName == "" {
		displayName = userInfo.Name
	}

	user := &entity.User{
		ChannelID:     channelID,
		SlackUserID:   slackUserID,
		SlackUserName: userInfo.Name,
		DisplayName:   displayName,
		IsActive:      true,
	}

	return s.dm.User().Create(user)
}

func (s *rotationService) RemoveUser(channelID int64, slackUserID string) error {
	user, err := s.dm.User().GetByChannelAndSlackID(channelID, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found in rotation")
	}

	return s.dm.User().Delete(user.ID)
}

func (s *rotationService) ListUsers(channelID int64) ([]*entity.User, error) {
	return s.dm.User().GetActiveUsersByChannel(channelID)
}

func (s *rotationService) GetNextPresenter(channelID int64) (*entity.User, error) {
	// Get all active users ordered by joined_at (rotation order)
	users, err := s.dm.User().GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no active users in rotation")
	}

	// Get last presenter
	lastPresenter, err := s.dm.User().GetLastPresenter(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last presenter: %w", err)
	}

	// If no one has presented yet, return first user
	if lastPresenter == nil {
		return users[0], nil
	}

	// Find the index of last presenter in the current rotation order
	lastPresenterIndex := -1
	for i, user := range users {
		if user.ID == lastPresenter.ID {
			lastPresenterIndex = i
			break
		}
	}

	// If last presenter not found (maybe removed), start from beginning
	if lastPresenterIndex == -1 {
		return users[0], nil
	}

	// Return next user in rotation
	nextIndex := (lastPresenterIndex + 1) % len(users)
	return users[nextIndex], nil
}

func (s *rotationService) RecordPresentation(ctx context.Context, channelID, userID int64) error {
	return s.dm.WithTransaction(ctx, func(tx contract.DataManager) error {
		// Clear previous presenter
		if err := tx.User().ClearLastPresenter(channelID); err != nil {
			return fmt.Errorf("failed to clear last presenter: %w", err)
		}

		// Set new presenter
		if err := tx.User().SetLastPresenter(userID); err != nil {
			return fmt.Errorf("failed to set last presenter: %w", err)
		}

		return nil
	})
}

func (s *rotationService) GetCurrentPresenter(channelID int64) (*entity.User, error) {
	return s.dm.User().GetLastPresenter(channelID)
}

func (s *rotationService) UpdateChannelConfig(channelID int64, configType, value string) error {
	// Get or create scheduler config
	scheduler, err := s.dm.Scheduler().GetByChannelID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get scheduler config: %w", err)
	}

	if scheduler == nil {
		// Create default scheduler config if it doesn't exist
		scheduler = &entity.Scheduler{
			ChannelID:        channelID,
			NotificationTime: "09:00",
			ActiveDays:       domain.DefaultActiveDays,
			IsEnabled:        true,
			Role:             "On duty", // Default role
		}
		if err := s.dm.Scheduler().Create(scheduler); err != nil {
			return fmt.Errorf("failed to create scheduler config: %w", err)
		}
	}

	switch configType {
	case "time":
		// Validate time format HH:MM
		if _, err := time.Parse("15:04", value); err != nil {
			return fmt.Errorf("invalid time format. Use HH:MM (24-hour format). Example: 09:30")
		}
		scheduler.NotificationTime = value
	case "days":
		// Parse days
		days := parseDays(value)
		if len(days) == 0 {
			return fmt.Errorf("invalid days. Use numbers 1-7 (1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun). Example: 1,2,4,5")
		}
		scheduler.ActiveDays = days
	case "role":
		// Set custom role name
		cleanValue := strings.TrimSpace(value)
		if cleanValue == "" {
			return fmt.Errorf("role cannot be empty. Example: presenter, reviewer, facilitator")
		}
		
		// Remove surrounding quotes if present (both single and double)
		if (strings.HasPrefix(cleanValue, `"`) && strings.HasSuffix(cleanValue, `"`)) ||
		   (strings.HasPrefix(cleanValue, `'`) && strings.HasSuffix(cleanValue, `'`)) {
			cleanValue = cleanValue[1 : len(cleanValue)-1]
		}
		
		// Trim again after removing quotes
		cleanValue = strings.TrimSpace(cleanValue)
		if cleanValue == "" {
			return fmt.Errorf("role cannot be empty. Example: presenter, reviewer, facilitator")
		}
		
		scheduler.Role = cleanValue
	default:
		return fmt.Errorf("invalid configuration type. Use 'time', 'days', or 'role'")
	}

	if err := s.dm.Scheduler().Update(scheduler); err != nil {
		return err
	}

	// Notify scheduler of configuration change
	if s.scheduler != nil {
		s.scheduler.NotifyConfigChange()
	}

	return nil
}

func (s *rotationService) GetChannelConfig(channelID int64) (*entity.Channel, error) {
	channel, err := s.dm.Channel().GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// ActiveDays is automatically loaded by the repository from JSON

	return channel, nil
}

func (s *rotationService) GetSchedulerConfig(channelID int64) (*entity.Scheduler, error) {
	return s.dm.Scheduler().GetByChannelID(channelID)
}

func (s *rotationService) PauseScheduler(channelID int64) error {
	err := s.dm.Scheduler().SetEnabled(channelID, false)
	if err != nil {
		return fmt.Errorf("failed to pause scheduler: %w", err)
	}

	// Notify scheduler of configuration change
	if s.scheduler != nil {
		s.scheduler.NotifyConfigChange()
	}

	return nil
}

func (s *rotationService) ResumeScheduler(channelID int64) error {
	err := s.dm.Scheduler().SetEnabled(channelID, true)
	if err != nil {
		return fmt.Errorf("failed to resume scheduler: %w", err)
	}

	// Notify scheduler of configuration change
	if s.scheduler != nil {
		s.scheduler.NotifyConfigChange()
	}

	return nil
}

func parseDays(input string) []int {
	parts := strings.Split(strings.TrimSpace(input), ",")
	var days []int

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if dayNum, ok := domain.WeekdayNumbers[part]; ok {
			days = append(days, dayNum)
		}
	}

	// Sort days in week order (1-7)
	sort.Ints(days)
	return days
}

// indexOf function removed - no longer needed with int sorting

func (s *rotationService) GetChannelStatus(channelID int) (*entity.Channel, error) {
	// This would need adjustment to get by ID instead of SlackID
	// For now, returning error
	return nil, fmt.Errorf("not implemented")
}
