package downloadclient

import (
	"context"
	"fmt"
	"time"

	gort "github.com/autobrr/go-rtorrent"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/rtorrent"
)

// RTorrentAdapter wraps an rTorrent client to implement the Client interface.
type RTorrentAdapter struct {
	client *rtorrent.Client
}

// NewRTorrentAdapter creates a new rTorrent adapter.
func NewRTorrentAdapter(client *rtorrent.Client) *RTorrentAdapter {
	return &RTorrentAdapter{client: client}
}

func (a *RTorrentAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(torrents))
	for _, t := range torrents {
		status, err := a.client.GetStatus(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("get status for %s: %w", t.Hash, err)
		}
		active, _ := a.client.IsActive(ctx, t)

		items = append(items, DownloadItem{
			ID:            t.Hash,
			Name:          t.Name,
			Status:        rtorrentStatusString(t, status, active),
			TotalSize:     int64(status.Size),
			RemainingSize: int64(status.Size - status.CompletedBytes),
			Progress:      rtorrentProgress(status),
			DownloadSpeed: int64(status.DownRate),
			UploadSpeed:   int64(status.UpRate),
			ETA:           rtorrentETA(status),
			SavePath:      t.Path,
			Ratio:         status.Ratio,
			Category:      t.Label,
			AddedAt:       t.Created.Unix(),
			CompletedAt:   rtorrentCompletedAt(t),
		})
	}
	return items, nil
}

func (a *RTorrentAdapter) GetHistory(ctx context.Context) ([]DownloadItem, error) {
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

func rtorrentStatusString(t gort.Torrent, s gort.Status, active bool) string {
	if s.Completed {
		if active && s.UpRate > 0 {
			return "seeding"
		}
		return "seeding"
	}
	if !active {
		return "paused"
	}
	if s.DownRate > 0 {
		return "downloading"
	}
	return "downloading"
}

func rtorrentProgress(s gort.Status) float64 {
	if s.Size == 0 {
		return 0
	}
	return float64(s.CompletedBytes) / float64(s.Size)
}

func rtorrentETA(s gort.Status) int64 {
	if s.DownRate <= 0 {
		return 0
	}
	remaining := s.Size - s.CompletedBytes
	if remaining <= 0 {
		return 0
	}
	return int64(remaining) / int64(s.DownRate)
}

func rtorrentCompletedAt(t gort.Torrent) int64 {
	if t.Finished.IsZero() {
		return 0
	}
	return t.Finished.Unix()
}

func (a *RTorrentAdapter) PauseAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	for _, t := range torrents {
		if err := a.client.PauseTorrent(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (a *RTorrentAdapter) ResumeAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	for _, t := range torrents {
		if err := a.client.ResumeTorrent(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (a *RTorrentAdapter) ResumeItem(ctx context.Context, id string) error {
	t, err := a.client.GetTorrent(ctx, id)
	if err != nil {
		return err
	}
	return a.client.ResumeTorrent(ctx, t)
}

func (a *RTorrentAdapter) RecheckItem(ctx context.Context, id string) error {
	// go-rtorrent doesn't expose d.check_hash directly; stop + start triggers recheck.
	t, err := a.client.GetTorrent(ctx, id)
	if err != nil {
		return err
	}
	if err := a.client.StopTorrent(ctx, t); err != nil {
		return err
	}
	return a.client.StartTorrent(ctx, t)
}

func (a *RTorrentAdapter) RemoveItem(ctx context.Context, id string, _ bool) error {
	t, err := a.client.GetTorrent(ctx, id)
	if err != nil {
		return err
	}
	return a.client.Delete(ctx, t)
}

func (a *RTorrentAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	name, err := a.client.Name(ctx)
	if err != nil {
		return nil, err
	}
	downRate, err := a.client.DownRate(ctx)
	if err != nil {
		return nil, err
	}
	upRate, err := a.client.UpRate(ctx)
	if err != nil {
		return nil, err
	}
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version:       name,
		DownloadSpeed: int64(downRate),
		UploadSpeed:   int64(upRate),
		ItemCount:     len(torrents),
	}, nil
}

func (a *RTorrentAdapter) TestConnection(ctx context.Context) (string, error) {
	name, err := a.client.Name(ctx)
	if err != nil {
		return "", err
	}
	return name, nil
}

// SeedingTime calculates seeding time for a torrent.
func rtorrentSeedingTime(t gort.Torrent) int64 {
	if t.Finished.IsZero() {
		return 0
	}
	return int64(time.Since(t.Finished).Seconds())
}
