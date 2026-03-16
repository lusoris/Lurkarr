package downloadclient

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/sabnzbd"
)

// SABnzbdAdapter wraps a SABnzbd client to implement the Client interface.
type SABnzbdAdapter struct {
	client *sabnzbd.Client
}

// NewSABnzbdAdapter creates a new SABnzbd adapter.
func NewSABnzbdAdapter(client *sabnzbd.Client) *SABnzbdAdapter {
	return &SABnzbdAdapter{client: client}
}

func (a *SABnzbdAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	queue, err := a.client.GetQueue(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(queue.Slots))
	for _, s := range queue.Slots {
		pct, err := strconv.ParseFloat(s.Percentage, 64)
		if err != nil {
			slog.Warn("sabnzbd: failed to parse percentage", "nzo_id", s.NzoID, "value", s.Percentage, "error", err)
		}
		items = append(items, DownloadItem{
			ID:       s.NzoID,
			Name:     s.Filename,
			Status:   s.Status,
			Category: s.Category,
			Progress: pct / 100.0,
		})
	}
	return items, nil
}

func (a *SABnzbdAdapter) GetHistory(ctx context.Context) ([]DownloadItem, error) {
	hist, err := a.client.GetHistory(ctx, 200)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(hist.Slots))
	for _, s := range hist.Slots {
		items = append(items, DownloadItem{
			ID:          s.NzoID,
			Name:        s.Name,
			Status:      s.Status,
			Category:    s.Category,
			Progress:    1.0,
			CompletedAt: s.CompletedAt,
		})
	}
	return items, nil
}

func (a *SABnzbdAdapter) PauseAll(ctx context.Context) error {
	return a.client.Pause(ctx)
}

func (a *SABnzbdAdapter) ResumeAll(ctx context.Context) error {
	return a.client.Resume(ctx)
}

func (a *SABnzbdAdapter) ResumeItem(_ context.Context, _ string) error { return nil }

func (a *SABnzbdAdapter) RecheckItem(_ context.Context, _ string) error { return nil }

func (a *SABnzbdAdapter) RemoveItem(ctx context.Context, id string, _ bool) error {
	return a.client.DeleteQueueItem(ctx, id)
}

func (a *SABnzbdAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	queue, err := a.client.GetQueue(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version:   ver,
		Paused:    queue.Paused,
		ItemCount: queue.NoOfSlots,
	}, nil
}

func (a *SABnzbdAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
