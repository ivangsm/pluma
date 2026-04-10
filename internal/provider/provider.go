package provider

import "context"

// ContactMessage holds the data from a contact form submission.
type ContactMessage struct {
	Name    string
	Email   string
	Message string
	Source  string
}

// Provider sends a contact message through a specific channel (Telegram, Discord, etc.).
type Provider interface {
	Send(ctx context.Context, msg ContactMessage) error
}
