package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/entity"
	"github.com/slack-go/slack"
)

type NotificationEvent struct {
	ChannelID int64
	Time      time.Time
}

type scheduler struct {
	dm            contract.DataManager
	slackClient   *slack.Client
	configChanged chan struct{}
	stopChan      chan struct{}
	running       bool
}

func newScheduler(dm contract.DataManager, slackClient *slack.Client) *scheduler {
	return &scheduler{
		dm:            dm,
		slackClient:   slackClient,
		configChanged: make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
		running:       false,
	}
}

func (s *scheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	log.Println("Scheduler starting...")
	go s.mainLoop()
}

func (s *scheduler) Stop() {
	if !s.running {
		return
	}
	log.Println("Scheduler stopping...")
	close(s.stopChan)
	s.running = false
}

func (s *scheduler) NotifyConfigChange() {
	// Non-blocking send to config change channel
	select {
	case s.configChanged <- struct{}{}:
	default:
		// Channel is full, scheduler will recalculate eventually
	}
}

func (s *scheduler) mainLoop() {
	for {
		nextTime, channelIDs := s.findNextNotification()

		if len(channelIDs) == 0 {
			// No active non-paused channels - wait 1 hour and check again
			log.Println("No active channels found, waiting 1 hour...")
			timer := time.NewTimer(1 * time.Hour)
			select {
			case <-timer.C:
				continue
			case <-s.configChanged:
				timer.Stop()
				continue
			case <-s.stopChan:
				timer.Stop()
				return
			}
		}

		log.Printf("Next notification at %s for %d channels", nextTime.Format("2006-01-02 15:04:05 UTC"), len(channelIDs))

		waitDuration := time.Until(nextTime)
		if waitDuration <= 0 {
			// Time has already passed, send notifications immediately
			s.sendNotifications(channelIDs)
			// Wait 1 minute to prevent re-processing the same time
			log.Println("Sent notifications, waiting 1 minute to prevent re-processing...")
			time.Sleep(1 * time.Minute)
			continue
		}

		timer := time.NewTimer(waitDuration)

		select {
		case <-timer.C:
			// Time to send notifications
			s.sendNotifications(channelIDs)
			// Wait 1 minute to prevent re-processing the same time
			log.Println("Sent notifications, waiting 1 minute to prevent re-processing...")
			time.Sleep(1 * time.Minute)

		case <-s.configChanged:
			// Configuration changed, recalculate
			timer.Stop()
			log.Println("Configuration changed, recalculating schedule...")
			continue

		case <-s.stopChan:
			timer.Stop()
			return
		}
	}
}

func (s *scheduler) findNextNotification() (time.Time, []int64) {
	schedulers, err := s.dm.Scheduler().GetEnabled()
	if err != nil {
		log.Printf("Error getting active channels: %v", err)
		return time.Time{}, nil
	}

	if len(schedulers) == 0 {
		return time.Time{}, nil
	}

	now := time.Now().UTC()

	type channelNext struct {
		channelID int64
		nextTime  time.Time
	}

	var allNext []channelNext

	for _, scheduler := range schedulers {
		nextTime := s.calculateNextForScheduler(scheduler, now)
		if !nextTime.IsZero() {
			allNext = append(allNext, channelNext{
				channelID: scheduler.ChannelID,
				nextTime:  nextTime,
			})
		}
	}

	if len(allNext) == 0 {
		return time.Time{}, nil
	}

	// Sort by time
	sort.Slice(allNext, func(i, j int) bool {
		return allNext[i].nextTime.Before(allNext[j].nextTime)
	})

	// Get earliest time
	earliestTime := allNext[0].nextTime

	// Collect all channels at the earliest time
	var channelIDs []int64
	for _, cn := range allNext {
		if cn.nextTime.Equal(earliestTime) {
			channelIDs = append(channelIDs, cn.channelID)
		} else {
			break // Since it's sorted, we can break early
		}
	}

	return earliestTime, channelIDs
}

func (s *scheduler) calculateNextForScheduler(scheduler *entity.Scheduler, now time.Time) time.Time {
	// Parse notification time
	parts := strings.Split(scheduler.NotificationTime, ":")
	if len(parts) != 2 {
		log.Printf("Invalid notification time format for scheduler %d: %s", scheduler.ID, scheduler.NotificationTime)
		return time.Time{}
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Printf("Invalid hour in notification time for scheduler %d: %s", scheduler.ID, parts[0])
		return time.Time{}
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Invalid minute in notification time for scheduler %d: %s", scheduler.ID, parts[1])
		return time.Time{}
	}

	// Check if any active days are configured
	if len(scheduler.ActiveDays) == 0 {
		log.Printf("No active days configured for scheduler %d", scheduler.ID)
		return time.Time{}
	}

	// Convert active days to map for faster lookup
	activeDaysMap := make(map[int]bool)
	for _, day := range scheduler.ActiveDays {
		activeDaysMap[day] = true
	}

	// Try today first
	today := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	todayWeekday := int(today.Weekday())
	if todayWeekday == 0 { // Sunday = 0 in Go, but we want 7 for ISO 8601
		todayWeekday = 7
	}

	// If today is an active day and the time hasn't passed yet
	if activeDaysMap[todayWeekday] && today.After(now) {
		return today
	}

	// Find next active day
	for i := 1; i <= 7; i++ {
		nextDay := today.AddDate(0, 0, i)
		nextWeekday := int(nextDay.Weekday())
		if nextWeekday == 0 {
			nextWeekday = 7
		}

		if activeDaysMap[nextWeekday] {
			return nextDay
		}
	}

	// Should never reach here if there's at least one active day
	log.Printf("Could not find next notification time for scheduler %d", scheduler.ID)
	return time.Time{}
}

func (s *scheduler) sendNotifications(channelIDs []int64) {
	log.Printf("Sending notifications to %d channels", len(channelIDs))

	for _, channelID := range channelIDs {
		go func(cID int64) {
			if err := s.sendNotificationToChannel(cID); err != nil {
				log.Printf("Failed to send notification to channel %d: %v", cID, err)
			}
		}(channelID)
	}
}

func (s *scheduler) sendNotificationToChannel(channelID int64) error {
	// Get channel info
	channel, err := s.dm.Channel().GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	if channel == nil {
		return fmt.Errorf("channel not found")
	}

	// Get scheduler info for role
	schedulerConfig, err := s.dm.Scheduler().GetByChannelID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get scheduler config: %w", err)
	}

	// Default role if no scheduler config
	role := "On duty"
	if schedulerConfig != nil && schedulerConfig.Role != "" {
		role = schedulerConfig.Role
	}

	// Get next presenter
	nextUser, err := s.getNextPresenter(channelID)
	if err != nil {
		return fmt.Errorf("failed to get next presenter: %w", err)
	}

	if nextUser == nil {
		// No users in rotation, send a message about it
		message := "🤖 *Rotation Reminder*\n\nNo users found in rotation. Use `/rotation add @user` to add team members!"

		_, _, err = s.slackClient.PostMessage(
			channel.SlackChannelID,
			slack.MsgOptionText(message, false),
			slack.MsgOptionAsUser(false),
		)
		return err
	}

	// Record the presentation
	if err := s.recordPresentation(channelID, nextUser.ID); err != nil {
		log.Printf("Failed to record presentation for channel %d, user %d: %v", channelID, nextUser.ID, err)
		// Continue anyway, better to send notification than fail completely
	}

	// Send notification with configurable role
	message := fmt.Sprintf("🎯 *Rotation Reminder*\n\n%s today: <@%s>\n\nUse `/rotation next` to skip to the next person if needed.", role, nextUser.SlackUserID)

	_, _, err = s.slackClient.PostMessage(
		channel.SlackChannelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(false),
	)

	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}

	log.Printf("Notification sent to channel %s for user %s", channel.SlackChannelID, nextUser.SlackUserID)
	return nil
}

func (s *scheduler) getNextPresenter(channelID int64) (*entity.User, error) {
	users, err := s.dm.User().GetActiveUsersByChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	// Find current presenter (last_presenter = true)
	var currentPresenterIndex = -1
	for i, user := range users {
		if user.LastPresenter {
			currentPresenterIndex = i
			break
		}
	}

	// Calculate next presenter index
	nextIndex := (currentPresenterIndex + 1) % len(users)
	return users[nextIndex], nil
}

func (s *scheduler) recordPresentation(channelID, userID int64) error {
	return s.dm.WithTransaction(context.Background(), func(tx contract.DataManager) error {
		// Clear all last_presenter flags for this channel
		if err := tx.User().ClearLastPresenter(channelID); err != nil {
			return fmt.Errorf("failed to clear last presenter: %w", err)
		}

		// Set the new presenter
		if err := tx.User().SetLastPresenter(userID); err != nil {
			return fmt.Errorf("failed to set last presenter: %w", err)
		}

		return nil
	})
}
