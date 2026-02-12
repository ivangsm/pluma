# Pluma 🪶

Ultra-lightweight Telegram contact API. Supports **multiple bots** and **multiple chats** via a single YAML config — no database required.

## Features

- 🪶 **Tiny** — ~3MB Docker image (scratch + UPX)
- 🤖 **Multi-bot** — N bots × N chats from one instance
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
    bot_token: "${WEBSITE_BOT_TOKEN}"
    chat_id: "${WEBSITE_CHAT_ID}"

  - path: "/contact/app"
    bot_token: "${APP_BOT_TOKEN}"
    chat_id: "${APP_CHAT_ID}"
    rate_limit: "5/h"     # Override per route
```

| Field | Required | Description |
|-------|----------|-------------|
| `server.port` | No | HTTP port (default: `8080`) |
| `server.rate_limit` | No | Global rate limit (default: `1/m`) |
| `routes[].path` | Yes | URL path for this contact endpoint |
| `routes[].bot_token` | Yes | Telegram Bot API token (supports `${ENV}`) |
| `routes[].chat_id` | Yes | Telegram chat/group ID (supports `${ENV}`) |
| `routes[].rate_limit` | No | Override global rate limit |

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
| `source` | No | Identifier for the origin site/page (shown in the Telegram message) |

**Responses:**

| Code | Description |
|------|-------------|
| `200` | Message sent successfully |
| `400` | Invalid request body or missing fields |
| `429` | Rate limit exceeded |
| `500` | Server or Telegram API error |

### `GET /health`

Returns server status and route count.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_PATH` | `/config.yaml` | Path to configuration file |
| `*_BOT_TOKEN` | — | Bot tokens referenced in config.yaml |
| `*_CHAT_ID` | — | Chat IDs referenced in config.yaml |

## Development

```bash
# Run locally
CONFIG_PATH=./config.yaml go run .

# Build binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o pluma .
```

## License

MIT
