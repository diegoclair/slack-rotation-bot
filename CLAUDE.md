# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Slack rotation bot that manages daily standup presenter rotations for multiple teams/channels. Each Slack channel can have its own configuration, members list, and schedule.

## Common Commands

### Go Development
```bash
# Download dependencies
go mod download

# Run the application locally
go run cmd/bot/main.go

# Build the application
go build -o slack-rotation-bot cmd/bot/main.go

# Run tests (note: no tests currently exist)
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint installation)
golangci-lint run

# Update dependencies
go mod tidy

# Run specific package tests
go test ./internal/rotation
go test ./internal/database

# Build for production
CGO_ENABLED=1 go build -ldflags="-s -w" -o slack-rotation-bot cmd/bot/main.go
```

### Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit with your Slack credentials
# Required: SLACK_BOT_TOKEN, SLACK_SIGNING_SECRET
# Optional: DATABASE_PATH (defaults to ./rotation.db), PORT (defaults to 3000)
```

## Architecture

### Project Structure
```
slack-rotation-bot/
├── cmd/bot/          # Main application entry point (main.go)
├── internal/         # Private application code
│   ├── config/      # Environment configuration loading
│   ├── database/    # SQLite connection + repositories (channel, user, rotation)
│   ├── handlers/    # HTTP/Slack webhook handlers
│   ├── rotation/    # Core business logic service
│   ├── scheduler/   # Cron-based daily notification scheduler
│   └── slack/       # Command parsing and help text
├── migrator/sqlite/  # Database migrations with embedded SQL files
├── pkg/models/      # Shared data models (Channel, User, Rotation)
└── go.mod           # Dependencies: slack-go/slack, robfig/cron, sqlite3
```

### Application Flow
1. **Startup**: Load config → Initialize SQLite → Run migrations → Start HTTP server + cron scheduler
2. **Slash Commands**: `/slack/commands` endpoint → Parse command → Route to appropriate handler
3. **Daily Notifications**: Cron scheduler → Check active channels → Send rotation reminders
4. **Multi-tenant**: Each Slack channel has independent config, users, and rotation history

### Core Data Models
- **Channel**: Slack channel config (notification time, active days, team ID)
- **User**: Channel-scoped users with Slack user info and active status  
- **Rotation**: History of who presented when, with skip reasons

### Database Design
- Uses SQLite with `github.com/diegoclair/sqlmigrator`
- Repository pattern: separate repos for channels, users, rotations
- Migrations auto-run on startup from embedded SQL files (`000001_*.sql` format)
- Multi-tenant by channel: users and rotations are scoped to channel_id

### Slack Integration
- **Authentication**: Bot Token + Signing Secret validation
- **Commands**: All use `/daily` prefix (setup, add, remove, config, who, next, etc.)
- **Required Scopes**: `chat:write`, `commands`, `channels:read`, `users:read`
- **Command Flow**: Parse in `slack/commands.go` → Route in `handlers/slack_handler.go`
- **Notifications**: Sent via `rotation/service.go` using Slack client

### Key Dependencies
- `slack-go/slack`: Official Slack API client
- `robfig/cron/v3`: Cron scheduler for daily notifications  
- `mattn/go-sqlite3`: SQLite driver (requires CGO)
- `joho/godotenv`: Environment file loading

## TODO: Critical Issues

### Timezone Management
**CRITICAL**: Timezone handling is currently broken and needs immediate attention.

**Current Problems:**
- `.env.example` has `TZ=America/Sao_Paulo` but it's not used in code
- All `time.Now()` calls use server timezone, not user timezone
- Users expect notifications at their local time (e.g., 9:00 AM Brazil) but may get server time
- Poor user experience: bot configured for 09:00 may notify at wrong time

**Impact:** Users will receive notifications at incorrect times, breaking the core functionality.

**Solutions to implement:**
1. **Per-channel timezone**: Add timezone field to Channel model
2. **Parse user timezone**: Use Slack API to get user's timezone 
3. **Proper time handling**: Convert all time operations to use channel timezone
4. **Scheduler fixes**: Ensure cron jobs respect timezone when sending notifications

**Files to modify:**
- `pkg/models/models.go`: Add timezone field to Channel
- `internal/rotation/service.go`: Update time handling logic
- `internal/scheduler/scheduler.go`: Implement timezone-aware cron jobs
- `migrator/sqlite/sql/`: Add migration for timezone column