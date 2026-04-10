# Pluma 🪶

Ultra-lightweight contact form API. Route messages to **Telegram**, **Discord**, or both — via a single YAML config. No database required.

## Features

- 🪶 **Tiny** — ~3MB Docker image (scratch + UPX)
- 🔌 **Multi-provider** — Telegram, Discord (easily extensible)
- 🤖 **Multi-route** — N providers × N destinations from one instance
- 🛡️ **Rate limiting** — Per IP, per route, in-memory
- 🐳 **Docker-first** — Scratch-based, production-ready

## Quick Start

### 1. Create your config

```bash
cp .env.example .env
# Edit .env with your real bot tokens and chat IDs
```

### 2. Run with Docker Compose

```bash
docker compose up -d
```

### 3. Send a message

```bash
curl -X POST http://localhost:8080/contact/website \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "message": "Hello, I'\''m interested in your services."
  }'
```

## Configuration

Secrets use `${ENV_VAR}` interpolation — resolved from environment at startup:

```yaml
server:
  port: 8080
  rate_limit: "1/m"       # Global default: 1 request/minute/IP

routes:
  - path: "/contact/website"
    provider: telegram
    telegram:
      bot_token: "${WEBSITE_BOT_TOKEN}"
      chat_id: "${WEBSITE_CHAT_ID}"

  - path: "/contact/discord"
    provider: discord
    discord:
      webhook_url: "${DISCORD_WEBHOOK_URL}"
    rate_limit: "5/h"     # Override per route
```

| Field | Required | Description |
|-------|----------|-------------|
| `server.port` | No | HTTP port (default: `8080`) |
| `server.rate_limit` | No | Global rate limit (default: `1/m`) |
| `routes[].path` | Yes | URL path for this contact endpoint |
| `routes[].provider` | Yes | Provider name: `telegram` or `discord` |
| `routes[].rate_limit` | No | Override global rate limit |

**Telegram config:**

| Field | Required | Description |
|-------|----------|-------------|
| `telegram.bot_token` | Yes | Telegram Bot API token (supports `${ENV}`) |
| `telegram.chat_id` | Yes | Telegram chat/group ID (supports `${ENV}`) |

**Discord config:**

| Field | Required | Description |
|-------|----------|-------------|
| `discord.webhook_url` | Yes | Discord webhook URL (supports `${ENV}`) |

Rate limit format: `N/m` (per minute) or `N/h` (per hour).

## API Reference

### `POST /{route_path}`

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "message": "Hello!",
  "source": "landing-page"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Sender's name |
| `email` | Yes | Sender's email |
| `message` | Yes | Message body |
| `source` | No | Identifier for the origin site/page |

**Responses:**

| Code | Description |
|------|-------------|
| `200` | Message sent successfully |
| `400` | Invalid request body or missing fields |
| `429` | Rate limit exceeded |
| `500` | Server or provider API error |

### `GET /health`

Returns server status and route count.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_PATH` | `/config.yaml` | Path to configuration file |
| `*_BOT_TOKEN` | — | Telegram bot tokens referenced in config |
| `*_CHAT_ID` | — | Telegram chat IDs referenced in config |
| `*_WEBHOOK_URL` | — | Discord webhook URLs referenced in config |

## Development

```bash
# Run locally
CONFIG_PATH=./config.yaml go run .

# Build binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o pluma .
```

## License

MIT
