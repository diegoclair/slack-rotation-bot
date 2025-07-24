package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/diegoclair/slack-rotation-bot/internal/rotation"
	slackcmd "github.com/diegoclair/slack-rotation-bot/internal/slack"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	slackClient     *slack.Client
	rotationService *rotation.Service
	signingSecret   string
}

func New(slackClient *slack.Client, rotationService *rotation.Service, signingSecret string) *SlackHandler {
	return &SlackHandler{
		slackClient:     slackClient,
		rotationService: rotationService,
		signingSecret:   signingSecret,
	}
}

func (h *SlackHandler) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	// Verify request from Slack
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Verify Slack signature
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if _, err := verifier.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse command
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse our command
	cmd, err := slackcmd.ParseCommand(s.Text)
	if err != nil {
		h.respondWithError(w, err.Error())
		return
	}

	// Handle command
	response := h.handleCommand(cmd, &s)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SlackHandler) handleCommand(cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	switch cmd.Type {
	case slackcmd.CmdAdd:
		return h.handleAddUser(cmd, slashCmd)
	case slackcmd.CmdRemove:
		return h.handleRemoveUser(cmd, slashCmd)
	case slackcmd.CmdList:
		return h.handleListUsers(slashCmd)
	case slackcmd.CmdConfig:
		return h.handleConfig(cmd, slashCmd)
	case slackcmd.CmdNext:
		return h.handleNext(slashCmd)
	case slackcmd.CmdHistory:
		return h.handleHistory(slashCmd)
	case slackcmd.CmdPause:
		return h.handlePause(slashCmd)
	case slackcmd.CmdResume:
		return h.handleResume(slashCmd)
	case slackcmd.CmdStatus:
		return h.handleStatus(slashCmd)
	case slackcmd.CmdHelp:
		return h.handleHelp()
	default:
		return h.createErrorResponse("Comando não reconhecido")
	}
}


func (h *SlackHandler) handleAddUser(cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Por favor, mencione o usuário: `/daily add @usuario`")
	}

	// Extract user ID from mention <@U12345>
	userMention := cmd.Args[0]
	userID := strings.TrimSpace(userMention)
	userID = strings.TrimPrefix(userID, "<@")
	userID = strings.TrimSuffix(userID, ">")

	// Get channel
	channel, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar canal")
	}

	// Add user
	if err := h.rotationService.AddUser(channel.ID, userID); err != nil {
		return h.createErrorResponse(fmt.Sprintf("Erro ao adicionar usuário: %v", err))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fmt.Sprintf("✅ <@%s> foi adicionado à rotação!", userID),
	}
}

func (h *SlackHandler) handleRemoveUser(cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Por favor, mencione o usuário: `/daily remove @usuario`")
	}

	// Extract user ID from mention
	userMention := cmd.Args[0]
	userID := strings.TrimSpace(userMention)
	userID = strings.TrimPrefix(userID, "<@")
	userID = strings.TrimSuffix(userID, ">")

	// Get channel
	channel, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar canal")
	}

	// Remove user
	if err := h.rotationService.RemoveUser(channel.ID, userID); err != nil {
		return h.createErrorResponse(fmt.Sprintf("Erro ao remover usuário: %v", err))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fmt.Sprintf("✅ <@%s> foi removido da rotação.", userID),
	}
}

func (h *SlackHandler) handleListUsers(slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel
	channel, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar canal")
	}

	// Get users
	users, err := h.rotationService.ListUsers(channel.ID)
	if err != nil {
		return h.createErrorResponse("Erro ao listar usuários")
	}

	if len(users) == 0 {
		return &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Text:         "Nenhum usuário na rotação. Use `/rotation add @usuario` para adicionar.",
		}
	}

	var userList strings.Builder
	userList.WriteString("*Membros na rotação:*\n")
	for i, user := range users {
		userList.WriteString(fmt.Sprintf("%d. %s\n", i+1, user.DisplayName))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         userList.String(),
	}
}


func (h *SlackHandler) handleNext(slashCmd *slack.SlashCommand) *slack.Msg {
	// Get channel
	channel, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar canal")
	}

	// Get current presenter
	currentUser, _, err := h.rotationService.GetTodaysPresenter(channel.ID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar apresentador atual")
	}

	// Get next presenter
	nextUser, err := h.rotationService.GetNextPresenter(channel.ID)
	if err != nil {
		return h.createErrorResponse(fmt.Sprintf("Erro ao determinar próximo apresentador: %v", err))
	}

	// Mark current as skipped if exists
	if currentUser != nil {
		if err := h.rotationService.RecordPresentation(channel.ID, currentUser.ID, false, "Pulado manualmente"); err != nil {
			return h.createErrorResponse("Erro ao registrar mudança")
		}
	}

	// Record new presenter
	if err := h.rotationService.RecordPresentation(channel.ID, nextUser.ID, true, ""); err != nil {
		return h.createErrorResponse("Erro ao registrar novo apresentador")
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         fmt.Sprintf("⏭️ Pulando para próximo apresentador: <@%s>", nextUser.SlackUserID),
	}
}

func (h *SlackHandler) handleConfig(cmd *slackcmd.Command, slashCmd *slack.SlashCommand) *slack.Msg {
	if len(cmd.Args) == 0 {
		return h.createErrorResponse("Use: `/daily config time HH:MM` ou `/daily config days seg,ter,qui,sex`")
	}

	if cmd.Args[0] == "show" {
		// TODO: Show current config
		return h.createErrorResponse("Funcionalidade em desenvolvimento")
	}

	if len(cmd.Args) < 2 {
		return h.createErrorResponse("Formato inválido. Use: `/daily config time HH:MM` ou `/daily config days seg,ter,qui,sex`")
	}

	configType := cmd.Args[0]
	configValue := strings.Join(cmd.Args[1:], " ")

	// Get channel
	channel, err := h.rotationService.SetupChannel(slashCmd.ChannelID, slashCmd.ChannelName, slashCmd.TeamID)
	if err != nil {
		return h.createErrorResponse("Erro ao verificar canal")
	}

	if err := h.rotationService.UpdateChannelConfig(channel.ID, configType, configValue); err != nil {
		return h.createErrorResponse(fmt.Sprintf("Erro ao atualizar configuração: %v", err))
	}

	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Configuração atualizada: %s = %s", configType, configValue),
	}
}

func (h *SlackHandler) handleHistory(slashCmd *slack.SlashCommand) *slack.Msg {
	// TODO: Implement history
	return h.createErrorResponse("Funcionalidade em desenvolvimento")
}

func (h *SlackHandler) handlePause(slashCmd *slack.SlashCommand) *slack.Msg {
	// TODO: Implement pause
	return h.createErrorResponse("Funcionalidade em desenvolvimento")
}

func (h *SlackHandler) handleResume(slashCmd *slack.SlashCommand) *slack.Msg {
	// TODO: Implement resume
	return h.createErrorResponse("Funcionalidade em desenvolvimento")
}

func (h *SlackHandler) handleStatus(slashCmd *slack.SlashCommand) *slack.Msg {
	// TODO: Implement status
	return h.createErrorResponse("Funcionalidade em desenvolvimento")
}

func (h *SlackHandler) handleHelp() *slack.Msg {
	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         slackcmd.GetHelpText(),
	}
}

func (h *SlackHandler) createErrorResponse(message string) *slack.Msg {
	return &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("❌ %s", message),
	}
}

func (h *SlackHandler) respondWithError(w http.ResponseWriter, message string) {
	response := h.createErrorResponse(message)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}