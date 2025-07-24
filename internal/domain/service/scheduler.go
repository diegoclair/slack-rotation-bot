package service

import (
	"log"
)

type scheduler struct {
	rotationService *rotationService
	// TODO: Add cron job implementation
}

func newScheduler(rotationService *rotationService) *scheduler {
	return &scheduler{
		rotationService: rotationService,
	}
}

func (s *scheduler) Start() {
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

func (s *scheduler) Stop() {
	log.Println("Scheduler stopped")
	// TODO: Stop all cron jobs
}
