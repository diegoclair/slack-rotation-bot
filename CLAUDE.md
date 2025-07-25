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
# or using Makefile
make run

# Build the application
go build -o slack-rotation-bot cmd/bot/main.go
# or using Makefile
make build

# Run all tests with coverage
go test -v -cover ./...
# or using Makefile
make tests

# Run specific test suites using Makefile
make test-db       # Test database layer
make test-service  # Test service layer  
make test-handler  # Test HTTP handlers

# Format code
go fmt ./...

# Lint code (requires golangci-lint installation)
golangci-lint run

# Update dependencies
go mod tidy

# Generate mocks for testing
make mocks         # Generates all mocks (requires mockgen)
make install-mockgen  # Install mockgen if needed

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
│   │   ├── consts.go    # ISO 8601 weekday constants and mappings
│   │   ├── contract/    # Repository interfaces (DataManager pattern)
│   │   ├── entity/      # Domain entities (Channel, User)
│   │   ├── service/     # Business logic services
│   │   │   ├── instance.go  # Service initialization
│   │   │   ├── rotation.go  # Rotation management logic
│   │   │   └── scheduler.go # Cron-based scheduler
│   │   └── slack/       # Slack command parsing and help text
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
- **Channel**: Slack channel basic info (team ID, channel name)
- **Scheduler**: Per-channel config (notification time, active days, role name, enabled/disabled)
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
- `go.uber.org/mock`: Mock generation for testing

### Testing Strategy
- **Mock Generation**: Uses `mockgen` to generate mocks from interfaces in `contract/`
- **Test Utils**: Database test utilities in `internal/database/testutil.go` for in-memory SQLite
- **Handler Tests**: HTTP handler tests with mock Slack client in `internal/handlers/test/`
- **Service Tests**: Business logic tests with mocked repositories
- **Test Pattern**: Table-driven tests with clear test cases and assertions

## Available Commands

### Configuration Commands
- `/rotation config time HH:MM` - Set notification time (24-hour format, e.g., 09:30)
- `/rotation config days 1,2,4,5` - Set active days using ISO 8601 standard:
  - 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday, 7=Sunday
- `/rotation config role NAME` - Set role name (e.g., presenter → "presenter today", "Code reviewer" → "Code reviewer today") - default: "On duty" → "On duty today"
- `/rotation config show` - Show current channel settings

### Member Management
- `/rotation add @user` - Add member to rotation
- `/rotation remove @user` - Remove member from rotation (hard delete)
- `/rotation list` - List all active members in rotation order

### Rotation Control
- `/rotation next` - Skip to next person and record change
- `/rotation pause` - Pause automatic notifications
- `/rotation resume` - Resume automatic notifications  
- `/rotation status` - Show current bot status and settings
- `/rotation help` - Show all available commands

### Important Notes
- **Hard Delete**: Removed users are permanently deleted from database (no soft delete)
- **Rotation Logic**: Uses `last_presenter` boolean flag to track current presenter
- **Auto-setup**: Channels are automatically configured on first command use with user feedback
  - Default settings: 09:00 notification time, Monday-Friday active days
  - Users see confirmation message: "✅ *Channel configured automatically with default settings:*"
  - Provides guidance to customize: "Use `/rotation config show` to view or `/rotation config` to customize"
- **ISO Days**: Uses numbers 1-7 instead of language-specific abbreviations for universal compatibility

## Known Issues & TODOs

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

