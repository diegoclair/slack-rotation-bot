package service

import (
	"github.com/diegoclair/slack-rotation-bot/internal/domain/contract"
)

type Instance struct {
	Rotation  *rotationService
	Scheduler *scheduler
}

func NewInstance(dm contract.DataManager, slackClient contract.SlackClient) *Instance {
	rotationService := newRotation(dm, slackClient)
	schedulerService := newScheduler(dm, slackClient)

	// Connect services to avoid circular dependency
	rotationService.SetScheduler(schedulerService)

	return &Instance{
		Rotation:  rotationService,
		Scheduler: schedulerService,
	}
}
