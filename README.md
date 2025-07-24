# Slack Rotation Bot

Bot to manage people rotation in different Slack teams/channels. Useful for dailies, presentations, code reviews, or any activity that requires automatic rotation.

## Features

- Independent configuration per channel
- Automatic people rotation
- Programmable notifications (daily or other intervals)
- Team member management
- Flexible for any type of rotation (dailies, presentations, reviews, etc.)

## Slack Commands

> **Note**: The bot configures itself automatically on first use. No initial setup command required.

### Manage Members
```bash
/rotation add @user         # Add member to rotation
/rotation remove @user      # Remove member from rotation
/rotation list              # List all active members in rotation
```

### Configuration
```bash
/rotation config time 09:30                    # Set daily notification time
/rotation config days 1,2,4,5                  # Set active days (1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun)
/rotation config show                          # Show current channel settings
```

### Rotation
```bash
/rotation next              # Force advance to next person
```

### Control and Monitoring
```bash
/rotation pause             # Pause automatic notifications temporarily
/rotation resume            # Resume automatic notifications
/rotation status            # Show general status: settings, members and next person
/rotation help              # Show all available commands
```

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

## Installation

```bash
# Clone the repository
git clone https://github.com/diegoclair/slack-rotation-bot

# Install dependencies
go mod download

# Configure environment variables
cp .env.example .env
# Edit .env with your Slack credentials

# Run
go run cmd/bot/main.go
```

## Slack Configuration

### Step 1: Create Slack App
1. **Access**: [api.slack.com/apps](https://api.slack.com/apps)
2. **Click**: green **"Create New App"** button
3. **Select**: **"From scratch"**
4. **Fill in**:
   - **App Name**: `People Rotation Bot` (or your preferred name)
   - **Pick a workspace**: Select your Slack workspace
5. **Click**: **"Create App"**

### Step 2: Configure Bot Permissions
1. **In the left sidebar**, click **"OAuth & Permissions"**
2. **Scroll to**: **"Scopes"** section
3. **Under "Bot Token Scopes"**, click **"Add an OAuth Scope"** and add each:
   - `chat:write` - To send messages to channels
   - `commands` - To receive slash commands  
   - `channels:read` - To read channel information
   - `users:read` - To read user information

### Step 3: Install Bot to Workspace
1. **Still on "OAuth & Permissions" page**, scroll to top
2. **Click**: **"Install to Workspace"** button
3. **Authorize**: permissions on the screen that opens
4. **IMPORTANT**: After installation, **copy the "Bot User OAuth Token"** 
   - Starts with `xoxb-...`
   - You'll need it in the `.env` file

### Step 4: Get Signing Secret
1. **In the sidebar**, click **"Basic Information"**
2. **Scroll to**: **"App Credentials"** section
3. **Click**: **"Show"** next to **"Signing Secret"**
4. **Copy**: the secret (you'll need it in the `.env` file)

### Step 5: Configure Slash Command
1. **In the sidebar**, click **"Slash Commands"**
2. **Click**: **"Create New Command"**
3. **Fill in the fields**:
   - **Command**: `/rotation`
   - **Request URL**: `https://your-server.com/slack/commands` 
     - ⚠️ **For local development**: Use ngrok (see next step)
   - **Short Description**: `Manage people rotation in the team`
   - **Usage Hint**: `add @user | list | config time 09:30`
4. **Click**: **"Save"**

### Step 6: Configure Webhook for Local Development

**6.1. Install ngrok:**
- Download at: [ngrok.com/download](https://ngrok.com/download)
- Or via package manager: `brew install ngrok` (Mac) / `choco install ngrok` (Windows)

**6.2. Run application and ngrok:**
```bash
# Terminal 1: Run Go application
go run cmd/bot/main.go

# Terminal 2: Expose localhost via ngrok  
ngrok http 3000
```

**6.3. Update URL in Slack:**
1. **Copy** the ngrok URL (e.g., `https://abc123.ngrok.io`)
2. **Go back** to **"Slash Commands"** in Slack App
3. **Click** on the `/rotation` command to edit it
4. **Update Request URL** to: `https://abc123.ngrok.io/slack/commands`
5. **Save**

### Step 7: Configure Environment Variables

**Create `.env` file** at project root:
```bash
SLACK_BOT_TOKEN=xoxb-your-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here
PORT=3000
DATABASE_PATH=./rotation.db
```

**Replace with actual values:**
- `SLACK_BOT_TOKEN`: Token copied in Step 3
- `SLACK_SIGNING_SECRET`: Secret copied in Step 4

## How to Test

### Basic Test
```bash
# Check if application is running
curl http://localhost:3000/health  # Should return "OK"
```

### Test in Slack
After configuration, test in Slack channel:
```bash
/rotation add @your-user       # Add yourself to rotation
/rotation list                 # List members
/rotation config time 09:30    # Set time (for dailies or other schedule)
/rotation config days 1,2,4,5  # Set active days (Mon-Tue-Thu-Fri)
/rotation status               # View settings
```

### Usage Examples
```bash
# For daily standup (Monday to Friday)
/rotation config time 09:00
/rotation config days 1,2,3,4,5

# For weekly presentations (Friday)
/rotation config time 14:00
/rotation config days 5

# For code reviews (Monday, Wednesday, Friday)
/rotation config time 10:30
/rotation config days 1,3,5
```