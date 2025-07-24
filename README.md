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

### Passo 1: Criar Slack App
1. **Acesse**: [api.slack.com/apps](https://api.slack.com/apps)
2. **Clique**: botão verde **"Create New App"**
3. **Selecione**: **"From scratch"**
4. **Preencha**:
   - **App Name**: `People Rotation Bot` (ou nome de sua preferência)
   - **Pick a workspace**: Selecione seu workspace do Slack
5. **Clique**: **"Create App"**

### Passo 2: Configurar Permissões do Bot
1. **No menu lateral esquerdo**, clique em **"OAuth & Permissions"**
2. **Role até**: seção **"Scopes"**
3. **Em "Bot Token Scopes"**, clique **"Add an OAuth Scope"** e adicione cada um:
   - `chat:write` - Para enviar mensagens nos canais
   - `commands` - Para receber slash commands  
   - `channels:read` - Para ler informações dos canais
   - `users:read` - Para ler informações dos usuários

### Passo 3: Instalar Bot no Workspace
1. **Ainda na página "OAuth & Permissions"**, role para o topo
2. **Clique**: botão **"Install to Workspace"**
3. **Autorize**: as permissões na tela que abrir
4. **IMPORTANTE**: Após instalação, **copie o "Bot User OAuth Token"** 
   - Começa com `xoxb-...`
   - Você precisará dele no arquivo `.env`

### Passo 4: Pegar Signing Secret
1. **No menu lateral**, clique em **"Basic Information"**
2. **Role até**: seção **"App Credentials"**
3. **Clique**: **"Show"** ao lado de **"Signing Secret"**
4. **Copie**: o secret (você precisará no arquivo `.env`)

### Passo 5: Configurar Slash Command
1. **No menu lateral**, clique em **"Slash Commands"**
2. **Clique**: **"Create New Command"**
3. **Preencha os campos**:
   - **Command**: `/rotation`
   - **Request URL**: `https://seu-servidor.com/slack/commands` 
     - ⚠️ **Para desenvolvimento local**: Use ngrok (veja próximo passo)
   - **Short Description**: `Gerenciar rotação de pessoas no time`
   - **Usage Hint**: `add @usuario | list | config time 09:30`
4. **Clique**: **"Save"**

### Passo 6: Configurar Webhook para Desenvolvimento Local

**6.1. Instalar ngrok:**
- Baixe em: [ngrok.com/download](https://ngrok.com/download)
- Ou via package manager: `brew install ngrok` (Mac) / `choco install ngrok` (Windows)

**6.2. Executar aplicação e ngrok:**
```bash
# Terminal 1: Rodar aplicação Go
go run cmd/bot/main.go

# Terminal 2: Expor localhost via ngrok  
ngrok http 3000
```

**6.3. Atualizar URL no Slack:**
1. **Copie** a URL do ngrok (ex: `https://abc123.ngrok.io`)
2. **Volte** para **"Slash Commands"** no Slack App
3. **Clique** no comando `/rotation` para editá-lo
4. **Atualize Request URL** para: `https://abc123.ngrok.io/slack/commands`
5. **Salve**

### Passo 7: Configurar Variáveis de Ambiente

**Crie arquivo `.env`** na raiz do projeto:
```bash
SLACK_BOT_TOKEN=xoxb-seu-token-aqui
SLACK_SIGNING_SECRET=seu-signing-secret-aqui
PORT=3000
DATABASE_PATH=./rotation.db
```

**Substitua pelos valores reais:**
- `SLACK_BOT_TOKEN`: Token copiado no Passo 3
- `SLACK_SIGNING_SECRET`: Secret copiado no Passo 4

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