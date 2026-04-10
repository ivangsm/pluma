package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ivangsm/pluma/internal/provider"
)

var client = &http.Client{Timeout: 10 * time.Second}

type response struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// Telegram sends contact messages via the Telegram Bot API.
type Telegram struct {
	BotToken string
	ChatID   string
}

// Send implements provider.Provider.
func (t *Telegram) Send(_ context.Context, msg provider.ContactMessage) error {
	text := fmt.Sprintf(
		"📩 <b>New Contact Message</b>\n\n"+
			"<b>Name:</b> %s\n"+
			"<b>Email:</b> %s\n\n"+
			"<b>Message:</b>\n%s",
		escapeHTML(msg.Name),
		escapeHTML(msg.Email),
		escapeHTML(msg.Message),
	)

	if msg.Source != "" {
		text += fmt.Sprintf("\n\n🌐 <b>Source:</b> %s", escapeHTML(msg.Source))
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	resp, err := client.PostForm(apiURL, url.Values{
		"chat_id":    {t.ChatID},
		"text":       {text},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	var tgResp response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("decoding telegram response: %w", err)
	}

	if !tgResp.OK {
		return fmt.Errorf("telegram API error: %s", tgResp.Description)
	}

	return nil
}

func escapeHTML(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			result = append(result, []byte("&amp;")...)
		case '<':
			result = append(result, []byte("&lt;")...)
		case '>':
			result = append(result, []byte("&gt;")...)
		case '"':
			result = append(result, []byte("&quot;")...)
		case '\'':
			result = append(result, []byte("&#39;")...)
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}
