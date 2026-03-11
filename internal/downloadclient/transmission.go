package downloadclient

import (
	"context"
	"strconv"

	"github.com/lusoris/lurkarr/internal/transmission"
)

// TransmissionAdapter wraps a Transmission client to implement the Client interface.
type TransmissionAdapter struct {
	client *transmission.Client
}

// NewTransmissionAdapter creates a new Transmission adapter.
func NewTransmissionAdapter(client *transmission.Client) *TransmissionAdapter {
	return &TransmissionAdapter{client: client}
}

func (a *TransmissionAdapter) GetItems(ctx context.Context) ([]DownloadItem, error) {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DownloadItem, 0, len(torrents))
	for _, t := range torrents {
		items = append(items, DownloadItem{
			ID:            strconv.Itoa(t.ID),
			Name:          t.Name,
			Status:        transmissionStatusString(t.Status),
			TotalSize:     t.TotalSize,
			RemainingSize: t.LeftUntilDone,
			Progress:      t.PercentDone,
			DownloadSpeed: t.RateDownload,
			UploadSpeed:   t.RateUpload,
			ETA:           t.ETA,
			SavePath:      t.DownloadDir,
		})
	}
	return items, nil
}

func transmissionStatusString(status int) string {
	switch status {
	case transmission.StatusStopped:
		return "stopped"
	case transmission.StatusCheckWait:
		return "check_wait"
	case transmission.StatusChecking:
		return "checking"
	case transmission.StatusDownloadWait:
		return "download_wait"
	case transmission.StatusDownloading:
		return "downloading"
	case transmission.StatusSeedWait:
		return "seed_wait"
	case transmission.StatusSeeding:
		return "seeding"
	default:
		return "unknown"
	}
}

func (a *TransmissionAdapter) PauseAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	ids := make([]int, len(torrents))
	for i, t := range torrents {
		ids[i] = t.ID
	}
	return a.client.PauseTorrents(ctx, ids)
}

func (a *TransmissionAdapter) ResumeAll(ctx context.Context) error {
	torrents, err := a.client.GetTorrents(ctx)
	if err != nil {
		return err
	}
	ids := make([]int, len(torrents))
	for i, t := range torrents {
		ids[i] = t.ID
	}
	return a.client.ResumeTorrents(ctx, ids)
}

func (a *TransmissionAdapter) RemoveItem(ctx context.Context, id string, deleteData bool) error {
	torrentID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return a.client.DeleteTorrents(ctx, []int{torrentID}, deleteData)
}

func (a *TransmissionAdapter) GetStatus(ctx context.Context) (*ClientStatus, error) {
	ver, err := a.client.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := a.client.GetSessionStats(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientStatus{
		Version:       ver,
		DownloadSpeed: stats.DownloadSpeed,
		UploadSpeed:   stats.UploadSpeed,
		ItemCount:     stats.TorrentCount,
	}, nil
}

func (a *TransmissionAdapter) TestConnection(ctx context.Context) (string, error) {
	return a.client.TestConnection(ctx)
}
