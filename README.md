# Slack Rotation Bot

Bot para gerenciar rotação de pessoas em diferentes times/canais do Slack. Útil para dailies, apresentações, code reviews, ou qualquer atividade que precise de rotação automática.

## Funcionalidades

- Configuração independente por canal
- Rotação automática de pessoas
- Notificações programáveis (diárias ou em outros intervalos)
- Gerenciamento de membros do time
- Flexível para qualquer tipo de rotação (dailies, apresentações, reviews, etc.)

## Comandos Slack

> **Nota**: O bot se configura automaticamente no primeiro uso. Não é necessário comando de setup inicial.

### Gerenciar Membros
```bash
/rotation add @usuario      # Adiciona um membro à rotação
/rotation remove @usuario   # Remove um membro da rotação
/rotation list              # Lista todos os membros ativos na rotação
```

### Configurações
```bash
/rotation config time 09:30                    # Define horário da notificação diária
/rotation config days seg,ter,qui,sex          # Define quais dias da semana são ativos
/rotation config show                          # Exibe as configurações atuais do canal
```

### Rotação
```bash
/rotation next              # Força avançar para a próxima pessoa
```

### Controle e Monitoramento
```bash
/rotation pause             # Pausa as notificações automáticas temporariamente
/rotation resume            # Reativa as notificações automáticas
/rotation status            # Exibe status geral: configurações, membros e próxima pessoa
/rotation help              # Mostra todos os comandos disponíveis
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

### 1. Criar Slack App
1. Acesse [api.slack.com](https://api.slack.com/apps)
2. Clique em **"Create New App"** → **"From scratch"**
3. Nome: `People Rotation Bot` (ou nome de sua preferência)
4. Selecione seu workspace

### 2. Configurar Bot Token Scopes
1. Vá em **"OAuth & Permissions"** no menu lateral
2. Em **"Scopes"** → **"Bot Token Scopes"**, adicione:
   - `chat:write` - Enviar mensagens
   - `commands` - Receber slash commands
   - `channels:read` - Ler informações do canal
   - `users:read` - Ler informações dos usuários

### 3. Configurar Slash Commands
1. Vá em **"Slash Commands"** no menu lateral
2. Clique **"Create New Command"**
3. Configure:
   - **Command**: `/rotation`
   - **Request URL**: `https://seu-servidor.com/slack/commands`
   - **Short Description**: `Gerenciar rotação de pessoas no time`
   - **Usage Hint**: `add @usuario | list | config time 09:30`

### 4. Instalar no Workspace
1. Vá em **"OAuth & Permissions"**
2. Clique **"Install to Workspace"** 
3. Autorize as permissões
4. **Copie o Bot User OAuth Token** (`xoxb-...`)

### 5. Configurar Webhooks (Desenvolvimento)
Para desenvolvimento local, use [ngrok](https://ngrok.com/):

```bash
# Terminal 1: Rodar aplicação
go run cmd/bot/main.go

# Terminal 2: Expor localhost
ngrok http 3000

# Use a URL do ngrok nos Slash Commands
# Exemplo: https://abc123.ngrok.io/slack/commands
```

### 6. Configurar Variáveis de Ambiente
No arquivo `.env`:
```bash
SLACK_BOT_TOKEN=xoxb-sua-bot-token-aqui
SLACK_SIGNING_SECRET=seu-signing-secret-aqui
PORT=3000
DATABASE_PATH=./rotation.db
```

> **Onde encontrar Signing Secret**: Slack App → **"Basic Information"** → **"App Credentials"**

## Como Testar

### Teste Básico
```bash
# Verificar se aplicação está rodando
curl http://localhost:3000/health  # Deve retornar "OK"
```

### Teste no Slack
Depois de configurado, teste no canal do Slack:
```bash
/rotation add @seu-usuario     # Adiciona você à rotação
/rotation list                 # Lista membros
/rotation config time 09:30    # Define horário (para dailies, ou outro horário)
/rotation config days seg,ter,qui,sex  # Define dias ativos
/rotation status               # Vê configurações
```

### Exemplos de Uso
```bash
# Para daily standup
/rotation config time 09:00
/rotation config days seg,ter,qua,qui,sex

# Para apresentações semanais  
/rotation config time 14:00
/rotation config days sex

# Para code reviews
/rotation config time 10:30
/rotation config days seg,qua,sex
```