// Package kapowarrclient provides an HTTP client for the Kapowarr comic manager API.
package kapowarrclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with a Kapowarr instance.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Kapowarr API client.
func NewClient(baseURL, apiKey string, timeout time.Duration, sslVerify bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !sslVerify {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// APIResponse wraps all Kapowarr API responses.
type APIResponse[T any] struct {
	Error  *string `json:"error"`
	Result T       `json:"result"`
}

// AboutInfo is the response from /api/v1/system/about.
type AboutInfo struct {
	Version    string `json:"version"`
	PythonPath string `json:"python_path,omitempty"`
	DataFolder string `json:"data_folder,omitempty"`
}

// VolumeStats is the response from /api/v1/volumes/stats.
type VolumeStats struct {
	Total             int `json:"total"`
	Monitored         int `json:"monitored"`
	Unmonitored       int `json:"unmonitored"`
	Downloaded        int `json:"downloaded"`
	TotalIssues       int `json:"total_issues"`
	DownloadedIssues  int `json:"downloaded_issues"`
	MissingIssues     int `json:"missing_issues"`
	MonitoredIssues   int `json:"monitored_issues"`
	UnmonitoredIssues int `json:"unmonitored_issues"`
}

// QueueItem represents a download in the Kapowarr queue.
type QueueItem struct {
	ID       int     `json:"id"`
	VolumeID int     `json:"volume_id"`
	Status   string  `json:"status"`
	Size     int64   `json:"size"`
	Speed    int64   `json:"speed"`
	Progress float64 `json:"progress"`
}

// TaskInfo represents a running or queued task.
type TaskInfo struct {
	ID       int    `json:"id"`
	Action   string `json:"action"`
	VolumeID int    `json:"volume_id,omitempty"`
	IssueID  int    `json:"issue_id,omitempty"`
}

func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	reqURL := c.baseURL + path
	if strings.Contains(reqURL, "?") {
		reqURL += "&api_key=" + c.apiKey
	} else {
		reqURL += "?api_key=" + c.apiKey
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		errBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return nil, fmt.Errorf("api error %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(errBody))
	}
	return resp, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// TestConnection tests connectivity by hitting the system about endpoint.
func (c *Client) TestConnection(ctx context.Context) (*AboutInfo, error) {
	var resp APIResponse[AboutInfo]
	if err := c.get(ctx, "/api/v1/system/about", &resp); err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("api error: %s", *resp.Error)
	}
	return &resp.Result, nil
}

// GetVolumeStats returns library volume statistics.
func (c *Client) GetVolumeStats(ctx context.Context) (*VolumeStats, error) {
	var resp APIResponse[VolumeStats]
	if err := c.get(ctx, "/api/v1/volumes/stats", &resp); err != nil {
		return nil, fmt.Errorf("get volume stats: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("api error: %s", *resp.Error)
	}
	return &resp.Result, nil
}

// GetQueue returns active downloads.
func (c *Client) GetQueue(ctx context.Context) ([]QueueItem, error) {
	var resp APIResponse[[]QueueItem]
	if err := c.get(ctx, "/api/v1/activity/queue", &resp); err != nil {
		return nil, fmt.Errorf("get queue: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("api error: %s", *resp.Error)
	}
	return resp.Result, nil
}

// GetTasks returns current task queue.
func (c *Client) GetTasks(ctx context.Context) ([]TaskInfo, error) {
	var resp APIResponse[[]TaskInfo]
	if err := c.get(ctx, "/api/v1/system/tasks", &resp); err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("api error: %s", *resp.Error)
	}
	return resp.Result, nil
}
