package service

import (
	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
	"github.com/slack-go/slack"
)

type Instance struct {
	Rotation  *rotationService
	Scheduler *scheduler
}

func NewInstance(dm contract.DataManager, slackClient *slack.Client) *Instance {
	rotationService := newRotation(dm, slackClient)

	return &Instance{
		Rotation:  rotationService,
		Scheduler: newScheduler(rotationService),
	}
}
