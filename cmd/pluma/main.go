package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ivangsm/pluma/internal/config"
	"github.com/ivangsm/pluma/internal/discord"
	"github.com/ivangsm/pluma/internal/provider"
	"github.com/ivangsm/pluma/internal/server"
	"github.com/ivangsm/pluma/internal/telegram"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config.yaml"
	}

	slog.Info("Pluma starting", "config", configPath)

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	providers, err := buildProviders(cfg)
	if err != nil {
		slog.Error("Failed to build providers", "error", err)
		os.Exit(1)
	}

	slog.Info("Routes loaded", "routes", len(cfg.Routes), "port", cfg.Server.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, err := server.New(ctx, cfg, providers)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Pluma ready")

	<-done
	slog.Info("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("Goodbye.")
}

func buildProviders(cfg *config.Config) (map[string]provider.Provider, error) {
	providers := make(map[string]provider.Provider, len(cfg.Routes))

	for _, route := range cfg.Routes {
		switch route.Provider {
		case "telegram":
			providers[route.Path] = &telegram.Telegram{
				BotToken: route.Telegram.BotToken,
				ChatID:   route.Telegram.ChatID,
			}
		case "discord":
			providers[route.Path] = &discord.Discord{
				WebhookURL: route.Discord.WebhookURL,
			}
		default:
			return nil, fmt.Errorf("unknown provider %q for route %s", route.Provider, route.Path)
		}
	}

	return providers, nil
}
