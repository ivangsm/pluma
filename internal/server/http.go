package server

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// ContactRequest is the expected JSON body for contact endpoints.
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse is a standard error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse is a standard success payload.
type SuccessResponse struct {
	Status string `json:"status"`
}

// GetClientIP extracts the real client IP from the request,
// respecting X-Forwarded-For and X-Real-IP headers only when trustProxy is true.
func GetClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i > 0 {
				return strings.TrimSpace(xff[:i])
			}
			return strings.TrimSpace(xff)
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
