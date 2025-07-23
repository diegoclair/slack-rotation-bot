# Slack Rotation Bot

Bot para gerenciar rotação de apresentadores da daily em diferentes times/canais do Slack.

## Funcionalidades

- Configuração independente por canal
- Rotação automática de apresentadores
- Notificações diárias configuráveis
- Gerenciamento de membros do time

## Comandos Slack

### Setup Inicial
```
/daily setup
```
Configura o bot para o canal atual.

### Gerenciar Membros
```
/daily add @usuario
/daily remove @usuario
/daily list
```

### Configurações
```
/daily config time 09:30
/daily config days seg,ter,qui,sex
/daily config show
```

### Rotação
```
/daily next              # Pula para próximo apresentador
/daily who               # Mostra quem apresenta hoje
/daily history           # Mostra histórico recente
```

### Controle
```
/daily pause             # Pausa notificações
/daily resume            # Retoma notificações
/daily status            # Status do canal
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