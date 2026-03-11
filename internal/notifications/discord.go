package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Discord sends notifications via Discord webhooks.
type Discord struct {
	WebhookURL string
	Username   string // optional bot username override
	AvatarURL  string // optional avatar override
	client     *http.Client
}

// NewDiscord creates a Discord notification provider.
func NewDiscord(webhookURL, username, avatarURL string) *Discord {
	return &Discord{
		WebhookURL: webhookURL,
		Username:   username,
		AvatarURL:  avatarURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *Discord) Name() string { return "Discord" }

func (d *Discord) Send(ctx context.Context, event Event) error {
	embed := map[string]any{
		"title":       event.Title,
		"description": event.Message,
		"color":       discordColor(event.Type),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	if len(event.Fields) > 0 {
		fields := make([]map[string]any, 0, len(event.Fields))
		for k, v := range event.Fields {
			fields = append(fields, map[string]any{
				"name":   k,
				"value":  v,
				"inline": true,
			})
		}
		embed["fields"] = fields
	}

	if event.AppType != "" {
		embed["footer"] = map[string]string{"text": event.AppType}
	}

	payload := map[string]any{
		"embeds": []any{embed},
	}
	if d.Username != "" {
		payload["username"] = d.Username
	}
	if d.AvatarURL != "" {
		payload["avatar_url"] = d.AvatarURL
	}

	return d.post(ctx, payload)
}

func (d *Discord) Test(ctx context.Context) error {
	return d.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, Discord notifications are working!",
	})
}

func (d *Discord) post(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}
	return nil
}

func discordColor(et EventType) int {
	switch et {
	case EventLurkCompleted:
		return 0x2ECC71 // green
	case EventQueueItemRemoved:
		return 0xE67E22 // orange
	case EventDownloadStuck:
		return 0xE74C3C // red
	case EventSchedulerAction:
		return 0x3498DB // blue
	case EventError:
		return 0xE74C3C // red
	default:
		return 0x95A5A6 // grey
	}
}
