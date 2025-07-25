package slack

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CmdAdd    CommandType = "add"
	CmdRemove CommandType = "remove"
	CmdList   CommandType = "list"
	CmdConfig CommandType = "config"
	CmdNext   CommandType = "next"
	CmdPause  CommandType = "pause"
	CmdResume CommandType = "resume"
	CmdStatus CommandType = "status"
	CmdHelp   CommandType = "help"
)

type Command struct {
	Type CommandType
	Args []string
	Raw  string
}

func ParseCommand(text string) (*Command, error) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return &Command{Type: CmdHelp}, nil
	}

	cmd := &Command{
		Raw: text,
	}

	switch parts[0] {
	case "add":
		cmd.Type = CmdAdd
		if len(parts) > 1 {
			cmd.Args = parts[1:]
		}
	case "remove", "rm":
		cmd.Type = CmdRemove
		if len(parts) > 1 {
			cmd.Args = parts[1:]
		}
	case "list", "ls":
		cmd.Type = CmdList
	case "config":
		cmd.Type = CmdConfig
		if len(parts) > 1 {
			cmd.Args = parts[1:]
		}
	case "next":
		cmd.Type = CmdNext
	case "pause":
		cmd.Type = CmdPause
	case "resume":
		cmd.Type = CmdResume
	case "status":
		cmd.Type = CmdStatus
	case "help", "":
		cmd.Type = CmdHelp
	default:
		return nil, fmt.Errorf("unknown command: %s", parts[0])
	}

	return cmd, nil
}

func GetHelpText() string {
	return `*🔄 People Rotation Bot - Commands*

*⚙️ Configuration:*
• ` + "`/rotation config time HH:MM`" + ` - Set daily notification time (24-hour format, UTC)
  _Example: ` + "`/rotation config time 09:30`" + ` for 9:30 AM UTC or ` + "`/rotation config time 14:00`" + ` for 2:00 PM UTC_
  
• ` + "`/rotation config days 1,2,3,4,5`" + ` - Choose active weekdays
  _Days: 1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun_
  _Example: ` + "`/rotation config days 1,3,5`" + ` (Mon, Wed, Fri only)_
  
• ` + "`/rotation config role NAME`" + ` - Set custom role name
  _Example: ` + "`/rotation config role presenter`" + ` → "presenter today: @user"_
  _Default: "On duty" → "On duty today: @user"_
  
• ` + "`/rotation config show`" + ` - Display current channel settings

*👥 Member Management:*
• ` + "`/rotation add @user1 @user2 ...`" + ` - Add one or more users to the rotation
  _Example: ` + "`/rotation add @john.doe @jane.smith`" + `_
  
• ` + "`/rotation remove @user1 @user2 ...`" + ` - Remove one or more users from rotation
  _Example: ` + "`/rotation remove @jane.smith @john.doe`" + `_
  
• ` + "`/rotation list`" + ` - Show all members in rotation order
  _Current person on duty is marked with 👉 and role name_

*🎯 Rotation Control:*
• ` + "`/rotation next`" + ` - Manually skip to next person
  _Use when current person is unavailable (vacation, sick, etc.)_

*⏸️ Notification Control:*
• ` + "`/rotation pause`" + ` - Temporarily stop daily notifications
• ` + "`/rotation resume`" + ` - Restart daily notifications  
• ` + "`/rotation status`" + ` - Check if bot is active & see current settings

💡 *Quick Start:* Just add members with ` + "`/rotation add @user`" + ` and the bot auto-configures with defaults (9 AM, Mon-Fri)`
}
