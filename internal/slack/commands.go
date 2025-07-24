package slack

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CmdAdd        CommandType = "add"
	CmdRemove     CommandType = "remove"
	CmdList       CommandType = "list"
	CmdConfig     CommandType = "config"
	CmdNext       CommandType = "next"
	CmdPause      CommandType = "pause"
	CmdResume     CommandType = "resume"
	CmdStatus     CommandType = "status"
	CmdHelp       CommandType = "help"
)

type Command struct {
	Type   CommandType
	Args   []string
	Raw    string
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
		return nil, fmt.Errorf("comando desconhecido: %s", parts[0])
	}

	return cmd, nil
}

func GetHelpText() string {
	return `*Comandos disponíveis:*

*Configuração:*
• ` + "`/rotation config time HH:MM`" + ` - Define horário da notificação (ex: 09:30)
• ` + "`/rotation config days seg,ter,qui,sex`" + ` - Define dias ativos
• ` + "`/rotation config show`" + ` - Mostra configurações atuais

*Gerenciar Membros:*
• ` + "`/rotation add @usuario`" + ` - Adiciona membro à rotação
• ` + "`/rotation remove @usuario`" + ` - Remove membro da rotação
• ` + "`/rotation list`" + ` - Lista todos os membros

*Rotação:*
• ` + "`/rotation next`" + ` - Pula para próximo apresentador

*Controle:*
• ` + "`/rotation pause`" + ` - Pausa notificações automáticas
• ` + "`/rotation resume`" + ` - Retoma notificações automáticas
• ` + "`/rotation status`" + ` - Mostra status do bot neste canal`
}