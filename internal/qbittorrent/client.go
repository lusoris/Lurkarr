package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is a qBittorrent WebUI API v2 client.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client

	mu        sync.Mutex
	authenticated bool
}

// NewClient creates a new qBittorrent client.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
	}
}

// login authenticates with qBittorrent and stores the SID cookie.
func (c *Client) login(ctx context.Context) error {
	form := url.Values{
		"username": {c.Username},
		"password": {c.Password},
	}
	reqURL := c.BaseURL + "/api/v2/auth/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return fmt.Errorf("read login response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: status %d: %s", resp.StatusCode, string(body))
	}
	text := strings.TrimSpace(string(body))
	if text != "Ok." {
		return fmt.Errorf("login rejected: %s", text)
	}
	c.authenticated = true
	return nil
}

// ensureAuth performs login if not already authenticated.
func (c *Client) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.authenticated {
		return nil
	}
	return c.login(ctx)
}

// apiGet performs an authenticated GET request and returns the response body.
func (c *Client) apiGet(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	reqURL := c.BaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode == http.StatusForbidden {
		// Session expired — re-authenticate and retry once.
		c.mu.Lock()
		c.authenticated = false
		c.mu.Unlock()
		if err := c.ensureAuth(ctx); err != nil {
			return nil, err
		}
		req2, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		resp2, err := c.HTTPClient.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("retry request: %w", err)
		}
		defer func() { _ = resp2.Body.Close() }()
		body, err = io.ReadAll(io.LimitReader(resp2.Body, 1<<20))
		if err != nil {
			return nil, fmt.Errorf("read retry body: %w", err)
		}
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("api error %d: %s", resp2.StatusCode, string(body))
		}
		return body, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// apiPost performs an authenticated POST request with form data.
func (c *Client) apiPost(ctx context.Context, path string, form url.Values) ([]byte, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	reqURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode == http.StatusForbidden {
		c.mu.Lock()
		c.authenticated = false
		c.mu.Unlock()
		if err := c.ensureAuth(ctx); err != nil {
			return nil, err
		}
		req2, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp2, err := c.HTTPClient.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("retry request: %w", err)
		}
		defer func() { _ = resp2.Body.Close() }()
		body, err = io.ReadAll(io.LimitReader(resp2.Body, 1<<20))
		if err != nil {
			return nil, fmt.Errorf("read retry body: %w", err)
		}
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("api error %d: %s", resp2.StatusCode, string(body))
		}
		return body, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// Torrent represents a torrent in qBittorrent.
type Torrent struct {
	Hash           string  `json:"hash"`
	Name           string  `json:"name"`
	Size           int64   `json:"size"`
	Progress       float64 `json:"progress"`
	DownloadSpeed  int64   `json:"dlspeed"`
	UploadSpeed    int64   `json:"upspeed"`
	Priority       int     `json:"priority"`
	NumSeeds       int     `json:"num_seeds"`
	NumLeechers    int     `json:"num_leechs"`
	Ratio          float64 `json:"ratio"`
	State          string  `json:"state"`
	Category       string  `json:"category"`
	AddedOn        int64   `json:"added_on"`
	CompletionOn   int64   `json:"completion_on"`
	SavePath       string  `json:"save_path"`
	ContentPath    string  `json:"content_path"`
	AmountLeft     int64   `json:"amount_left"`
	TimeActive     int64   `json:"time_active"`
	Tracker        string  `json:"tracker"`
	TotalSize      int64   `json:"total_size"`
	ETA            int64   `json:"eta"`
}

// TransferInfo represents global transfer information.
type TransferInfo struct {
	DownloadSpeed   int64  `json:"dl_info_speed"`
	UploadSpeed     int64  `json:"up_info_speed"`
	DownloadTotal   int64  `json:"dl_info_data"`
	UploadTotal     int64  `json:"up_info_data"`
	DHT             int    `json:"dht_nodes"`
	ConnectionStatus string `json:"connection_status"`
}

// GetTorrents returns all torrents, optionally filtered by state or category.
func (c *Client) GetTorrents(ctx context.Context, filter, category string) ([]Torrent, error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if category != "" {
		params.Set("category", category)
	}
	body, err := c.apiGet(ctx, "/api/v2/torrents/info", params)
	if err != nil {
		return nil, err
	}
	var torrents []Torrent
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, fmt.Errorf("decode torrents: %w", err)
	}
	return torrents, nil
}

// PauseTorrents pauses torrents by hash. Use "all" to pause all.
func (c *Client) PauseTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{"hashes": {strings.Join(hashes, "|")}}
	_, err := c.apiPost(ctx, "/api/v2/torrents/pause", form)
	return err
}

// ResumeTorrents resumes torrents by hash. Use "all" to resume all.
func (c *Client) ResumeTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{"hashes": {strings.Join(hashes, "|")}}
	_, err := c.apiPost(ctx, "/api/v2/torrents/resume", form)
	return err
}

// DeleteTorrents deletes torrents by hash. If deleteFiles is true, downloaded data is also removed.
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	deleteStr := "false"
	if deleteFiles {
		deleteStr = "true"
	}
	form := url.Values{
		"hashes":      {strings.Join(hashes, "|")},
		"deleteFiles": {deleteStr},
	}
	_, err := c.apiPost(ctx, "/api/v2/torrents/delete", form)
	return err
}

// GetTransferInfo returns global transfer stats.
func (c *Client) GetTransferInfo(ctx context.Context) (*TransferInfo, error) {
	body, err := c.apiGet(ctx, "/api/v2/transfer/info", nil)
	if err != nil {
		return nil, err
	}
	var info TransferInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("decode transfer info: %w", err)
	}
	return &info, nil
}

// GetVersion returns the qBittorrent application version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	body, err := c.apiGet(ctx, "/api/v2/app/version", nil)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

// TestConnection verifies qBittorrent is reachable and credentials are valid.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}

// AddTorrentByURL adds a torrent via magnet URI or HTTP URL.
func (c *Client) AddTorrentByURL(ctx context.Context, torrentURL, category, savePath string) error {
	form := url.Values{"urls": {torrentURL}}
	if category != "" {
		form.Set("category", category)
	}
	if savePath != "" {
		form.Set("savepath", savePath)
	}
	_, err := c.apiPost(ctx, "/api/v2/torrents/add", form)
	return err
}
