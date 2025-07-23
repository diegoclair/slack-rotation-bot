package slack

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CmdSetup      CommandType = "setup"
	CmdAdd        CommandType = "add"
	CmdRemove     CommandType = "remove"
	CmdList       CommandType = "list"
	CmdConfig     CommandType = "config"
	CmdNext       CommandType = "next"
	CmdWho        CommandType = "who"
	CmdHistory    CommandType = "history"
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
	case "setup":
		cmd.Type = CmdSetup
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
	case "who":
		cmd.Type = CmdWho
	case "history":
		cmd.Type = CmdHistory
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

*Setup e Configuração:*
• \`/daily setup\` - Configura o bot para este canal
• \`/daily config time HH:MM\` - Define horário da notificação (ex: 09:30)
• \`/daily config days seg,ter,qui,sex\` - Define dias ativos
• \`/daily config show\` - Mostra configurações atuais

*Gerenciar Membros:*
• \`/daily add @usuario\` - Adiciona membro à rotação
• \`/daily remove @usuario\` - Remove membro da rotação
• \`/daily list\` - Lista todos os membros

*Rotação:*
• \`/daily who\` - Mostra quem apresenta hoje
• \`/daily next\` - Pula para próximo apresentador
• \`/daily history\` - Mostra histórico recente

*Controle:*
• \`/daily pause\` - Pausa notificações automáticas
• \`/daily resume\` - Retoma notificações automáticas
• \`/daily status\` - Mostra status do bot neste canal`
}