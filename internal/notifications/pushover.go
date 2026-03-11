package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Pushover sends notifications via the Pushover API.
type Pushover struct {
	APIToken string
	UserKey  string
	Device   string // optional: target specific device
	Priority int    // -2 to 2 (lowest to emergency)
	client   *http.Client
}

// NewPushover creates a Pushover notification provider.
func NewPushover(apiToken, userKey, device string, priority int) *Pushover {
	return &Pushover{
		APIToken: apiToken,
		UserKey:  userKey,
		Device:   device,
		Priority: priority,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *Pushover) Name() string { return "Pushover" }

func (p *Pushover) Send(ctx context.Context, event Event) error {
	data := url.Values{
		"token":   {p.APIToken},
		"user":    {p.UserKey},
		"title":   {event.Title},
		"message": {formatPlainMessage(event)},
	}

	if p.Device != "" {
		data.Set("device", p.Device)
	}

	priority := p.Priority
	if event.Type == EventError {
		priority = max(priority, 1) // at least high for errors
	}
	data.Set("priority", fmt.Sprintf("%d", priority))

	// Emergency priority requires retry/expire params
	if priority == 2 {
		data.Set("retry", "60")
		data.Set("expire", "3600")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.pushover.net/1/messages.json", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create pushover request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pushover request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover returned status %d", resp.StatusCode)
	}
	return nil
}

func (p *Pushover) Test(ctx context.Context) error {
	return p.Send(ctx, Event{
		Type:    EventTestNotification,
		Title:   "Lurkarr Test Notification",
		Message: "If you see this, Pushover notifications are working!",
	})
}
