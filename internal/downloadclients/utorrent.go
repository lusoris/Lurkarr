package downloadclient

import (
	"context"

	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/utorrent"
)

// UTorrentAdapter wraps a uTorrent client to implement the Client interface.
type UTorrentAdapter struct {
	client *utorrent.Client
}

// NewUTorrentAdapter creates a new uTorrent adapter.
func NewUTorrentAdapter(client *utorrent.Client) *UTorrentAdapter {
	return &UTorrentAdapter{client: client}
}

func (a *UTorrentAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(torrents))
	for _, t := range torrents {
		var tags []string
		if t.Label != "" {
			tags = []string{t.Label}
		}
		items = append(items, DownloadItem{
			ID:            t.Hash,
			Name:          t.Name,
			Status:        utorrentStatusString(t.Status),
			TotalSize:     t.Size,
			RemainingSize: t.Size - t.Downloaded,
			Progress:      float64(t.Progress) / 1000.0,
			DownloadSpeed: t.DownloadSpeed,
			UploadSpeed:   t.UploadSpeed,
			ETA:           t.ETA,
			Category:      t.Label,
			SavePath:      t.SavePath,
			Ratio:         float64(t.Ratio) / 1000.0,
			CompletedAt:   t.CompletedOn,
			AddedAt:       t.AddedOn,
			Tags:          tags,
		})
	}
	return items, nil
}

func utorrentStatusString(status int) string {
	switch {
	case status&128 != 0:
		return "error"
	case status&2 != 0:
		return "checking"
	case status&16 != 0:
		return "paused"
	case status&1 != 0:
		return "downloading"
	case status&32 != 0:
		return "queued"
	case status&64 != 0:
		return "stopped"
	default:
		return "unknown"
	}
}

func (a *UTorrentAdapter) GetHistory(ctx context.Context) ([]DownloadItem, error) {
	items, err := a.GetItems(ctx)
	if err != nil {
		return nil, err
	}
	return filterCompleted(items), nil
}

func (a *UTorrentAdapter) PauseAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	for _, t := range torrents {
		if err := a.client.PauseTorrent(ctx, t.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (a *UTorrentAdapter) ResumeAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	for _, t := range torrents {
		if err := a.client.UnpauseTorrent(ctx, t.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (a *UTorrentAdapter) ResumeItem(ctx context.Context, id string) error {
	return a.client.UnpauseTorrent(ctx, id)
}

func (a *UTorrentAdapter) RecheckItem(ctx context.Context, id string) error {
	return a.client.RecheckTorrent(ctx, id)
}

func (a *UTorrentAdapter) RemoveItem(ctx context.Context, id string, deleteData bool) error {
	return a.client.RemoveTorrent(ctx, id, deleteData)
}

func (a *UTorrentAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{Version: ver}, nil
}

func (a *UTorrentAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
