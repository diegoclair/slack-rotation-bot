package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/database"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/slack-go/slack"
)

type rotationService struct {
	channelRepo *database.ChannelRepository
	userRepo    *database.UserRepository
	slackClient *slack.Client
}

func newRotation(db *database.DB, slackClient *slack.Client) *rotationService {
	return &rotationService{
		channelRepo: database.NewChannelRepository(db),
		userRepo:    database.NewUserRepository(db),
		slackClient: slackClient,
	}
}

func (s *rotationService) SetupChannel(slackChannelID, slackChannelName, slackTeamID string) (*entity.Channel, error) {
	// Check if channel already exists
	channel, err := s.channelRepo.GetBySlackID(slackChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to check channel: %w", err)
	}

	if channel != nil {
		return channel, nil
	}

	// Create new channel with default settings
	channel = &entity.Channel{
		SlackChannelID:   slackChannelID,
		SlackChannelName: slackChannelName,
		SlackTeamID:      slackTeamID,
		NotificationTime: "09:00",
		ActiveDays:       `["Monday","Tuesday","Thursday","Friday"]`,
		IsActive:         true,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

func (s *rotationService) AddUser(channelID int, slackUserID string) error {
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
	existingUser, err := s.userRepo.GetByChannelAndSlackID(channelID, slackUserID)
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

	return s.userRepo.Create(user)
}

func (s *rotationService) RemoveUser(channelID int, slackUserID string) error {
	user, err := s.userRepo.GetByChannelAndSlackID(channelID, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found in rotation")
	}

	return s.userRepo.Delete(user.ID)
}

func (s *rotationService) ListUsers(channelID int) ([]*entity.User, error) {
	return s.userRepo.GetActiveUsersByChannel(channelID)
}

func (s *rotationService) GetNextPresenter(channelID int) (*entity.User, error) {
	// Get all active users ordered by joined_at (rotation order)
	users, err := s.userRepo.GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no active users in rotation")
	}

	// Get last presenter
	lastPresenter, err := s.userRepo.GetLastPresenter(channelID)
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

func (s *rotationService) RecordPresentation(channelID, userID int) error {
	return s.userRepo.SetLastPresenter(channelID, userID)
}

func (s *rotationService) GetCurrentPresenter(channelID int) (*entity.User, error) {
	return s.userRepo.GetLastPresenter(channelID)
}

func (s *rotationService) UpdateChannelConfig(channelID int, configType, value string) error {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	
	if channel == nil {
		return fmt.Errorf("channel not found")
	}

	switch configType {
	case "time":
		// Validate time format HH:MM
		if _, err := time.Parse("15:04", value); err != nil {
			return fmt.Errorf("invalid time format. Use HH:MM (24-hour format). Example: 09:30")
		}
		channel.NotificationTime = value
	case "days":
		// Parse days
		days := parseDays(value)
		if len(days) == 0 {
			return fmt.Errorf("invalid days. Use numbers 1-7 (1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun). Example: 1,2,4,5")
		}
		daysJSON, _ := json.Marshal(days)
		channel.ActiveDays = string(daysJSON)
	default:
		return fmt.Errorf("invalid configuration type. Use 'time' or 'days'")
	}

	return s.channelRepo.Update(channel)
}

func (s *rotationService) GetChannelConfig(channelID int) (*entity.Channel, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}
	
	return channel, nil
}

func parseDays(input string) []string {
	// ISO 8601 standard: 1=Monday, 2=Tuesday, ..., 7=Sunday
	dayMap := map[string]string{
		"1": "Monday",
		"2": "Tuesday",
		"3": "Wednesday",
		"4": "Thursday",
		"5": "Friday",
		"6": "Saturday",
		"7": "Sunday",
	}

	parts := strings.Split(strings.ToLower(input), ",")
	var days []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if day, ok := dayMap[part]; ok {
			days = append(days, day)
		}
	}

	// Sort days in week order
	weekOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	sort.Slice(days, func(i, j int) bool {
		iIndex := indexOf(weekOrder, days[i])
		jIndex := indexOf(weekOrder, days[j])
		return iIndex < jIndex
	})

	return days
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func (s *rotationService) GetChannelStatus(channelID int) (*entity.Channel, error) {
	// This would need adjustment to get by ID instead of SlackID
	// For now, returning error
	return nil, fmt.Errorf("not implemented")
}
