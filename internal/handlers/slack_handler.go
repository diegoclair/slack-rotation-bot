package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/diegoclair/slack-rotation-bot/internal/domain"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	slackcmd "github.com/diegoclair/slack-rotation-bot/internal/domain/slack"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	slackClient     contract.SlackClient
	rotationService contract.RotationService
	signingSecret   string
}

func New(slackClient contract.SlackClient, rotationService contract.RotationService, signingSecret string) *SlackHandler {
	return &SlackHandler{
		slackClient:     slackClient,
		rotationService: rotationService,
		signingSecret:   signingSecret,
	}
}

func (h *SlackHandler) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	// Verify request from Slack
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Verify Slack signature
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		log.Printf("ERROR creating verifier: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if _, err := verifier.Write(body); err != nil {
		log.Printf("ERROR writing to verifier: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := verifier.Ensure(); err != nil {
		log.Printf("ERROR verifying signature: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse command
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Printf("ERROR parsing command: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Received command: %s %s from user: %s in channel: %s", s.Command, s.Text, s.UserID, s.ChannelID)
	log.Printf("DEBUG Slash Command data - TeamID: %s, ResponseURL: %s, TriggerID: %s", s.TeamID, s.ResponseURL, s.TriggerID)

	// Parse our command
	cmd, err := slackcmd.ParseCommand(s.Text)
	if err != nil {
		h.respondWithError(w, err.Error())
		return
	}

	// Handle command
	response := h.handleCommand(r.Context(), cmd, &s)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SlackHandler) handleCommand(ctx context.Context, cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	switch cmd.Type {
	case slackcmd.CmdAdd:
		return h.handleAddUser(ctx, cmd, slashCmd)
	case slackcmd.CmdRemove:
		return h.handleRemoveUser(ctx, cmd, slashCmd)
	case slackcmd.CmdList:
		return h.handleListUsers(ctx, slashCmd)
	case slackcmd.CmdConfig:
		return h.handleConfig(ctx, cmd, slashCmd)
	case slackcmd.CmdNext:
		return h.handleNext(ctx, slashCmd)
	case slackcmd.CmdPause:
		return h.handlePause(slashCmd)
	case slackcmd.CmdResume:
		return h.handleResume(slashCmd)
	case slackcmd.CmdStatus:
		return h.handleStatus(slashCmd)
	case slackcmd.CmdHelp:
		return h.handleHelp()
	default:
		return h.createErrorResponse("Unknown command")
	}
}

func (h *SlackHandler) handleAddUser(ctx context.Context, cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Please mention the user: `/rotation add @user`")
	}

	// Extract user ID from mention <@U12345> or <@U12345|username>
	userMention := cmd.Args[0]
	log.Printf("DEBUG: Raw user mention: %s", userMention)

	userID := strings.TrimSpace(userMention)
	userID = strings.TrimPrefix(userID, "<@")
	userID = strings.TrimSuffix(userID, ">")

	// Handle format <@U12345|username> - take only the ID part
	if idx := strings.Index(userID, "|"); idx != -1 {
		userID = userID[:idx]
	}

	log.Printf("DEBUG: Extracted user ID: %s", userID)

	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		log.Printf("ERROR setting up channel: %v", err)
		return h.createErrorResponse("Error checking channel")
	}

	// Add user
	if err := h.rotationService.AddUser(channel.ID, userID); err != nil {
		log.Printf("ERROR adding user %s: %v", userID, err)
		return h.createErrorResponse(fmt.Sprintf("Error adding user: %v", err))
	}

	// Get user info for display name
	userInfo, err := h.slackClient.GetUserInfo(userID)
	userName := userID // fallback
	if err == nil {
		userName = userInfo.Profile.RealName
		if userName == "" {
			userName = userInfo.Profile.DisplayName
		}
		if userName == "" {
			userName = userInfo.Name
		}
	}

	responseText := feedback + fmt.Sprintf("✅ %s has been added to the rotation!", userName)

	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         responseText,
	}
}

func (h *SlackHandler) handleRemoveUser(ctx context.Context, cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Please mention the user: `/rotation remove @user`")
	}

	// Extract user ID from mention <@U12345> or <@U12345|username>
	userMention := cmd.Args[0]
	userID := strings.TrimSpace(userMention)
	userID = strings.TrimPrefix(userID, "<@")
	userID = strings.TrimSuffix(userID, ">")

	// Handle format <@U12345|username> - take only the ID part
	if idx := strings.Index(userID, "|"); idx != -1 {
		userID = userID[:idx]
	}

	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Remove user
	if err := h.rotationService.RemoveUser(channel.ID, userID); err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error removing user: %v", err))
	}

	// Get user info for display name
	userInfo, err := h.slackClient.GetUserInfo(userID)
	userName := userID // fallback
	if err == nil {
		userName = userInfo.Profile.RealName
		if userName == "" {
			userName = userInfo.Profile.DisplayName
		}
		if userName == "" {
			userName = userInfo.Name
		}
	}

	responseText := feedback + fmt.Sprintf("✅ %s has been removed from the rotation.", userName)
	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         responseText,
	}
}

func (h *SlackHandler) handleListUsers(ctx context.Context, slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Get users
	users, err := h.rotationService.ListUsers(channel.ID)
	if err != nil {
		return h.createErrorResponse("Error listing users")
	}

	if len(users) == 0 {
		return &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Text:         "No users in rotation. Use `/rotation add @user` to add members.",
		}
	}

	var userList strings.Builder
	userList.WriteString(feedback + "*Members in rotation:*\n")
	for i, user := range users {
		userList.WriteString(fmt.Sprintf("%d. %s\n", i+1, user.GetDisplayName()))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         userList.String(),
	}
}

func (h *SlackHandler) handleNext(ctx context.Context, slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Get next presenter
	nextUser, err := h.rotationService.GetNextPresenter(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error determining next presenter: %v", err))
	}

	// Record new presenter
	if err := h.rotationService.RecordPresentation(ctx, channel.ID, nextUser.ID); err != nil {
		return h.createErrorResponse("Error recording new presenter")
	}

	responseText := feedback + fmt.Sprintf("⏭️ Skipping to next presenter: <@%s>", nextUser.SlackUserID)
	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         responseText,
	}
}

func (h *SlackHandler) handleConfig(ctx context.Context, cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Use: `/rotation config time HH:MM` or `/rotation config days 1,2,4,5`")
	}

	if cmd.Args[0] == "show" {
		// Get channel with feedback
		channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
		if err != nil {
			return h.createErrorResponse("Error checking channel")
		}

		// Get current configuration
		config, err := h.rotationService.GetChannelConfig(channel.ID)
		if err != nil {
			return h.createErrorResponse(fmt.Sprintf("Error getting configuration: %v", err))
		}

		// Get scheduler configuration
		scheduler, err := h.rotationService.GetSchedulerConfig(channel.ID)
		if err != nil {
			return h.createErrorResponse(fmt.Sprintf("Error getting scheduler configuration: %v", err))
		}

		// Default values if scheduler doesn't exist
		notificationTime := "09:00"
		activeDays := domain.DefaultActiveDays
		isEnabled := true

		if scheduler != nil {
			notificationTime = scheduler.NotificationTime
			activeDays = scheduler.ActiveDays
			isEnabled = scheduler.IsEnabled
		}

		// Convert active days from ISO numbers to names for display
		var activeDaysNames []string
		for _, dayNum := range activeDays {
			if dayName, ok := domain.WeekdayNames[dayNum]; ok {
				activeDaysNames = append(activeDaysNames, dayName)
			}
		}

		configText := feedback + fmt.Sprintf("📋 *Current Configuration for #%s*\n\n"+
			"⏰ *Notification Time:* %s\n"+
			"📅 *Active Days:* %s\n"+
			"🔔 *Channel Status:* %s\n"+
			"📅 *Scheduler Status:* %s",
			config.SlackChannelName,
			notificationTime,
			strings.Join(activeDaysNames, ", "),
			func() string {
				if config.IsActive {
					return "Active"
				}
				return "Inactive"
			}(),
			func() string {
				if isEnabled {
					return "Enabled"
				}
				return "Disabled"
			}(),
		)

		return &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Text:         configText,
		}
	}

	if len(cmd.Args) < 2 {
		return h.createErrorResponse("Invalid format. Use: `/rotation config time HH:MM` or `/rotation config days 1,2,4,5`")
	}

	configType := cmd.Args[0]
	configValue := strings.Join(cmd.Args[1:], " ")

	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	if err := h.rotationService.UpdateChannelConfig(channel.ID, configType, configValue); err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error updating configuration: %v", err))
	}

	responseText := feedback + fmt.Sprintf("✅ Configuration updated: %s = %s", configType, configValue)
	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         responseText,
	}
}

func (h *SlackHandler) handlePause(slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Check if scheduler exists
	scheduler, err := h.rotationService.GetSchedulerConfig(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error getting scheduler configuration: %v", err))
	}

	if scheduler == nil {
		// Create default scheduler config if it doesn't exist
		err = h.rotationService.UpdateChannelConfig(channel.ID, "time", "09:00")
		if err != nil {
			return h.createErrorResponse(fmt.Sprintf("Error creating scheduler configuration: %v", err))
		}
		scheduler, _ = h.rotationService.GetSchedulerConfig(channel.ID)
	}

	// Check if already paused
	if scheduler != nil && !scheduler.IsEnabled {
		return &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Text:         feedback + "⏸️ Notifications are already paused for this channel.",
		}
	}

	// Pause scheduler
	err = h.rotationService.PauseScheduler(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error pausing scheduler: %v", err))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         feedback + "⏸️ Daily rotation notifications have been paused. Use `/rotation resume` to re-enable them.",
	}
}

func (h *SlackHandler) handleResume(slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Check if scheduler exists
	scheduler, err := h.rotationService.GetSchedulerConfig(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error getting scheduler configuration: %v", err))
	}

	if scheduler == nil {
		// Create default scheduler config if it doesn't exist
		err = h.rotationService.UpdateChannelConfig(channel.ID, "time", "09:00")
		if err != nil {
			return h.createErrorResponse(fmt.Sprintf("Error creating scheduler configuration: %v", err))
		}
		scheduler, _ = h.rotationService.GetSchedulerConfig(channel.ID)
	}

	// Check if already enabled
	if scheduler != nil && scheduler.IsEnabled {
		return &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Text:         feedback + "▶️ Notifications are already enabled for this channel.",
		}
	}

	// Resume scheduler
	err = h.rotationService.ResumeScheduler(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error resuming scheduler: %v", err))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         feedback + "▶️ Daily rotation notifications have been resumed.",
	}
}

func (h *SlackHandler) handleStatus(slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel with feedback
	channel, feedback, err := h.setupChannelWithFeedback(slashCmd)
	if err != nil {
		return h.createErrorResponse("Error checking channel")
	}

	// Get channel configuration
	config, err := h.rotationService.GetChannelConfig(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error getting configuration: %v", err))
	}

	// Get scheduler configuration
	scheduler, err := h.rotationService.GetSchedulerConfig(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error getting scheduler configuration: %v", err))
	}

	// Get current presenter
	currentPresenter, err := h.rotationService.GetCurrentPresenter(channel.ID)
	if err != nil && err.Error() != "user not found" {
		return h.createErrorResponse(fmt.Sprintf("Error getting current presenter: %v", err))
	}

	// Get next presenter
	nextPresenter, err := h.rotationService.GetNextPresenter(channel.ID)
	if err != nil && err.Error() != "no active users in rotation" {
		return h.createErrorResponse(fmt.Sprintf("Error getting next presenter: %v", err))
	}

	// Get total users
	users, err := h.rotationService.ListUsers(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Error getting users: %v", err))
	}

	// Build status message
	statusText := feedback + fmt.Sprintf("📊 *Rotation Status for #%s*\n\n", config.SlackChannelName)

	// Channel status
	statusText += fmt.Sprintf("🔔 *Channel Status:* %s\n", func() string {
		if config.IsActive {
			return "Active"
		}
		return "Inactive"
	}())

	// Scheduler status
	if scheduler != nil {
		statusText += fmt.Sprintf("📅 *Scheduler Status:* %s\n", func() string {
			if scheduler.IsEnabled {
				return "Enabled ✅"
			}
			return "Paused ⏸️"
		}())

		// Notification time
		statusText += fmt.Sprintf("⏰ *Notification Time:* %s UTC\n", scheduler.NotificationTime)

		// Active days
		var activeDaysNames []string
		for _, dayNum := range scheduler.ActiveDays {
			if dayName, ok := domain.WeekdayNames[dayNum]; ok {
				activeDaysNames = append(activeDaysNames, dayName)
			}
		}
		statusText += fmt.Sprintf("📅 *Active Days:* %s\n", strings.Join(activeDaysNames, ", "))
		
		// Role
		statusText += fmt.Sprintf("🎭 *Role:* %s\n", scheduler.Role)
	} else {
		statusText += "📅 *Scheduler:* Not configured\n"
	}

	statusText += "\n"

	// Rotation info
	statusText += fmt.Sprintf("👥 *Total Members:* %d\n", len(users))

	// Use role from scheduler or default
	role := "On duty"
	if scheduler != nil && scheduler.Role != "" {
		role = scheduler.Role
	}

	if currentPresenter != nil {
		statusText += fmt.Sprintf("🎯 *Current %s:* <@%s>\n", role, currentPresenter.SlackUserID)
	} else {
		statusText += fmt.Sprintf("🎯 *Current %s:* None\n", role)
	}

	if nextPresenter != nil {
		statusText += fmt.Sprintf("⏭️ *Next %s:* <@%s>\n", role, nextPresenter.SlackUserID)
	} else {
		statusText += fmt.Sprintf("⏭️ *Next %s:* None\n", role)
	}

	// Add help text
	statusText += "\n💡 Use `/rotation help` to see all available commands."

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         statusText,
	}
}

func (h *SlackHandler) handleHelp() *slack.Msg {
	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         slackcmd.GetHelpText(),
	}
}

func (h *SlackHandler) createErrorResponse(message string) *slack.Msg {
	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("❌ %s", message),
	}
}

// setupChannelWithFeedback handles channel setup and provides feedback if auto-configured
func (h *SlackHandler) setupChannelWithFeedback(slashCmd *slack.SlashCommand) (*entity.Channel, string, error) {
	channel, wasCreated, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return nil, "", err
	}

	var feedback string
	if wasCreated {
		feedback = "✅ *Channel configured automatically with default settings:*\n" +
			"⏰ Time: 09:00 | 📅 Days: Mon, Tue, Wed, Thu, Fri\n" +
			"Use `/rotation config show` to view or `/rotation config` to customize.\n\n"
	}

	return channel, feedback, nil
}

func (h *SlackHandler) respondWithError(w http.ResponseWriter, message string) {
	response := h.createErrorResponse(message)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
