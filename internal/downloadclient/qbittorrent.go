package downloadclient

import (
	"context"

	"github.com/lusoris/lurkarr/internal/qbittorrent"
)

// QBittorrentAdapter wraps a qBittorrent client to implement the Client interface.
type QBittorrentAdapter struct {
	client *qbittorrent.Client
}

// NewQBittorrentAdapter creates a new qBittorrent adapter.
func NewQBittorrentAdapter(client *qbittorrent.Client) *QBittorrentAdapter {
	return &QBittorrentAdapter{client: client}
}

func (a *QBittorrentAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	torrents, err := a.client.GetTorrents(ctx, "", "")
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
			RemainingSize: t.AmountLeft,
			Progress:      t.Progress,
			DownloadSpeed: t.DownloadSpeed,
			UploadSpeed:   t.UploadSpeed,
			ETA:           t.ETA,
			Category:      t.Category,
			SavePath:      t.SavePath,
		})
	}
	return items, nil
}

func (a *QBittorrentAdapter) PauseAll(ctx context.Context) error {
	return a.client.PauseTorrents(ctx, []string{"all"})
}

func (a *QBittorrentAdapter) ResumeAll(ctx context.Context) error {
	return a.client.ResumeTorrents(ctx, []string{"all"})
}

func (a *QBittorrentAdapter) RemoveItem(ctx context.Context, id string, deleteData bool) error {
	return a.client.DeleteTorrents(ctx, []string{id}, deleteData)
}

func (a *QBittorrentAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	info, err := a.client.GetTransferInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version:       ver,
		DownloadSpeed: info.DownloadSpeed,
		UploadSpeed:   info.UploadSpeed,
	}, nil
}

func (a *QBittorrentAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
