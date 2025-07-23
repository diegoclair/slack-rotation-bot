package scheduler

import (
	"log"

	"github.com/diegoclair/slack-rotation-bot/internal/rotation"
)

type Scheduler struct {
	rotationService *rotation.Service
	// TODO: Add cron job implementation
}

func New(rotationService *rotation.Service) *Scheduler {
	return &Scheduler{
		rotationService: rotationService,
	}
}

func (s *Scheduler) Start() {
	log.Println("Scheduler started (not implemented yet)")
	// TODO: Implement cron jobs for daily notifications
	// This will:
	// 1. Check all active channels
	// 2. Check if today is an active day for each channel
	// 3. Send notification at configured time
	// 4. Record the presenter for the day
	//
	// CRITICAL TODO: Fix timezone handling before implementing
	// - Currently using server timezone, need per-channel timezone
	// - Must convert notification times to proper timezone
	// - See CLAUDE.md for detailed timezone fix requirements
}

func (s *Scheduler) Stop() {
	log.Println("Scheduler stopped")
	// TODO: Stop all cron jobs
}