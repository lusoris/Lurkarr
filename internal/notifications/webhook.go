package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Webhook sends notifications as JSON POST requests to a custom URL.
type Webhook struct {
	URL     string
	Headers map[string]string // optional custom headers
	client  *http.Client
}

// NewWebhook creates a generic webhook notification provider.
func NewWebhook(webhookURL string, headers map[string]string) *Webhook {
	return &Webhook{
		URL:     webhookURL,
		Headers: headers,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *Webhook) Name() string { return "Webhook" }

// WebhookPayload is the JSON body sent to the webhook URL.
type WebhookPayload struct {
	Event    string            `json:"event"`
	Title    string            `json:"title"`
	Message  string            `json:"message"`
	AppType  string            `json:"app_type,omitempty"`
	Instance string            `json:"instance,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
	Time     string            `json:"time"`
}

func (w *Webhook) Send(ctx context.Context, event Event) error {
	payload := WebhookPayload{
		Event:    string(event.Type),
		Title:    event.Title,
		Message:  event.Message,
		AppType:  event.AppType,
		Instance: event.Instance,
		Fields:   event.Fields,
		Time:     time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Lurkarr")

	for k, v := range w.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (w *Webhook) Test(ctx context.Context) error {
	return w.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, webhook notifications are working!",
	})
}
