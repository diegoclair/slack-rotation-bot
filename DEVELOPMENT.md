# Development Guide

This document contains technical information for developers who want to contribute to or understand the Slack Rotation Bot codebase.

## Architecture

### Multi-tenancy per Channel
- Each Slack channel has its own configuration
- Users are managed per channel
- Independent rotation history

### Technologies
- **Language**: Go
- **Database**: SQLite
- **Integration**: Slack API (Slash Commands + Bot)
- **Scheduler**: Internal Cron

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
│   │   └── slack/       # Slack command parsing and help text
│   └── handlers/    # HTTP/Slack webhook handlers
├── migrator/sqlite/  # Database migrations with embedded SQL files
└── go.mod           # Dependencies: slack-go/slack, robfig/cron, sqlite3
```

## Development Setup

### Prerequisites
- Go 1.19+
- SQLite
- ngrok (for local development with Slack)

### Installation

```bash
# Clone the repository
git clone https://github.com/diegoclair/slack-rotation-bot

# Navigate to project directory
cd slack-rotation-bot

# Install dependencies
go mod download

# Configure environment variables
cp .env.example .env
# Edit .env with your Slack credentials
```

### Running Locally

```bash
# Run the application
go run cmd/bot/main.go

# Or build and run
go build -o slack-rotation-bot cmd/bot/main.go
./slack-rotation-bot
```

### Local Development with Slack

For local development, you need to expose your local server to the internet using ngrok:

#### 1. Install ngrok
- Download at: [ngrok.com/download](https://ngrok.com/download)
- Or via package manager: `brew install ngrok` (Mac) / `choco install ngrok` (Windows)

#### 2. Run application and ngrok
```bash
# Terminal 1: Run Go application
go run cmd/bot/main.go

# Terminal 2: Expose localhost via ngrok  
ngrok http 3000
```

#### 3. Update Slack webhook URL
1. Copy the ngrok URL (e.g., `https://abc123.ngrok.io`)
2. Go to your Slack App settings → "Slash Commands"
3. Edit the `/rotation` command
4. Update Request URL to: `https://abc123.ngrok.io/slack/commands`
5. Save

## Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test packages
go test ./internal/database/...
go test ./internal/domain/service/...
go test ./internal/handlers/...
```

### Manual Testing
```bash
# Check if application is running
curl http://localhost:3000/health  # Should return "OK"
```

### Test Commands in Slack
After setting up your Slack app and webhook:

```bash
/rotation add @your-user       # Add yourself to rotation
/rotation list                 # List members
/rotation config time 09:30    # Set time
/rotation config days 1,2,4,5  # Set active days
/rotation status               # View settings
```

## Building for Production

```bash
# Build for current platform
go build -o slack-rotation-bot cmd/bot/main.go

# Build with optimizations
CGO_ENABLED=1 go build -ldflags="-s -w" -o slack-rotation-bot cmd/bot/main.go

# Build for Linux (common for Docker/servers)
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o slack-rotation-bot cmd/bot/main.go
```

## Database

The application uses SQLite with automatic migrations. The database file is created automatically on first run.

### Migrations
- Located in `migrator/sqlite/sql/`
- Run automatically on application startup
- Naming format: `000001_description.sql`, `000002_description.sql`, etc.

## Common Development Commands

```bash
# Download dependencies
go mod download

# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Update dependencies
go mod tidy

# Generate mocks (requires mockgen)
make mocks

# Run specific test suites
make test-db       # Test database layer
make test-service  # Test service layer  
make test-handler  # Test HTTP handlers
```

## Environment Variables

Create a `.env` file with:

```bash
# Required
SLACK_BOT_TOKEN=xoxb-your-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here

# Optional
PORT=3000                    # Default: 3000
DATABASE_PATH=./rotation.db  # Default: ./rotation.db
TZ=America/Sao_Paulo        # Default: UTC
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## Troubleshooting

### Common Issues

**"permission denied" when building:**
- Ensure you have CGO enabled: `CGO_ENABLED=1`

**Slack webhook not working:**
- Check if ngrok is running and URL is updated in Slack
- Verify the signing secret is correct
- Check application logs for errors

**Database errors:**
- Ensure the directory for `DATABASE_PATH` exists
- Check file permissions for SQLite database

**Tests failing:**
- Run `go mod tidy` to ensure dependencies are correct
- Generate fresh mocks with `make mocks`