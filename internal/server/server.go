package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ivangsm/pluma/internal/config"
	"github.com/ivangsm/pluma/internal/provider"
)

// Server is the main HTTP server that handles contact routes.
type Server struct {
	cfg     *config.Config
	limiter *RateLimiter
	mux     *http.ServeMux
}

// New creates a Server and registers all routes from config.
// The providers map must contain a provider.Provider for each route path.
func New(ctx context.Context, cfg *config.Config, providers map[string]provider.Provider) (*Server, error) {
	s := &Server{
		cfg:     cfg,
		limiter: NewRateLimiter(ctx),
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("GET /health", s.handleHealth)

	for _, route := range cfg.Routes {
		r := route
		p, ok := providers[r.Path]
		if !ok {
			return nil, fmt.Errorf("no provider registered for route %s", r.Path)
		}
		window, err := config.ParseRateLimit(r.RateLimit)
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", r.Path, err)
		}
		s.mux.HandleFunc("POST "+r.Path, s.contactHandler(r, p, window))
		slog.Info("route registered", "method", "POST", "path", r.Path, "provider", r.Provider)
	}

	return s, nil
}

// ServeHTTP implements http.Handler with CORS support.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	allowed := s.isOriginAllowed(origin)

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if s.cfg.Server.AllowedOrigins == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

// isOriginAllowed checks if the given origin is in the allowed list.
func (s *Server) isOriginAllowed(origin string) bool {
	if s.cfg.Server.AllowedOrigins == "*" {
		return true
	}
	return s.cfg.Server.ParsedOrigins[origin]
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"routes": fmt.Sprintf("%d", len(s.cfg.Routes)),
	})
}

func (s *Server) contactHandler(route config.Route, p provider.Provider, window time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r, s.cfg.Server.TrustProxy)

		// Rate limit check
		if !s.limiter.Allow(ip, route.Path, window) {
			JSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: "Rate limit exceeded. Please try again later.",
			})
			slog.Warn("rate limited", "method", r.Method, "path", route.Path, "ip", ip)
			return
		}

		// Parse request body
		var req ContactRequest
		if err := decodeJSON(w, r, &req); err != nil {
			JSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "Invalid request body. Expected JSON with name, email, and message.",
			})
			return
		}

		// Validate
		if req.Name == "" || req.Email == "" || req.Message == "" {
			JSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "All fields (name, email, message) are required.",
			})
			return
		}

		if !config.ValidateEmail(req.Email) {
			JSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "Invalid email address.",
			})
			return
		}

		// Send via provider
		msg := provider.ContactMessage{
			Name:    req.Name,
			Email:   req.Email,
			Message: req.Message,
			Source:  req.Source,
		}
		if err := p.Send(r.Context(), msg); err != nil {
			JSON(w, http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to send message. Please try again later.",
			})
			slog.Error("send failed", "method", r.Method, "path", route.Path, "provider", route.Provider, "ip", ip, "error", err)
			return
		}

		JSON(w, http.StatusOK, SuccessResponse{
			Status: "Message sent successfully.",
		})
		slog.Info("message sent", "method", r.Method, "path", route.Path, "provider", route.Provider, "ip", ip)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	return json.NewDecoder(r.Body).Decode(v)
}
