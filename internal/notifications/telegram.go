package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Telegram sends notifications via the Telegram Bot API.
type Telegram struct {
	BotToken string
	ChatID   string
	client   *http.Client
}

// NewTelegram creates a Telegram notification provider.
func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{
		BotToken: botToken,
		ChatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Telegram) Name() string { return "Telegram" }

func (t *Telegram) Send(ctx context.Context, event Event) error {
	text := formatTelegramMessage(event)
	return t.sendMessage(ctx, text)
}

func (t *Telegram) Test(ctx context.Context) error {
	return t.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, Telegram notifications are working!",
	})
}

func (t *Telegram) sendMessage(ctx context.Context, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	data := url.Values{
		"chat_id":    {t.ChatID},
		"text":       {text},
		"parse_mode": {"HTML"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}

func formatTelegramMessage(event Event) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s</b>\n", event.Title))
	sb.WriteString(event.Message)

	if event.AppType != "" || event.Instance != "" {
		sb.WriteString("\n")
		if event.AppType != "" {
			sb.WriteString(fmt.Sprintf("\n<i>App:</i> %s", event.AppType))
		}
		if event.Instance != "" {
			sb.WriteString(fmt.Sprintf("\n<i>Instance:</i> %s", event.Instance))
		}
	}

	if len(event.Fields) > 0 {
		sb.WriteString("\n")
		for k, v := range event.Fields {
			sb.WriteString(fmt.Sprintf("\n<b>%s:</b> %s", k, v))
		}
	}
	return sb.String()
}
