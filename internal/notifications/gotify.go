package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Gotify sends notifications via a Gotify server.
type Gotify struct {
	ServerURL string
	AppToken  string
	Priority  int // 0-10
	client    *http.Client
}

// NewGotify creates a Gotify notification provider.
func NewGotify(serverURL, appToken string, priority int) *Gotify {
	return &Gotify{
		ServerURL: strings.TrimRight(serverURL, "/"),
		AppToken:  appToken,
		Priority:  priority,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *Gotify) Name() string { return "Gotify" }

func (g *Gotify) Send(ctx context.Context, event Event) error {
	payload := map[string]any{
		"title":    event.Title,
		"message":  formatPlainMessage(event),
		"priority": g.Priority,
		"extras": map[string]any{
			"client::display": map[string]string{
				"contentType": "text/plain",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal gotify payload: %w", err)
	}

	apiURL := fmt.Sprintf("%s/message", g.ServerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create gotify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", g.AppToken)

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gotify request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotify returned status %d", resp.StatusCode)
	}
	return nil
}

func (g *Gotify) Test(ctx context.Context) error {
	return g.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, Gotify notifications are working!",
	})
}
