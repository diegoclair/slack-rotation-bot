package service

import (
	"github.com/diegoclair/slack-rotation-bot/internal/database"
	"github.com/slack-go/slack"
)

type Services struct {
	Rotation  *rotationService
	Scheduler *scheduler
}

func New(db *database.DB, slackClient *slack.Client) *Services {
	rotationService := newRotation(db, slackClient)

	return &Services{
		Rotation:  rotationService,
		Scheduler: newScheduler(rotationService),
	}
}
