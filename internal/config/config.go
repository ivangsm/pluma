package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the entire application configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	Routes []Route      `yaml:"routes"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port           int    `yaml:"port"`
	RateLimit      string `yaml:"rate_limit"`
	AllowedOrigins string `yaml:"allowed_origins"`
}

// Route maps a URL path to a Telegram bot and chat.
type Route struct {
	Path      string `yaml:"path"`
	BotToken  string `yaml:"bot_token"`
	ChatID    string `yaml:"chat_id"`
	RateLimit string `yaml:"rate_limit"`
}

// Load reads and parses a YAML config file.
// Values containing ${ENV_VAR} are expanded from environment variables.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	expanded := expandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.RateLimit == "" {
		cfg.Server.RateLimit = "1/m"
	}
	if cfg.Server.AllowedOrigins == "" {
		cfg.Server.AllowedOrigins = "*"
	}

	// Validate routes
	for i, r := range cfg.Routes {
		if r.Path == "" {
			return nil, fmt.Errorf("route %d: path is required", i)
		}
		if r.BotToken == "" {
			return nil, fmt.Errorf("route %d (%s): bot_token is required", i, r.Path)
		}
		if r.ChatID == "" {
			return nil, fmt.Errorf("route %d (%s): chat_id is required", i, r.Path)
		}
		if r.RateLimit == "" {
			cfg.Routes[i].RateLimit = cfg.Server.RateLimit
		}
	}

	if len(cfg.Routes) == 0 {
		return nil, fmt.Errorf("at least one route is required")
	}

	return &cfg, nil
}

// ParseRateLimit parses "N/m" or "N/h" into a duration between allowed requests.
func ParseRateLimit(rl string) (time.Duration, error) {
	parts := strings.Split(rl, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rate limit format: %s (expected N/m or N/h)", rl)
	}

	n, err := strconv.Atoi(parts[0])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid rate limit count: %s", parts[0])
	}

	var window time.Duration
	switch parts[1] {
	case "m":
		window = time.Minute
	case "h":
		window = time.Hour
	default:
		return 0, fmt.Errorf("invalid rate limit unit: %s (use m or h)", parts[1])
	}

	return window / time.Duration(n), nil
}

// expandEnv replaces ${VAR} patterns with environment variable values.
func expandEnv(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '$' && s[i+1] == '{' {
			end := -1
			for j := i + 2; j < len(s); j++ {
				if s[j] == '}' {
					end = j
					break
				}
			}
			if end > 0 {
				varName := s[i+2 : end]
				result = append(result, os.Getenv(varName)...)
				i = end + 1
				continue
			}
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
