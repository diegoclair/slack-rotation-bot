package rotation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/database"
	"github.com/diegoclair/slack-rotation-bot/pkg/models"
	"github.com/slack-go/slack"
)

type Service struct {
	channelRepo  *database.ChannelRepository
	userRepo     *database.UserRepository
	rotationRepo *database.RotationRepository
	slackClient  *slack.Client
}

func New(db *database.DB, slackClient *slack.Client) *Service {
	return &Service{
		channelRepo:  database.NewChannelRepository(db),
		userRepo:     database.NewUserRepository(db),
		rotationRepo: database.NewRotationRepository(db),
		slackClient:  slackClient,
	}
}

func (s *Service) SetupChannel(slackChannelID, slackChannelName, slackTeamID string) (*models.Channel, error) {
	// Check if channel already exists
	channel, err := s.channelRepo.GetBySlackID(slackChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to check channel: %w", err)
	}

	if channel != nil {
		return channel, nil
	}

	// Create new channel with default settings
	channel = &models.Channel{
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

func (s *Service) AddUser(channelID int, slackUserID string) error {
	// Get user info from Slack
	userInfo, err := s.slackClient.GetUserInfo(slackUserID)
	if err != nil {
		return fmt.Errorf("failed to get user info from Slack: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByChannelAndSlackID(channelID, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		if !existingUser.IsActive {
			// Reactivate user
			return s.userRepo.UpdateActiveStatus(existingUser.ID, true)
		}
		return fmt.Errorf("usuário já está na rotação")
	}

	// Create new user
	displayName := userInfo.Profile.RealName
	if displayName == "" {
		displayName = userInfo.Profile.DisplayName
	}
	if displayName == "" {
		displayName = userInfo.Name
	}

	user := &models.User{
		ChannelID:     channelID,
		SlackUserID:   slackUserID,
		SlackUserName: userInfo.Name,
		DisplayName:   displayName,
		IsActive:      true,
	}

	return s.userRepo.Create(user)
}

func (s *Service) RemoveUser(channelID int, slackUserID string) error {
	user, err := s.userRepo.GetByChannelAndSlackID(channelID, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("usuário não encontrado na rotação")
	}

	return s.userRepo.UpdateActiveStatus(user.ID, false)
}

func (s *Service) ListUsers(channelID int) ([]*models.User, error) {
	return s.userRepo.GetActiveUsersByChannel(channelID)
}

func (s *Service) GetNextPresenter(channelID int) (*models.User, error) {
	// Get all active users
	users, err := s.userRepo.GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("nenhum usuário ativo na rotação")
	}

	// Get last presenter
	lastRotation, err := s.rotationRepo.GetLastPresenterByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last presenter: %w", err)
	}

	// If no one has presented yet, return first user
	if lastRotation == nil {
		return users[0], nil
	}

	// Find the index of last presenter
	lastPresenterIndex := -1
	for i, user := range users {
		if user.ID == lastRotation.UserID {
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

func (s *Service) RecordPresentation(channelID, userID int, wasPresenter bool, skippedReason string) error {
	rotation := &models.Rotation{
		ChannelID:     channelID,
		UserID:        userID,
		PresentedAt:   time.Now(),
		WasPresenter:  wasPresenter,
		SkippedReason: skippedReason,
	}

	return s.rotationRepo.Create(rotation)
}

func (s *Service) GetTodaysPresenter(channelID int) (*models.User, *models.Rotation, error) {
	// TODO CRITICAL: Fix timezone handling - currently using server timezone
	// Should use channel-specific timezone for proper "today" calculation
	today := time.Now()
	rotation, err := s.rotationRepo.GetTodaysPresenter(channelID, today)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get today's rotation: %w", err)
	}

	if rotation == nil {
		return nil, nil, nil
	}

	// Get user info
	users, err := s.userRepo.GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		if user.ID == rotation.UserID {
			return user, rotation, nil
		}
	}

	return nil, rotation, nil
}

func (s *Service) UpdateChannelConfig(channelID int, configType, value string) error {
	channel, err := s.channelRepo.GetBySlackID("")
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	switch configType {
	case "time":
		// Validate time format HH:MM
		if _, err := time.Parse("15:04", value); err != nil {
			return fmt.Errorf("formato de horário inválido. Use HH:MM")
		}
		channel.NotificationTime = value
	case "days":
		// Parse days
		days := parseDays(value)
		if len(days) == 0 {
			return fmt.Errorf("dias inválidos. Use: seg,ter,qui,sex")
		}
		daysJSON, _ := json.Marshal(days)
		channel.ActiveDays = string(daysJSON)
	default:
		return fmt.Errorf("tipo de configuração inválido")
	}

	return s.channelRepo.Update(channel)
}

func parseDays(input string) []string {
	dayMap := map[string]string{
		"seg": "Monday",
		"ter": "Tuesday", 
		"qua": "Wednesday",
		"qui": "Thursday",
		"sex": "Friday",
		"sab": "Saturday",
		"dom": "Sunday",
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

func (s *Service) GetChannelStatus(channelID int) (*models.Channel, error) {
	// This would need adjustment to get by ID instead of SlackID
	// For now, returning error
	return nil, fmt.Errorf("not implemented")
}