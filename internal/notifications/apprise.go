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

// Apprise sends notifications via an Apprise API server.
// See https://github.com/caronc/apprise-api
type Apprise struct {
	ServerURL string   // Apprise API server URL (e.g. "http://apprise:8000")
	URLs      []string // Apprise-compatible notification URLs
	Tag       string   // optional Apprise tag
	client    *http.Client
}

// NewApprise creates an Apprise notification provider.
func NewApprise(serverURL string, urls []string, tag string) *Apprise {
	return &Apprise{
		ServerURL: strings.TrimRight(serverURL, "/"),
		URLs:      urls,
		Tag:       tag,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *Apprise) Name() string { return "Apprise" }

func (a *Apprise) Send(ctx context.Context, event Event) error {
	payload := map[string]any{
		"urls":  strings.Join(a.URLs, ","),
		"title": event.Title,
		"body":  formatPlainMessage(event),
		"type":  appriseType(event.Type),
	}

	if a.Tag != "" {
		payload["tag"] = a.Tag
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal apprise payload: %w", err)
	}

	apiURL := fmt.Sprintf("%s/notify/", a.ServerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create apprise request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("apprise request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apprise returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *Apprise) Test(ctx context.Context) error {
	return a.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, Apprise notifications are working!",
	})
}

func appriseType(et EventType) string {
	switch et {
	case EventError, EventDownloadStuck:
		return "failure"
	case EventLurkCompleted:
		return "success"
	default:
		return "info"
	}
}
