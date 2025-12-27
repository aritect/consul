# Consul

<p align="center">
  <img src=".github/assets/consul.svg" alt="Consul" width="100%">
</p>

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/b416a29aea444b6096d5c7271b712919)](https://app.codacy.com/gh/aritect/consul/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/aritect/consul)
![Solana](https://img.shields.io/badge/solana-%239945FF.svg?logo=solana&logoColor=white)
![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/aritectlabs)

Your AI-powered community assistant for Telegram.

Consul is an intelligent Telegram bot developed by Aritect, designed to streamline community management and deliver real-time ecosystem updates. Acting as your digital concierge, Consul bridges the gap between on-chain activity and community engagement — keeping your members informed without the noise.

Think of it as your community's AI co-pilot—automating routine tasks, broadcasting critical updates, and providing instant access to ecosystem resources.

**See It in Action** — Check out the live demo in [Aritect Community](https://t.me/+-gbpQ8ooxu5hYTYy).

**Open Source & Self-Hostable** — Consul is fully open source under the MIT license. Deploy it in your own Telegram community, customize it to your needs, and make it truly yours.

**Need Help?** — Join our [official community](https://t.me/+-gbpQ8ooxu5hYTYy) for setup assistance and onboarding support.

**Contributions Welcome** — We're open to issues and pull requests. Found a bug or have a feature idea? Let us know on [GitHub](https://github.com/aritect/consul).

## Naming

The name "Consul" draws from the Roman Republic's highest elected officials—trusted advisors who guided citizens through complex matters of state. Like its historical namesake, Consul serves as a reliable guide through the Aritect ecosystem, providing authoritative information and facilitating community governance.

## Features

### Core Capabilities

- **Ad-Free Experience** — Clean, distraction-free interactions without promotional interruptions.
- **Buy Bot Implementation** — Real-time monitoring and notifications for token purchases on Solana, with intelligent throttling to prevent notification spam.
- **Cross-Platform Retransmission** — Seamlessly broadcast updates from X directly to designated Telegram threads using the `/retransmit` command.
- **Ecosystem Navigation** — Instant access to charts, contract addresses, and platform resources.
- **Context-Aware Summaries** — AI-generated summaries of the last 100 community messages using LLM (Groq/OpenAI), helping members stay informed without scrolling through endless conversations.
- **Enhanced LLM Integrations** — AI-powered `/consul` command for intelligent Q&A with customizable context. Supports direct questions and reply-based interactions for natural conversations.
- **Community Leaderboards** — Gamified ranking system tracking member engagement and contributions. Reply to any message with `/up` to give points, and use `/leaderboard` to see top contributors.

### Upcoming Features
- **Customizable Buy Alerts** — Personalize your buy notifications with custom GIFs to match your community's style.
- **Achievement System** — Unlockable badges and rewards for community milestones, early adopters, and active participants.
- **Referral Tracking** — Built-in referral system with attribution and reward distribution.
- **Airdrop Distribution** — Automated airdrop campaigns based on community ranking and engagement scores.

We're constantly building new features. Stay tuned for announcements.

## Architecture

### System Design

```
┌─────────────────────────────────────────────────────────────┐
│                        Telegram API                         │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────┐
                    │     Client     │
                    │  (telebot.v3)  │
                    └────────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │     Router     │  ◄──── Config (env vars)
                    └────────┬───────┘
                             │
              ┌──────────────┼─────────────┐
              ▼              ▼             ▼
         ┌──────────┐  ┌───────────┐  ┌──────────┐
         │ Commands │  │ Commands  │  │ Commands │
         │ /start   │  │ /website  │  │ /chart   │
         │ /help    │  │ /ca       │  │ /set     │
         │ /id      │  │ /setup    │  │ ...      │
         └────┬─────┘  └────┬──────┘  └────┬─────┘
              │             │              │
              └─────────────┼──────────────┘
                            ▼
                   ┌─────────────────┐
                   │  Context Layer  │
                   │  - SendAnswer   │
                   │  - Logging      │
                   └────────┬────────┘
                            │
              ┌─────────────┼────────────────┐
              ▼             ▼                ▼
        ┌──────────┐   ┌────────────┐  ┌────────────┐
        │  Logger  │   │  Metrics   │  │    Store   │
        │          │   │(Prometheus)│  │  (LevelDB) │
        └──────────┘   └────────────┘  └────────────┘
                            │
                            ▼
                     ┌───────────────┐
                     │ :8080/metrics │
                     └───────────────┘
```

### Components

| Component | Description |
|-----------|-------------|
| **Bot Client** | Telegram API wrapper (gopkg.in/telebot.v3). |
| **Router** | Message routing and command parsing. |
| **Commands** | Business logic for each bot command. |
| **Context** | Request context with helpers (SendAnswer, logging). |
| **Config** | Environment-based configuration. |
| **Store** | LevelDB for persistent data (recipients, messages). |
| **Metrics** | Prometheus metrics endpoint. |
| **Logger** | Structured logging. |

### Data Flow

1. User sends message to Telegram.
2. Bot Client receives update via long polling.
3. Router parses command and extracts arguments.
4. Corresponding command handler executes business logic.
5. Context layer sends formatted response.
6. Metrics and logs are recorded for observability.

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Initialize bot interaction. |
| `/help` | Display available commands. |
| `/id` | Get current chat ID. |
| `/website` | Get website link. |
| `/ca` | Get token contract address. |
| `/chart` | View chart on Dexscreener. |
| `/retransmit` | Broadcast message to all recipients (admin only). |
| `/setup` | Interactive setup wizard (admin only). |
| `/set` | Configure settings (admin only). |
| `/summary` | Generate AI summary of recent chat messages. |
| `/consul` | Ask AI questions about your project. Supports direct questions and reply-based interactions. |
| `/set_llm_context` | Set custom context for AI responses (admin only). |
| `/up` | Give a point to a message author (reply to message). |
| `/leaderboard` | View top community contributors. |

### Per-Community Configuration

Run `/setup` to see the configuration wizard, then use `/set` to configure:

```
/set name Aritect
/set ticker ARITECT
/set description Your project description
/set website_url https://example.com
/set token_address ABC123...
/set axiom_url https://axiom.trade/your_link
/set dex_url https://dexscreener.com/solana/your_pair
```

All settings are stored per-chat, allowing each community to have its own configuration.

### AI Assistant Usage

#### Setting Custom Context (Admin Only)

Configure AI responses with project-specific knowledge. You can provide context in two ways:

**Option 1: Direct text input**
```
/set_llm_context You are an expert in Aritect platform. Aritect is building trust infrastructure for Web3. We provide risk scoring, wallet analytics, and behavioral analysis for businesses — helping lenders, exchanges, payment processors, and token issuers make informed decisions about on-chain counterparties.
```

**Option 2: Upload a .txt file**
1. Attach a `.txt` file to your message
2. Add the command `/set_llm_context` as caption or in the message
3. The bot will read the file content and set it as context

This is useful for large contexts that exceed Telegram's message size limit.

#### Using the Consul AI Assistant

**Direct questions with conversation context:**
```
/consul What role can utility token play in Aritect platform? How can utility token contribute to growth?
```
The bot analyzes the **last 10 messages** in the chat to understand the conversation context before answering.

**Reply-based interaction with focused context:**
1. Reply to any message with `/consul what do you think about this?`.
2. The bot analyzes **±5 messages around the target message** for context.
3. The bot responds with AI-generated insights, also using reply.

**Context analysis features:**
- **Without reply**: Analyzes last 10 messages to understand ongoing discussion.
- **With reply**: Analyzes ±5 messages around the target message for focused context.
- Messages are truncated to 200 characters each to optimize token usage.
- Usernames are preserved for context attribution

The AI uses:
- System prompt (hardcoded, cannot be overridden).
- Custom context (set via `/set_llm_context`).
- Conversation history (automatic, based on recent messages).

## Getting Started

### Prerequisites

- Go 1.24+
- Docker (optional, for containerized deployment).
- Telegram Bot Token (obtain from [@BotFather](https://t.me/BotFather)).
- Helius API key (for Solana RPC, optional for buy bot functionality).
- LLM API key (Groq or OpenAI, optional for `/summary` and `/consul` commands).

### Configuration

Create a `.env` file in the project root:

```env
# Credentials (required)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
MANAGER_ID=your_telegram_chat_id

# Buy bot (optional)
TOKEN_ADDRESS=your_token_contract_address
HELIUS_RPC_URL=https://mainnet.helius-rpc.com/?api-key=your_api_key

# LLM for summaries (optional)
LLM_PROVIDER=groq
LLM_API_KEY=your_groq_or_openai_api_key
LLM_MODEL=meta-llama/llama-4-scout-17b-16e-instruct

# Default values (optional, can be overridden via /set per community)
PROJECT_NAME=Aritect
TOKEN_TICKER=ARITECT
DESCRIPTION=Your project description
WEBSITE_URL=https://example.com
DEX_URL=https://dexscreener.com/solana/your_pair
AXIOM_URL=https://axiom.trade/your_link
```

Settings priority: `/set` command > env variables. Use `/setup` to see current configuration status.

### Local Development

**Clone the repository:**
```bash
git clone https://github.com/aritect/consul.git
cd consul
```

**Install dependencies:**
```bash
go mod download
```

**Run the bot:**
```bash
go run ./cmd/consul-telegram-bot
```

### Containerized Deployment

**Build the image:**
```bash
docker build -t consul-telegram-bot .
```

**Run the container:**
```bash
docker run -d \
  --name consul \
  --env-file .env \
  -v $(pwd)/data:/workspace/data \
  -p 8080:8080 \
  consul-telegram-bot
```

### Production Deployment

For production environments, we recommend:

1. Use Docker with proper resource limits.
2. Mount a persistent volume for `/workspace/data` (LevelDB storage).
3. Set up monitoring via the Prometheus metrics endpoint at `:8080/metrics`.
4. Use a process manager or orchestrator (Docker Compose, Kubernetes, etc.).

## License

The MIT License (MIT)

Copyright (c) 2025 Aritect

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
