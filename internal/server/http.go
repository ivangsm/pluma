package server

import (
	"encoding/json"
	"net"
	"net/http"
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
// respecting X-Forwarded-For and X-Real-IP headers.
func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
