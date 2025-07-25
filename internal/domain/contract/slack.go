package contract

import "github.com/slack-go/slack"

// SlackClient defines the interface for Slack operations
// This allows mocking in tests while keeping the real implementation simple
type SlackClient interface {
	// GetUserInfo retrieves user information from Slack
	GetUserInfo(userID string) (*slack.User, error)
	
	// PostMessage sends a message to a Slack channel
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}