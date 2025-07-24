package rotation

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/database"
	"github.com/diegoclair/slack-rotation-bot/pkg/models"
	"github.com/slack-go/slack"
)

type Service struct {
	channelRepo *database.ChannelRepository
	userRepo    *database.UserRepository
	slackClient *slack.Client
}

func New(db *database.DB, slackClient *slack.Client) *Service {
	return &Service{
		channelRepo: database.NewChannelRepository(db),
		userRepo:    database.NewUserRepository(db),
		slackClient: slackClient,
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

	return s.userRepo.Delete(user.ID)
}

func (s *Service) ListUsers(channelID int) ([]*models.User, error) {
	return s.userRepo.GetActiveUsersByChannel(channelID)
}

func (s *Service) GetNextPresenter(channelID int) (*models.User, error) {
	// Get all active users ordered by joined_at (rotation order)
	users, err := s.userRepo.GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("nenhum usuário ativo na rotação")
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

func (s *Service) RecordPresentation(channelID, userID int) error {
	return s.userRepo.SetLastPresenter(channelID, userID)
}

func (s *Service) GetCurrentPresenter(channelID int) (*models.User, error) {
	return s.userRepo.GetLastPresenter(channelID)
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
