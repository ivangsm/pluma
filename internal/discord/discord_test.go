package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ivangsm/pluma/internal/provider"
)

func TestSend(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected application/json, got %q", ct)
			}

			var payload webhookPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decoding payload: %v", err)
			}

			if len(payload.Embeds) != 1 {
				t.Fatalf("expected 1 embed, got %d", len(payload.Embeds))
			}

			embed := payload.Embeds[0]
			if len(embed.Fields) != 3 {
				t.Fatalf("expected 3 fields (no source), got %d", len(embed.Fields))
			}
			if embed.Fields[0].Value != "John" {
				t.Fatalf("expected name 'John', got %q", embed.Fields[0].Value)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		d := &Discord{WebhookURL: srv.URL}
		err := d.Send(context.Background(), provider.ContactMessage{
			Name:    "John",
			Email:   "john@example.com",
			Message: "Hello",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with source field", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload webhookPayload
			json.NewDecoder(r.Body).Decode(&payload)

			if len(payload.Embeds[0].Fields) != 4 {
				t.Fatalf("expected 4 fields (with source), got %d", len(payload.Embeds[0].Fields))
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		d := &Discord{WebhookURL: srv.URL}
		err := d.Send(context.Background(), provider.ContactMessage{
			Name:    "John",
			Email:   "john@example.com",
			Message: "Hello",
			Source:  "website",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("api error returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer srv.Close()

		d := &Discord{WebhookURL: srv.URL}
		err := d.Send(context.Background(), provider.ContactMessage{
			Name:    "John",
			Email:   "john@example.com",
			Message: "Hello",
		})
		if err == nil {
			t.Fatal("expected error for bad status")
		}
	})
}
