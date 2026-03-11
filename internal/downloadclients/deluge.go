package downloadclient

import (
	"context"

	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/deluge"
)

// DelugeAdapter wraps a Deluge client to implement the Client interface.
type DelugeAdapter struct {
	client *deluge.Client
}

// NewDelugeAdapter creates a new Deluge adapter.
func NewDelugeAdapter(client *deluge.Client) *DelugeAdapter {
	return &DelugeAdapter{client: client}
}

func (a *DelugeAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(torrents))
	for _, t := range torrents {
		items = append(items, DownloadItem{
			ID:            t.Hash,
			Name:          t.Name,
			Status:        t.State,
			TotalSize:     t.TotalSize,
			RemainingSize: t.TotalSize - t.TotalDone,
			Progress:      t.Progress / 100.0,
			DownloadSpeed: int64(t.DownloadSpeed),
			UploadSpeed:   int64(t.UploadSpeed),
			ETA:           t.ETA,
			Category:      t.Label,
			SavePath:      t.SavePath,
			Ratio:         t.Ratio,
			SeedingTime:   t.SeedingTime,
			AddedAt:       int64(t.TimeAdded),
		})
	}
	return items, nil
}

func (a *DelugeAdapter) GetHistory(ctx context.Context) ([]DownloadItem, error) {
	// Deluge keeps all torrents in the main list.
	// Filter to completed items only.
	items, err := a.GetItems(ctx)
	if err != nil {
		return nil, err
	}
	var completed []DownloadItem
	for _, item := range items {
		if item.Progress >= 1.0 {
			completed = append(completed, item)
		}
	}
	return completed, nil
}

func (a *DelugeAdapter) PauseAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	hashes := make([]string, len(torrents))
	for i, t := range torrents {
		hashes[i] = t.Hash
	}
	return a.client.PauseTorrents(ctx, hashes)
}

func (a *DelugeAdapter) ResumeAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	hashes := make([]string, len(torrents))
	for i, t := range torrents {
		hashes[i] = t.Hash
	}
	return a.client.ResumeTorrents(ctx, hashes)
}

func (a *DelugeAdapter) RemoveItem(ctx context.Context, id string, deleteData bool) error {
	return a.client.DeleteTorrents(ctx, []string{id}, deleteData)
}

func (a *DelugeAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version: ver,
	}, nil
}

func (a *DelugeAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
