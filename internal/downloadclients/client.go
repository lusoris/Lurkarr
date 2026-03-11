// Package downloadclients provides a unified interface for interacting with
// download clients (both torrent and usenet).
package downloadclient

import "context"

// DownloadItem represents a generic download item across all client types.
type DownloadItem struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	TotalSize     int64   `json:"total_size"`
	RemainingSize int64   `json:"remaining_size"`
	Progress      float64 `json:"progress"`       // 0.0 to 1.0
	DownloadSpeed int64   `json:"download_speed"` // bytes/sec
	UploadSpeed   int64   `json:"upload_speed"`   // bytes/sec (torrent only)
	ETA           int64   `json:"eta"`            // seconds
	Category      string  `json:"category"`
	SavePath      string  `json:"save_path"`
	Ratio         float64 `json:"ratio"`        // upload/download ratio (torrent only)
	SeedingTime   int64   `json:"seeding_time"` // seconds spent seeding (torrent only)
	CompletedAt   int64   `json:"completed_at"` // unix timestamp of completion
	AddedAt       int64   `json:"added_at"`     // unix timestamp when added
}

// ClientStatus represents the overall status of a download client.
type ClientStatus struct {
	Version       string `json:"version"`
	DownloadSpeed int64  `json:"download_speed"` // bytes/sec
	UploadSpeed   int64  `json:"upload_speed"`   // bytes/sec
	Paused        bool   `json:"paused"`
	ItemCount     int    `json:"item_count"`
}

// Client is the unified interface for all download clients.
type Client interface {
	// GetItems returns all download items (queue + active).
	GetItems(ctx context.Context) ([]DownloadItem, error)
	// GetHistory returns completed/historical items no longer actively downloading.
	GetHistory(ctx context.Context) ([]DownloadItem, error)
	// PauseAll pauses all downloads.
	PauseAll(ctx context.Context) error
	// ResumeAll resumes all downloads.
	ResumeAll(ctx context.Context) error
	// RemoveItem removes a download by its ID. If deleteData is true, downloaded files are also removed.
	RemoveItem(ctx context.Context, id string, deleteData bool) error
	// GetStatus returns overall client status.
	GetStatus(ctx context.Context) (*ClientStatus, error)
	// TestConnection verifies the client is reachable.
	TestConnection(ctx context.Context) (string, error)
}

// ClientType represents the type of download client.
type ClientType string

const (
	TypeSABnzbd      ClientType = "sabnzbd"
	TypeNZBGet       ClientType = "nzbget"
	TypeQBittorrent  ClientType = "qbittorrent"
	TypeTransmission ClientType = "transmission"
	TypeDeluge       ClientType = "deluge"
)
