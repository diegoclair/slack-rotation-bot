# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Slack rotation bot that manages people rotations for multiple teams/channels. Each Slack channel can have its own configuration, members list, and schedule. Perfect for daily standups, presentations, code reviews, or any activity requiring organized rotation.

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
│   ├── database/    # SQLite connection + repositories (channel, user)
│   ├── domain/      # Domain layer (following DDD principles)
│   │   ├── entity/  # Domain entities (Channel, User)
│   │   ├── service/ # Business logic services
│   │   │   ├── service.go    # Service initialization
│   │   │   ├── rotation.go   # Rotation management logic
│   │   │   └── scheduler.go  # Cron-based scheduler
│   │   └── slack/   # Slack command parsing and help text
│   └── handlers/    # HTTP/Slack webhook handlers
├── migrator/sqlite/  # Database migrations with embedded SQL files
└── go.mod           # Dependencies: slack-go/slack, robfig/cron, sqlite3
```

### Application Flow
1. **Startup**: Load config → Initialize SQLite → Run migrations → Start HTTP server + cron scheduler
2. **Slash Commands**: `/slack/commands` endpoint → Parse command → Route to appropriate handler
3. **Daily Notifications**: Cron scheduler → Check active channels → Send rotation reminders
4. **Multi-tenant**: Each Slack channel has independent config and users

### Core Data Models
- **Channel**: Slack channel config (notification time, active days, team ID)
- **User**: Channel-scoped users with Slack user info, active status, and last_presenter flag

### Database Design
- **Optimized Architecture**: Uses SQLite with `github.com/diegoclair/sqlmigrator`
- **Repository Pattern**: Separate repos for channels and users (rotations table removed for optimization)
- **Zero Growth Database**: Uses `last_presenter` boolean flag instead of infinite history table
- **Migrations**: Auto-run on startup from embedded SQL files (`000001_*.sql`, `000002_*.sql` format)
- **Multi-tenant**: Users are scoped to channel_id (each channel independent)

### Slack Integration
- **Authentication**: Bot Token + Signing Secret validation
- **Commands**: All use `/rotation` prefix (add, remove, list, config, next, status, etc.)
- **Required Scopes**: `chat:write`, `commands`, `channels:read`, `users:read`
- **Command Flow**: Parse in `domain/slack/commands.go` → Route in `handlers/slack_handler.go`
- **Notifications**: Sent via `domain/service/rotation.go` using Slack client
- **Internationalization**: All messages and responses are in English for universal compatibility

### Key Dependencies
- `slack-go/slack`: Official Slack API client
- `robfig/cron/v3`: Cron scheduler for daily notifications  
- `mattn/go-sqlite3`: SQLite driver (requires CGO)
- `joho/godotenv`: Environment file loading

## Available Commands

### Configuration Commands
- `/rotation config time HH:MM` - Set notification time (24-hour format, e.g., 09:30)
- `/rotation config days 1,2,4,5` - Set active days using ISO 8601 standard:
  - 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday, 7=Sunday
- `/rotation config show` - Show current channel settings

### Member Management
- `/rotation add @user` - Add member to rotation
- `/rotation remove @user` - Remove member from rotation (hard delete)
- `/rotation list` - List all active members in rotation order

### Rotation Control
- `/rotation next` - Skip to next presenter and record change
- `/rotation pause` - Pause automatic notifications (TODO: not implemented)
- `/rotation resume` - Resume automatic notifications (TODO: not implemented)
- `/rotation status` - Show current bot status and settings (TODO: not implemented)
- `/rotation help` - Show all available commands

### Important Notes
- **Hard Delete**: Removed users are permanently deleted from database (no soft delete)
- **Rotation Logic**: Uses `last_presenter` boolean flag to track current presenter
- **Auto-setup**: Channels are automatically configured on first command use
- **ISO Days**: Uses numbers 1-7 instead of language-specific abbreviations for universal compatibility

## Recent Optimizations (Completed)

### Database Architecture Overhaul
- ✅ **Removed `rotations` table** - Eliminated infinite growth problem
- ✅ **Added `last_presenter` boolean flag** - Simple rotation tracking
- ✅ **Hard delete users** - No soft delete complexity
- ✅ **Zero database growth** - Only active users + 1 presenter flag per channel

### Internationalization
- ✅ **ISO 8601 days format** - Uses numbers 1-7 instead of language-specific abbreviations
- ✅ **English responses** - All commands, errors, and logs in English
- ✅ **Universal compatibility** - Works across different locales

### Code Simplification
- ✅ **Removed unused commands** - Eliminated `setup`, `who`, `history`, `purge`
- ✅ **Simplified rotation logic** - Clean algorithm using boolean flag
- ✅ **Repository cleanup** - Removed `rotation_repo.go`

## TODO: Critical Issues

### Timezone Management
**CRITICAL**: Timezone handling is currently broken and needs immediate attention.

**Current Problems:**
- All `time.Now()` calls use server timezone, not user timezone
- Users expect notifications at their local time but may get server time
- Poor user experience: bot configured for 09:00 may notify at wrong time

**Impact:** Users will receive notifications at incorrect times, breaking the core functionality.

**Solutions to implement:**
1. **Per-channel timezone**: Add timezone field to Channel model
2. **Parse user timezone**: Use Slack API to get user's timezone 
3. **Proper time handling**: Convert all time operations to use channel timezone
4. **Scheduler fixes**: Ensure cron jobs respect timezone when sending notifications

### Missing Features
- **Status command**: Show current rotation status and settings
- **Pause/Resume**: Temporarily disable notifications
- **Config show**: Display current channel configuration
- **Scheduler Implementation**: Daily notification cron job not implemented yet