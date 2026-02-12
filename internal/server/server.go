package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ivangsm/pluma/internal/config"
	"github.com/ivangsm/pluma/internal/telegram"
)

// Server is the main HTTP server that handles contact routes.
type Server struct {
	cfg     *config.Config
	limiter *RateLimiter
	mux     *http.ServeMux
}

// New creates a Server and registers all routes from config.
func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg:     cfg,
		limiter: NewRateLimiter(),
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("GET /health", s.handleHealth)

	for _, route := range cfg.Routes {
		r := route
		window, err := config.ParseRateLimit(r.RateLimit)
		if err != nil {
			return nil, fmt.Errorf("route %s: %w", r.Path, err)
		}
		s.mux.HandleFunc("POST "+r.Path, s.contactHandler(r, window))
		log.Printf("Registered route: POST %s → bot:%s…%s chat:%s",
			r.Path, r.BotToken[:6], r.BotToken[len(r.BotToken)-4:], r.ChatID)
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
	for _, o := range strings.Split(s.cfg.Server.AllowedOrigins, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"routes": fmt.Sprintf("%d", len(s.cfg.Routes)),
	})
}

func (s *Server) contactHandler(route config.Route, window time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)

		// Rate limit check
		if !s.limiter.Allow(ip, route.Path, window) {
			JSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: "Rate limit exceeded. Please try again later.",
			})
			log.Printf("[429] %s %s from %s", r.Method, route.Path, ip)
			return
		}

		// Parse request body
		var req ContactRequest
		if err := decodeJSON(r, &req); err != nil {
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

		// Send to Telegram
		if err := telegram.SendMessage(route.BotToken, route.ChatID, req.Name, req.Email, req.Message, req.Source); err != nil {
			JSON(w, http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to send message. Please try again later.",
			})
			log.Printf("[500] %s %s from %s: %v", r.Method, route.Path, ip, err)
			return
		}

		JSON(w, http.StatusOK, SuccessResponse{
			Status: "Message sent successfully.",
		})
		log.Printf("[200] %s %s from %s", r.Method, route.Path, ip)
	}
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
