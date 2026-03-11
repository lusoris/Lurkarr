package downloadclient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/nzbget"
)

// NZBGetAdapter wraps an NZBGet client to implement the Client interface.
type NZBGetAdapter struct {
	client *nzbget.Client
}

// NewNZBGetAdapter creates a new NZBGet adapter.
func NewNZBGetAdapter(client *nzbget.Client) *NZBGetAdapter {
	return &NZBGetAdapter{client: client}
}

func (a *NZBGetAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	queue, err := a.client.GetQueue(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(queue))
	for _, q := range queue {
		// NZBGet splits sizes into Hi/Lo 32-bit parts.
		totalSize := (q.FileSizeHi << 32) | q.FileSizeLo
		remaining := (q.RemainingSizeHi << 32) | q.RemainingSizeLo
		var progress float64
		if totalSize > 0 {
			progress = float64(totalSize-remaining) / float64(totalSize)
		}
		items = append(items, DownloadItem{
			ID:            strconv.Itoa(q.NZBID),
			Name:          q.NZBName,
			Status:        q.Status,
			TotalSize:     totalSize,
			RemainingSize: remaining,
			Progress:      progress,
			Category:      q.Category,
		})
	}
	return items, nil
}

func (a *NZBGetAdapter) GetHistory(ctx context.Context) ([]DownloadItem, error) {
	hist, err := a.client.GetHistory(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(hist))
	for _, h := range hist {
		totalSize := (h.FileSizeHi << 32) | h.FileSizeLo
		items = append(items, DownloadItem{
			ID:        strconv.Itoa(h.NZBID),
			Name:      h.NZBName,
			Status:    h.Status,
			TotalSize: totalSize,
			Category:  h.Category,
			SavePath:  h.DestDir,
			Progress:  1.0,
		})
	}
	return items, nil
}

func (a *NZBGetAdapter) PauseAll(ctx context.Context) error {
	return a.client.Pause(ctx)
}

func (a *NZBGetAdapter) ResumeAll(ctx context.Context) error {
	return a.client.Resume(ctx)
}

func (a *NZBGetAdapter) RemoveItem(ctx context.Context, id string, _ bool) error {
	nzbID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid NZBGet NZBID: %w", err)
	}
	return a.client.DeleteItem(ctx, nzbID)
}

func (a *NZBGetAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	status, err := a.client.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version:       ver,
		DownloadSpeed: status.DownloadRate,
		Paused:        status.DownloadPaused,
	}, nil
}

func (a *NZBGetAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
