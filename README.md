# Slack Rotation Bot

<div align="center">
  <img src="slack-rotation-bot.png" alt="Slack Rotation Bot" width="600" height="600">
</div>

**Stop the "who's turn is it today?" confusion in your team!**

Slack Rotation Bot automatically manages people rotation for any recurring team activity. No more manual tracking, forgotten turns, or awkward "I think it's your turn" conversations.

## Why Use This Bot?

### Problems It Solves:
- **👥 Manual tracking headaches** - No more spreadsheets or mental notes about whose turn it is
- **⏰ Missed rotations** - Automatic daily reminders so no one forgets their turn
- **🔄 Unfair distribution** - Ensures everyone gets equal participation over time
- **📋 Multiple team chaos** - Each Slack channel manages its own independent rotation
- **🎯 Flexibility needs** - Works for daily standups, weekly demos, code reviews, or any team activity

### Real-World Use Cases:
- **Daily Standups**: Rotate who facilitates the daily standup meeting
- **Code Reviews**: Fairly distribute code review responsibilities
- **Demo Presentations**: Take turns presenting in sprint demos
- **Meeting Notes**: Rotate who takes notes during team meetings
- **On-Call Duties**: Manage first-responder rotations
- **Team Building**: Rotate who organizes team activities

> 💡 **Tip**: Use the bot image above (`slack-rotation-bot.png`) as your Slack app icon when setting up the bot in your company's Slack workspace!

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
/rotation config time 09:30                    # Set notification time
/rotation config days 1,2,4,5                  # Set active days (1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun)
/rotation config role presenter                # Set role name (e.g., presenter, reviewer, facilitator)
/rotation config show                          # Show current channel settings
```

> 💡 **Configuration Details**:
> - **`time`**: Set the notification time in 24-hour format (HH:MM). This is when the bot will send rotation reminders on active days.
> - **`days`**: Configure which days of the week are active using ISO 8601 numbers (1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday, 7=Sunday). Use comma-separated values for multiple days.
> - **`role`**: Customize the role name used in notifications. The bot automatically adds "today" after the role name. Examples: `presenter` → "presenter today", `reviewer` → "reviewer today", `Code reviewer` → "Code reviewer today" (quotes optional for multi-word roles). Default is "On duty" → "On duty today".
> - **`show`**: Display current channel configuration including notification time, active days, role, and channel status.

### Rotation
```bash
/rotation next              # Skip to next person in rotation
```

> 💡 **When to use `/rotation next`**: Use this command when the current presenter is unavailable (vacation, sick leave, day off, meetings, etc.) to manually advance the rotation to the next person.

### Control and Monitoring
```bash
/rotation pause             # Pause automatic notifications temporarily
/rotation resume            # Resume automatic notifications
/rotation status            # Show general status: settings, members and next person
/rotation help              # Show all available commands
```

## Quick Setup

Ready to eliminate rotation confusion in your team? Here's how to get started:

### Prerequisites
- Admin access to your Slack workspace
- A server to host the bot (cloud service, VPS, or local machine)

> 💡 **For developers**: See [DEVELOPMENT.md](DEVELOPMENT.md) for technical setup, architecture details, and local development instructions.

### Step 1: Download and Deploy
1. Download the latest release from [GitHub Releases](https://github.com/diegoclair/slack-rotation-bot/releases)
2. Deploy to your server (Docker, cloud service, or run directly)
3. Ensure the bot is accessible via HTTPS (required by Slack)

### Step 2: Create Slack App
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
   - **Request URL**: `https://your-server.com/slack/commands` (replace with your actual server URL)
   - **Short Description**: `Manage people rotation in the team`
   - **Usage Hint**: `add @user | list | config time 09:30`
4. **Click**: **"Save"**

### Step 6: Configure Environment Variables

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

## Getting Started

Once your bot is configured and running, test it in any Slack channel:

```bash
/rotation add @yourself        # Add yourself to the rotation
/rotation add @teammate        # Add a teammate
/rotation list                 # See who's in the rotation
/rotation status               # Check current settings
```

## Common Usage Patterns

### Daily Standup Facilitator
```bash
/rotation add @alice @bob @charlie @diana
/rotation config time 09:00              # 9:00 AM daily reminder
/rotation config days 1,2,3,4,5          # Monday through Friday
/rotation config role facilitator        # "facilitator today: @alice"
```

### Weekly Demo Presentations
```bash
/rotation add @team-lead @senior-dev @junior-dev
/rotation config time 14:00              # 2:00 PM reminder
/rotation config days 5                  # Fridays only
/rotation config role presenter          # "presenter today: @team-lead"
```

### Code Review Assignments
```bash
/rotation add @reviewer1 @reviewer2 @reviewer3
/rotation config time 10:30              # 10:30 AM reminder
/rotation config days 1,3,5              # Monday, Wednesday, Friday
/rotation config role "Code reviewer"    # "Code reviewer today: @reviewer1"
```

### On-Call Rotation
```bash
/rotation add @oncall-eng1 @oncall-eng2 @oncall-eng3
/rotation config time 08:00              # 8:00 AM handoff
/rotation config days 1                  # Mondays only (weekly rotation)
/rotation config role "On call"          # "On call today: @oncall-eng1"
```

## Support & Contributing

- 📖 **Documentation**: Check [DEVELOPMENT.md](DEVELOPMENT.md) for technical details
- 🐛 **Bug Reports**: [Open an issue](https://github.com/diegoclair/slack-rotation-bot/issues)
- 💡 **Feature Requests**: [Start a discussion](https://github.com/diegoclair/slack-rotation-bot/discussions)
- 🛠️ **Contributing**: Pull requests are welcome! See [DEVELOPMENT.md](DEVELOPMENT.md) for setup instructions

## License

This project is open source and available under the [MIT License](LICENSE).