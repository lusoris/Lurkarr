package sabnzbd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a SABnzbd API client.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new SABnzbd client.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) apiCall(ctx context.Context, mode string, extra url.Values) (json.RawMessage, error) {
	params := url.Values{
		"apikey": {c.APIKey},
		"mode":   {mode},
		"output": {"json"},
	}
	for k, v := range extra {
		params[k] = v
	}
	reqURL := c.BaseURL + "/api?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}

// QueueSlot represents an item in the SABnzbd download queue.
type QueueSlot struct {
	NzoID      string  `json:"nzo_id"`
	Filename   string  `json:"filename"`
	Status     string  `json:"status"`
	MB         string  `json:"mb"`
	MBLeft     string  `json:"mbleft"`
	Percentage string  `json:"percentage"`
	TimeLeft   string  `json:"timeleft"`
	Category   string  `json:"cat"`
}

// Queue represents the SABnzbd download queue.
type Queue struct {
	Status      string      `json:"status"`
	SpeedLimit  string      `json:"speedlimit"`
	Speed       string      `json:"speed"`
	SizeLeft    string      `json:"sizeleft"`
	NoOfSlots   int         `json:"noofslots_total"`
	Slots       []QueueSlot `json:"slots"`
	Paused      bool        `json:"paused"`
}

// HistorySlot represents a completed download.
type HistorySlot struct {
	NzoID       string `json:"nzo_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Size        string `json:"size"`
	Category    string `json:"category"`
	CompletedAt int64  `json:"completed"`
	FailMessage string `json:"fail_message"`
}

// History represents the SABnzbd download history.
type History struct {
	TotalSize  string        `json:"total_size"`
	NoOfSlots  int           `json:"noofslots"`
	Slots      []HistorySlot `json:"slots"`
}

// ServerStats represents SABnzbd server statistics.
type ServerStats struct {
	Total int64            `json:"total"`
	Day   int64            `json:"day"`
	Week  int64            `json:"week"`
	Month int64            `json:"month"`
	Servers map[string]struct {
		Total int64 `json:"total"`
		Day   int64 `json:"day"`
		Week  int64 `json:"week"`
		Month int64 `json:"month"`
	} `json:"servers"`
}

// Version info from SABnzbd.
type VersionInfo struct {
	Version string `json:"version"`
}

// GetQueue returns the current download queue.
func (c *Client) GetQueue(ctx context.Context) (*Queue, error) {
	raw, err := c.apiCall(ctx, "queue", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Queue Queue `json:"queue"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode queue: %w", err)
	}
	return &resp.Queue, nil
}

// GetHistory returns completed downloads.
func (c *Client) GetHistory(ctx context.Context, limit int) (*History, error) {
	raw, err := c.apiCall(ctx, "history", url.Values{"limit": {fmt.Sprintf("%d", limit)}})
	if err != nil {
		return nil, err
	}
	var resp struct {
		History History `json:"history"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}
	return &resp.History, nil
}

// GetServerStats returns download statistics.
func (c *Client) GetServerStats(ctx context.Context) (*ServerStats, error) {
	raw, err := c.apiCall(ctx, "server_stats", nil)
	if err != nil {
		return nil, err
	}
	var stats ServerStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, fmt.Errorf("decode server stats: %w", err)
	}
	return &stats, nil
}

// Pause pauses the download queue.
func (c *Client) Pause(ctx context.Context) error {
	_, err := c.apiCall(ctx, "pause", nil)
	return err
}

// Resume resumes the download queue.
func (c *Client) Resume(ctx context.Context) error {
	_, err := c.apiCall(ctx, "resume", nil)
	return err
}

// GetVersion returns the SABnzbd version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	raw, err := c.apiCall(ctx, "version", nil)
	if err != nil {
		return "", err
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", fmt.Errorf("decode version: %w", err)
	}
	return v, nil
}

// TestConnection verifies SABnzbd is reachable.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}
