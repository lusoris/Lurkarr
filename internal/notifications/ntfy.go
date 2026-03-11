package notifications

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Ntfy sends notifications via an ntfy server (https://ntfy.sh).
type Ntfy struct {
	ServerURL string
	Topic     string
	Token     string // optional access token
	Priority  int    // 1 (min) to 5 (max)
	client    *http.Client
}

// NewNtfy creates an ntfy notification provider.
func NewNtfy(serverURL, topic, token string, priority int) *Ntfy {
	return &Ntfy{
		ServerURL: strings.TrimRight(serverURL, "/"),
		Topic:     topic,
		Token:     token,
		Priority:  priority,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Ntfy) Name() string { return "ntfy" }

func (n *Ntfy) Send(ctx context.Context, event Event) error {
	apiURL := fmt.Sprintf("%s/%s", n.ServerURL, n.Topic)

	body := formatPlainMessage(event)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("create ntfy request: %w", err)
	}

	req.Header.Set("Title", event.Title)
	req.Header.Set("Priority", fmt.Sprintf("%d", n.Priority))
	req.Header.Set("Tags", ntfyTag(event.Type))

	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *Ntfy) Test(ctx context.Context) error {
	return n.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, ntfy notifications are working!",
	})
}

func ntfyTag(et EventType) string {
	switch et {
	case EventLurkCompleted:
		return "white_check_mark"
	case EventQueueItemRemoved:
		return "wastebasket"
	case EventDownloadStuck:
		return "warning"
	case EventError:
		return "rotating_light"
	case EventSchedulerAction:
		return "clock3"
	default:
		return "bell"
	}
}
