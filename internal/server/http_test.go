package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	t.Run("trust proxy reads X-Forwarded-For", func(t *testing.T) {
		r := &http.Request{
			RemoteAddr: "192.168.1.1:1234",
			Header:     http.Header{"X-Forwarded-For": {"10.0.0.1"}},
		}
		ip := GetClientIP(r, true)
		if ip != "10.0.0.1" {
			t.Fatalf("expected 10.0.0.1, got %q", ip)
		}
	})

	t.Run("no trust proxy ignores X-Forwarded-For", func(t *testing.T) {
		r := &http.Request{
			RemoteAddr: "192.168.1.1:1234",
			Header:     http.Header{"X-Forwarded-For": {"10.0.0.1"}},
		}
		ip := GetClientIP(r, false)
		if ip != "192.168.1.1" {
			t.Fatalf("expected 192.168.1.1, got %q", ip)
		}
	})

	t.Run("X-Forwarded-For multiple IPs returns first trimmed", func(t *testing.T) {
		r := &http.Request{
			RemoteAddr: "192.168.1.1:1234",
			Header:     http.Header{"X-Forwarded-For": {" 10.0.0.1 , 10.0.0.2, 10.0.0.3"}},
		}
		ip := GetClientIP(r, true)
		if ip != "10.0.0.1" {
			t.Fatalf("expected '10.0.0.1', got %q", ip)
		}
	})

	t.Run("trust proxy falls back to RemoteAddr", func(t *testing.T) {
		r := &http.Request{
			RemoteAddr: "192.168.1.1:1234",
			Header:     http.Header{},
		}
		ip := GetClientIP(r, true)
		if ip != "192.168.1.1" {
			t.Fatalf("expected 192.168.1.1, got %q", ip)
		}
	})
}

func TestJSON(t *testing.T) {
	t.Run("writes correct content type and status", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSON(w, http.StatusCreated, map[string]string{"status": "ok"})

		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected application/json, got %q", ct)
		}
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if body["status"] != "ok" {
			t.Fatalf("expected status 'ok', got %q", body["status"])
		}
	})
}
