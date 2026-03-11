package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const csrfHeader = "X-Transmission-Session-Id"

// Client is a Transmission RPC client.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client

	mu        sync.Mutex
	csrfToken string
}

// NewClient creates a new Transmission RPC client.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// rpcRequest represents a Transmission RPC request body.
type rpcRequest struct {
	Method    string      `json:"method"`
	Arguments interface{} `json:"arguments,omitempty"`
	Tag       int         `json:"tag,omitempty"`
}

// rpcResponse represents a Transmission RPC response body.
type rpcResponse struct {
	Result    string          `json:"result"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Tag       int             `json:"tag,omitempty"`
}

// doRPC performs an RPC call to Transmission with CSRF token handling.
func (c *Client) doRPC(ctx context.Context, method string, args interface{}) (json.RawMessage, error) {
	payload := rpcRequest{Method: method, Arguments: args}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	rpcURL := c.BaseURL + "/transmission/rpc"

	resp, err := c.sendRPC(ctx, rpcURL, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Transmission returns 409 with the CSRF token on first request.
	if resp.StatusCode == http.StatusConflict {
		_ = resp.Body.Close()
		token := resp.Header.Get(csrfHeader)
		if token == "" {
			return nil, fmt.Errorf("received 409 but no %s header", csrfHeader)
		}
		c.mu.Lock()
		c.csrfToken = token
		c.mu.Unlock()

		resp, err = c.sendRPC(ctx, rpcURL, body)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: check username/password")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("rpc error %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Result != "success" {
		return nil, fmt.Errorf("rpc result: %s", rpcResp.Result)
	}
	return rpcResp.Arguments, nil
}

func (c *Client) sendRPC(ctx context.Context, rpcURL string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
	c.mu.Lock()
	token := c.csrfToken
	c.mu.Unlock()
	if token != "" {
		req.Header.Set(csrfHeader, token)
	}
	return c.HTTPClient.Do(req)
}

// Torrent represents a Transmission torrent.
type Torrent struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	HashString    string  `json:"hashString"`
	Status        int     `json:"status"`
	TotalSize     int64   `json:"totalSize"`
	PercentDone   float64 `json:"percentDone"`
	RateDownload  int64   `json:"rateDownload"`
	RateUpload    int64   `json:"rateUpload"`
	UploadRatio   float64 `json:"uploadRatio"`
	ETA           int64   `json:"eta"`
	SeedRatioMode int     `json:"seedRatioMode"`
	AddedDate     int64   `json:"addedDate"`
	DoneDate      int64   `json:"doneDate"`
	DownloadDir   string  `json:"downloadDir"`
	SizeWhenDone  int64   `json:"sizeWhenDone"`
	LeftUntilDone int64   `json:"leftUntilDone"`
	Error         int     `json:"error"`
	ErrorString   string  `json:"errorString"`
}

// Transmission torrent status codes.
const (
	StatusStopped      = 0
	StatusCheckWait    = 1
	StatusChecking     = 2
	StatusDownloadWait = 3
	StatusDownloading  = 4
	StatusSeedWait     = 5
	StatusSeeding      = 6
)

// SessionStats represents Transmission session statistics.
type SessionStats struct {
	DownloadSpeed int64 `json:"downloadSpeed"`
	UploadSpeed   int64 `json:"uploadSpeed"`
	TorrentCount  int   `json:"torrentCount"`
	ActiveCount   int   `json:"activeTorrentCount"`
	PausedCount   int   `json:"pausedTorrentCount"`
}

var defaultFields = []string{
	"id", "name", "hashString", "status", "totalSize", "percentDone",
	"rateDownload", "rateUpload", "uploadRatio", "eta", "addedDate",
	"doneDate", "downloadDir", "sizeWhenDone", "leftUntilDone",
	"error", "errorString",
}

// GetTorrents returns all torrents with default fields.
func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	args := map[string]interface{}{
		"fields": defaultFields,
	}
	raw, err := c.doRPC(ctx, "torrent-get", args)
	if err != nil {
		return nil, err
	}
	var result struct {
		Torrents []Torrent `json:"torrents"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode torrents: %w", err)
	}
	return result.Torrents, nil
}

// PauseTorrents pauses the specified torrents by ID.
func (c *Client) PauseTorrents(ctx context.Context, ids []int) error {
	args := map[string]interface{}{"ids": ids}
	_, err := c.doRPC(ctx, "torrent-stop", args)
	return err
}

// ResumeTorrents starts/resumes the specified torrents by ID.
func (c *Client) ResumeTorrents(ctx context.Context, ids []int) error {
	args := map[string]interface{}{"ids": ids}
	_, err := c.doRPC(ctx, "torrent-start", args)
	return err
}

// DeleteTorrents removes the specified torrents. If deleteData is true, local data is also removed.
func (c *Client) DeleteTorrents(ctx context.Context, ids []int, deleteData bool) error {
	args := map[string]interface{}{
		"ids":               ids,
		"delete-local-data": deleteData,
	}
	_, err := c.doRPC(ctx, "torrent-remove", args)
	return err
}

// GetSessionStats returns Transmission session statistics.
func (c *Client) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	raw, err := c.doRPC(ctx, "session-stats", nil)
	if err != nil {
		return nil, err
	}
	var stats SessionStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, fmt.Errorf("decode session stats: %w", err)
	}
	return &stats, nil
}

// GetVersion returns the Transmission version string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	raw, err := c.doRPC(ctx, "session-get", nil)
	if err != nil {
		return "", err
	}
	var session struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &session); err != nil {
		return "", fmt.Errorf("decode session: %w", err)
	}
	return session.Version, nil
}

// TestConnection verifies Transmission is reachable.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}

// AddTorrentByURL adds a torrent via magnet URI or HTTP URL.
func (c *Client) AddTorrentByURL(ctx context.Context, torrentURL, downloadDir string) error {
	args := map[string]interface{}{
		"filename": torrentURL,
	}
	if downloadDir != "" {
		args["download-dir"] = downloadDir
	}
	_, err := c.doRPC(ctx, "torrent-add", args)
	return err
}
