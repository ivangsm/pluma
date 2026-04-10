package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ivangsm/pluma/internal/provider"
)

var client = &http.Client{Timeout: 10 * time.Second}

// Discord sends contact messages via a Discord webhook.
type Discord struct {
	WebhookURL string
}

type webhookPayload struct {
	Embeds []embed `json:"embeds"`
}

type embed struct {
	Title  string       `json:"title"`
	Color  int          `json:"color"`
	Fields []embedField `json:"fields"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Send implements provider.Provider.
func (d *Discord) Send(_ context.Context, msg provider.ContactMessage) error {
	fields := []embedField{
		{Name: "Name", Value: msg.Name, Inline: true},
		{Name: "Email", Value: msg.Email, Inline: true},
		{Name: "Message", Value: msg.Message},
	}

	if msg.Source != "" {
		fields = append(fields, embedField{Name: "Source", Value: msg.Source, Inline: true})
	}

	payload := webhookPayload{
		Embeds: []embed{
			{
				Title:  "📩 New Contact Message",
				Color:  0x5865F2, // Discord blurple
				Fields: fields,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling discord payload: %w", err)
	}

	resp, err := client.Post(d.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord API error: status %d", resp.StatusCode)
	}

	return nil
}
