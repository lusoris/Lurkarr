// Package utorrent provides an HTTP client for the uTorrent/BitTorrent WebUI API.
package utorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Client communicates with the uTorrent WebUI API.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client

	mu    sync.Mutex
	token string
}

// NewClient creates a new uTorrent API client.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

// Torrent represents a parsed uTorrent torrent from the positional array.
type Torrent struct {
	Hash          string
	Status        int // bitmask: 1=started, 2=checking, 16=paused, 32=queued, 64=loaded, 128=error
	Name          string
	Size          int64
	Progress      int // permille (0-1000)
	Downloaded    int64
	Uploaded      int64
	Ratio         int // permille
	UploadSpeed   int64
	DownloadSpeed int64
	ETA           int64
	Label         string
	AddedOn       int64  // field index 23
	CompletedOn   int64  // field index 24
	SavePath      string // field index 26
}

var tokenRegexp = regexp.MustCompile(`<div[^>]*id=['"]token['"][^>]*>([^<]+)</div>`)

// fetchToken gets the CSRF token from /gui/token.html.
func (c *Client) fetchToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/gui/token.html", http.NoBody)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return fmt.Errorf("utorrent: read token response: %w", err)
	}
	matches := tokenRegexp.FindSubmatch(body)
	if len(matches) < 2 {
		return fmt.Errorf("utorrent: token not found in response")
	}

	c.mu.Lock()
	c.token = string(matches[1])
	c.mu.Unlock()
	return nil
}

// apiGet performs an authenticated GET to /gui/ with the given params.
func (c *Client) apiGet(ctx context.Context, params url.Values) ([]byte, error) {
	c.mu.Lock()
	tok := c.token
	c.mu.Unlock()

	if tok == "" {
		if err := c.fetchToken(ctx); err != nil {
			return nil, err
		}
		c.mu.Lock()
		tok = c.token
		c.mu.Unlock()
	}

	params.Set("token", tok)
	reqURL := c.BaseURL + "/gui/?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	// Token expired — refetch once and retry.
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx); err != nil {
			return nil, err
		}
		c.mu.Lock()
		params.Set("token", c.token)
		c.mu.Unlock()

		retryURL := c.BaseURL + "/gui/?" + params.Encode()
		req2, err := http.NewRequestWithContext(ctx, http.MethodGet, retryURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		req2.SetBasicAuth(c.Username, c.Password)

		resp2, err := c.HTTPClient.Do(req2)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()

		body, err = io.ReadAll(io.LimitReader(resp2.Body, 1<<20))
		if err != nil {
			return nil, err
		}
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("utorrent api error %d: %s", resp2.StatusCode, body)
		}
		return body, nil
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("utorrent api error %d: %s", resp.StatusCode, body)
	}
	return body, nil
}

// parseTorrent converts a positional JSON array to a Torrent struct.
func parseTorrent(raw []json.RawMessage) (Torrent, error) {
	if len(raw) < 21 {
		return Torrent{}, fmt.Errorf("utorrent: torrent array too short (%d)", len(raw))
	}
	var t Torrent
	if err := json.Unmarshal(raw[0], &t.Hash); err != nil {
		return Torrent{}, fmt.Errorf("utorrent: unmarshal hash: %w", err)
	}
	if err := json.Unmarshal(raw[2], &t.Name); err != nil {
		return Torrent{}, fmt.Errorf("utorrent: unmarshal name: %w", err)
	}
	// Non-critical fields: log warnings on parse failure instead of silently discarding.
	for _, f := range []struct {
		idx  int
		name string
		dest any
	}{
		{1, "status", &t.Status},
		{3, "size", &t.Size},
		{4, "progress", &t.Progress},
		{5, "downloaded", &t.Downloaded},
		{6, "uploaded", &t.Uploaded},
		{7, "ratio", &t.Ratio},
		{8, "upload_speed", &t.UploadSpeed},
		{9, "download_speed", &t.DownloadSpeed},
		{10, "eta", &t.ETA},
		{11, "label", &t.Label},
	} {
		if err := json.Unmarshal(raw[f.idx], f.dest); err != nil {
			slog.Warn("utorrent: failed to parse torrent field",
				"field", f.name, "index", f.idx, "hash", t.Hash, "error", err)
		}
	}
	if len(raw) > 23 {
		if err := json.Unmarshal(raw[23], &t.AddedOn); err != nil {
			slog.Warn("utorrent: failed to parse torrent field",
				"field", "added_on", "index", 23, "hash", t.Hash, "error", err)
		}
	}
	if len(raw) > 24 {
		if err := json.Unmarshal(raw[24], &t.CompletedOn); err != nil {
			slog.Warn("utorrent: failed to parse torrent field",
				"field", "completed_on", "index", 24, "hash", t.Hash, "error", err)
		}
	}
	if len(raw) > 26 {
		if err := json.Unmarshal(raw[26], &t.SavePath); err != nil {
			slog.Warn("utorrent: failed to parse torrent field",
				"field", "save_path", "index", 26, "hash", t.Hash, "error", err)
		}
	}
	return t, nil
}

// GetTorrents returns all torrents.
func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	body, err := c.apiGet(ctx, url.Values{"list": {"1"}})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Torrents [][]json.RawMessage `json:"torrents"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("utorrent: unmarshal list: %w", err)
	}
	torrents := make([]Torrent, 0, len(resp.Torrents))
	for _, raw := range resp.Torrents {
		t, err := parseTorrent(raw)
		if err != nil {
			continue
		}
		torrents = append(torrents, t)
	}
	return torrents, nil
}

func (c *Client) doAction(ctx context.Context, action, hash string) error {
	_, err := c.apiGet(ctx, url.Values{"action": {action}, "hash": {hash}})
	return err
}

// PauseTorrent pauses a single torrent.
func (c *Client) PauseTorrent(ctx context.Context, hash string) error {
	return c.doAction(ctx, "pause", hash)
}

// UnpauseTorrent resumes a single torrent.
func (c *Client) UnpauseTorrent(ctx context.Context, hash string) error {
	return c.doAction(ctx, "unpause", hash)
}

// RecheckTorrent triggers a data integrity recheck.
func (c *Client) RecheckTorrent(ctx context.Context, hash string) error {
	return c.doAction(ctx, "recheck", hash)
}

// RemoveTorrent removes a torrent, optionally deleting its data.
func (c *Client) RemoveTorrent(ctx context.Context, hash string, deleteData bool) error {
	action := "remove"
	if deleteData {
		action = "removedatatorrent"
	}
	return c.doAction(ctx, action, hash)
}

// GetVersion returns the uTorrent build version string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	body, err := c.apiGet(ctx, url.Values{"list": {"1"}})
	if err != nil {
		return "", err
	}
	var resp struct {
		Build int `json:"build"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("utorrent: unmarshal version: %w", err)
	}
	return fmt.Sprintf("uTorrent (build %d)", resp.Build), nil
}

// TestConnection verifies the connection by fetching the version.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}
