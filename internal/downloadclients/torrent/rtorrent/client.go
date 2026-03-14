// Package rtorrent provides a thin wrapper around github.com/autobrr/go-rtorrent
// with a convenience constructor matching the project's download client patterns.
package rtorrent

import (
	"context"
	"net/http"
	"strings"
	"time"

	gort "github.com/autobrr/go-rtorrent"
)

// Client wraps the autobrr go-rtorrent client.
type Client struct {
	rt *gort.Client
}

// NewClient creates a new rTorrent client.
// baseURL should be the XML-RPC endpoint (e.g. http://rtorrent:8080/RPC2).
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	cfg := gort.Config{
		Addr:      strings.TrimRight(baseURL, "/"),
		BasicUser: username,
		BasicPass: password,
	}
	rt := gort.NewClientWithOpts(cfg, gort.WithCustomClient(&http.Client{Timeout: timeout}))
	return &Client{rt: rt}
}

// GetTorrents returns all torrents.
func (c *Client) GetTorrents(ctx context.Context) ([]gort.Torrent, error) {
	return c.rt.GetTorrents(ctx, gort.ViewMain)
}

// GetTorrent returns a single torrent by info hash.
func (c *Client) GetTorrent(ctx context.Context, hash string) (gort.Torrent, error) {
	return c.rt.GetTorrent(ctx, hash)
}

// GetStatus returns the status of a torrent (speeds, completion, etc.).
func (c *Client) GetStatus(ctx context.Context, t gort.Torrent) (gort.Status, error) {
	return c.rt.GetStatus(ctx, t)
}

// IsActive returns whether the torrent is active.
func (c *Client) IsActive(ctx context.Context, t gort.Torrent) (bool, error) {
	return c.rt.IsActive(ctx, t)
}

// PauseTorrent pauses a torrent.
func (c *Client) PauseTorrent(ctx context.Context, t gort.Torrent) error {
	return c.rt.PauseTorrent(ctx, t)
}

// ResumeTorrent resumes a paused torrent.
func (c *Client) ResumeTorrent(ctx context.Context, t gort.Torrent) error {
	return c.rt.ResumeTorrent(ctx, t)
}

// StopTorrent stops a torrent.
func (c *Client) StopTorrent(ctx context.Context, t gort.Torrent) error {
	return c.rt.StopTorrent(ctx, t)
}

// StartTorrent starts a torrent.
func (c *Client) StartTorrent(ctx context.Context, t gort.Torrent) error {
	return c.rt.StartTorrent(ctx, t)
}

// Delete removes a torrent from rTorrent.
func (c *Client) Delete(ctx context.Context, t gort.Torrent) error {
	return c.rt.Delete(ctx, t)
}

// DownRate returns the global download rate in bytes/sec.
func (c *Client) DownRate(ctx context.Context) (int, error) {
	return c.rt.DownRate(ctx)
}

// UpRate returns the global upload rate in bytes/sec.
func (c *Client) UpRate(ctx context.Context) (int, error) {
	return c.rt.UpRate(ctx)
}

// Name returns the rTorrent instance name/version.
func (c *Client) Name(ctx context.Context) (string, error) {
	return c.rt.Name(ctx)
}
