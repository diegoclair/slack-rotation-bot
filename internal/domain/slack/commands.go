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
	return `*Available Commands:*

*Configuration:*
• ` + "`/rotation config time HH:MM`" + ` - Set notification time (ex: 09:30)
• ` + "`/rotation config days 1,2,4,5`" + ` - Set active days (1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat, 7=Sun)
• ` + "`/rotation config role NAME`" + ` - Set role name (ex: presenter, "Code reviewer") - quotes optional
• ` + "`/rotation config show`" + ` - Show current settings

*Manage Members:*
• ` + "`/rotation add @user`" + ` - Add member to rotation
• ` + "`/rotation remove @user`" + ` - Remove member from rotation
• ` + "`/rotation list`" + ` - List all members

*Rotation:*
• ` + "`/rotation next`" + ` - Skip to next presenter

*Control:*
• ` + "`/rotation pause`" + ` - Pause automatic notifications
• ` + "`/rotation resume`" + ` - Resume automatic notifications
• ` + "`/rotation status`" + ` - Show bot status for this channel`
}
