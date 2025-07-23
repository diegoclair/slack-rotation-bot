# Slack Rotation Bot

Bot para gerenciar rotação de apresentadores da daily em diferentes times/canais do Slack.

## Funcionalidades

- Configuração independente por canal
- Rotação automática de apresentadores
- Notificações diárias configuráveis
- Gerenciamento de membros do time

## Comandos Slack

> **Nota**: O bot se configura automaticamente no primeiro uso. Não é necessário comando de setup inicial.

### Gerenciar Membros
```bash
/daily add @usuario      # Adiciona um membro à rotação de apresentadores
/daily remove @usuario   # Remove um membro da rotação de apresentadores  
/daily list              # Lista todos os membros ativos na rotação
```

### Configurações
```bash
/daily config time 09:30                    # Define horário da notificação diária
/daily config days seg,ter,qui,sex          # Define quais dias da semana são ativos
/daily config show                          # Exibe as configurações atuais do canal
```

### Rotação
```bash
/daily next              # Força avançar para o próximo apresentador
/daily history           # Mostra o histórico recente de apresentações
```

### Controle e Monitoramento
```bash
/daily pause             # Pausa as notificações automáticas temporariamente
/daily resume            # Reativa as notificações automáticas
/daily status            # Exibe status geral: configurações, membros e próximo apresentador
/daily help              # Mostra todos os comandos disponíveis
```

## Arquitetura

### Multi-tenancy por Canal
- Cada canal Slack tem sua própria configuração
- Usuários são gerenciados por canal
- Histórico de rotação independente

### Tecnologias
- **Linguagem**: Go
- **Banco de Dados**: SQLite
- **Integração**: Slack API (Slash Commands + Bot)
- **Scheduler**: Cron interno

## Instalação

```bash
# Clone o repositório
git clone https://github.com/diegoclair/slack-rotation-bot

# Instale dependências
go mod download

# Configure variáveis de ambiente
cp .env.example .env
# Edite .env com suas credenciais Slack

# Execute
go run cmd/bot/main.go
```

## Configuração no Slack

1. Crie um novo App em api.slack.com
2. Configure Slash Commands apontando para `/slack/commands`
3. Configure Bot Token Scopes:
   - `chat:write`
   - `commands`
   - `channels:read`
   - `users:read`
4. Instale o app no workspace