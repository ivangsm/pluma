package config

import (
	"os"
	"testing"
	"time"
)

func TestParseRateLimit(t *testing.T) {
	t.Run("valid 2/m", func(t *testing.T) {
		d, err := ParseRateLimit("2/m")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d != 30*time.Second {
			t.Fatalf("expected 30s, got %v", d)
		}
	})

	t.Run("valid 5/h", func(t *testing.T) {
		d, err := ParseRateLimit("5/h")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d != 12*time.Minute {
			t.Fatalf("expected 12m, got %v", d)
		}
	})

	t.Run("invalid format no slash", func(t *testing.T) {
		_, err := ParseRateLimit("2m")
		if err == nil {
			t.Fatal("expected error for missing slash")
		}
	})

	t.Run("invalid unit", func(t *testing.T) {
		_, err := ParseRateLimit("2/s")
		if err == nil {
			t.Fatal("expected error for invalid unit")
		}
	})

	t.Run("invalid count", func(t *testing.T) {
		_, err := ParseRateLimit("abc/m")
		if err == nil {
			t.Fatal("expected error for non-numeric count")
		}
	})

	t.Run("zero count", func(t *testing.T) {
		_, err := ParseRateLimit("0/m")
		if err == nil {
			t.Fatal("expected error for zero count")
		}
	})
}

func TestExpandEnv(t *testing.T) {
	t.Run("set variable", func(t *testing.T) {
		t.Setenv("PLUMA_TEST_VAR", "hello")
		result := expandEnv("value=${PLUMA_TEST_VAR}")
		if result != "value=hello" {
			t.Fatalf("expected 'value=hello', got %q", result)
		}
	})

	t.Run("unset variable", func(t *testing.T) {
		os.Unsetenv("PLUMA_TEST_UNSET")
		result := expandEnv("value=${PLUMA_TEST_UNSET}")
		if result != "value=" {
			t.Fatalf("expected 'value=', got %q", result)
		}
	})

	t.Run("no variables", func(t *testing.T) {
		result := expandEnv("plain text")
		if result != "plain text" {
			t.Fatalf("expected 'plain text', got %q", result)
		}
	})
}

func TestValidateEmail(t *testing.T) {
	t.Run("valid email", func(t *testing.T) {
		if !ValidateEmail("test@example.com") {
			t.Fatal("expected valid")
		}
	})

	t.Run("no at sign", func(t *testing.T) {
		if ValidateEmail("testexample.com") {
			t.Fatal("expected invalid")
		}
	})

	t.Run("no dot in domain", func(t *testing.T) {
		if ValidateEmail("test@localhost") {
			t.Fatal("expected invalid")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if ValidateEmail("") {
			t.Fatal("expected invalid")
		}
	})

	t.Run("at domain only", func(t *testing.T) {
		if ValidateEmail("@domain.com") {
			t.Fatal("expected invalid")
		}
	})
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "pluma-config-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad(t *testing.T) {
	t.Run("valid telegram config", func(t *testing.T) {
		path := writeTemp(t, `
server:
  port: 9090
  rate_limit: "3/m"
routes:
  - path: /contact
    provider: telegram
    telegram:
      bot_token: "token123"
      chat_id: "456"
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Server.Port != 9090 {
			t.Fatalf("expected port 9090, got %d", cfg.Server.Port)
		}
		if cfg.Server.RateLimit != "3/m" {
			t.Fatalf("expected rate_limit '3/m', got %q", cfg.Server.RateLimit)
		}
		if len(cfg.Routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
		}
		if cfg.Routes[0].Provider != "telegram" {
			t.Fatalf("expected provider 'telegram', got %q", cfg.Routes[0].Provider)
		}
	})

	t.Run("valid discord config", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    provider: discord
    discord:
      webhook_url: "https://discord.com/api/webhooks/123/abc"
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Routes[0].Provider != "discord" {
			t.Fatalf("expected provider 'discord', got %q", cfg.Routes[0].Provider)
		}
		if cfg.Routes[0].Discord.WebhookURL != "https://discord.com/api/webhooks/123/abc" {
			t.Fatalf("unexpected webhook_url: %q", cfg.Routes[0].Discord.WebhookURL)
		}
	})

	t.Run("missing routes", func(t *testing.T) {
		path := writeTemp(t, `
server:
  port: 8080
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing routes")
		}
	})

	t.Run("missing provider", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    telegram:
      bot_token: "tok"
      chat_id: "123"
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing provider")
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    provider: slack
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for unknown provider")
		}
	})

	t.Run("missing telegram bot_token", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    provider: telegram
    telegram:
      chat_id: "456"
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing bot_token")
		}
	})

	t.Run("missing discord webhook_url", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    provider: discord
    discord:
      webhook_url: ""
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for missing webhook_url")
		}
	})

	t.Run("legacy format migrates to telegram", func(t *testing.T) {
		path := writeTemp(t, `
server:
  port: 9090
  rate_limit: "3/m"
routes:
  - path: /contact
    bot_token: "token123"
    chat_id: "456"
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r := cfg.Routes[0]
		if r.Provider != "telegram" {
			t.Fatalf("expected provider 'telegram', got %q", r.Provider)
		}
		if r.Telegram == nil {
			t.Fatal("expected telegram config to be set")
		}
		if r.Telegram.BotToken != "token123" {
			t.Fatalf("expected bot_token 'token123', got %q", r.Telegram.BotToken)
		}
		if r.Telegram.ChatID != "456" {
			t.Fatalf("expected chat_id '456', got %q", r.Telegram.ChatID)
		}
		if r.BotToken != "" || r.ChatID != "" {
			t.Fatal("expected legacy fields to be cleared after migration")
		}
	})

	t.Run("defaults applied", func(t *testing.T) {
		path := writeTemp(t, `
routes:
  - path: /contact
    provider: telegram
    telegram:
      bot_token: "tok"
      chat_id: "123"
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Server.Port != DefaultPort {
			t.Fatalf("expected default port %d, got %d", DefaultPort, cfg.Server.Port)
		}
		if cfg.Server.RateLimit != DefaultRateLimit {
			t.Fatalf("expected default rate limit %q, got %q", DefaultRateLimit, cfg.Server.RateLimit)
		}
	})
}
