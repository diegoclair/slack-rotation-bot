package test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/diegoclair/slack-rotation-bot/internal/handlers"
	"github.com/diegoclair/slack-rotation-bot/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type ServiceMocks struct {
	RotationServiceMock *mocks.MockRotationService
	SlackClientMock     *mocks.MockSlackClient
}

func GetHandlerTest(t *testing.T) (m ServiceMocks, handler *handlers.SlackHandler, ctrl *gomock.Controller) {
	t.Helper()

	ctrl = gomock.NewController(t)
	m = ServiceMocks{
		RotationServiceMock: mocks.NewMockRotationService(ctrl),
		SlackClientMock:     mocks.NewMockSlackClient(ctrl),
	}

	signingSecret := "test-signing-secret"
	handler = handlers.New(m.SlackClientMock, m.RotationServiceMock, signingSecret)

	return
}

// CreateSlackRequest creates a properly signed Slack slash command request
func CreateSlackRequest(t *testing.T, command, text, channelID, channelName, userID, teamID, signingSecret string) *http.Request {
	t.Helper()

	// Create form data matching Slack's slash command format
	form := url.Values{
		"token":        {"test-token"},
		"team_id":      {teamID},
		"team_domain":  {"test-team"},
		"channel_id":   {channelID},
		"channel_name": {channelName},
		"user_id":      {userID},
		"user_name":    {"test-user"},
		"command":      {command},
		"text":         {text},
		"response_url": {"https://hooks.slack.com/commands/test"},
		"trigger_id":   {"test-trigger-id"},
	}

	body := form.Encode()

	req, err := http.NewRequest(http.MethodPost, "/slack/commands", strings.NewReader(body))
	require.NoError(t, err)

	// Set content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Generate Slack signature
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	sig := generateSlackSignature(signingSecret, timestamp, body)
	req.Header.Set("X-Slack-Signature", sig)

	return req
}

func generateSlackSignature(signingSecret, timestamp, body string) string {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(signingSecret))
	h.Write([]byte(baseString))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("v0=%s", signature)
}

func CreateTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}