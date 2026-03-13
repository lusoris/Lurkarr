package arrclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a generic *Arr API client.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new *Arr API client.
func NewClient(baseURL, apiKey string, timeout time.Duration, sslVerify bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !sslVerify {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(errBody))
	}
	return resp, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(ctx context.Context, path string, body io.Reader, result any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) del(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// DeleteQueueItem removes an item from the arr queue.
func (c *Client) DeleteQueueItem(ctx context.Context, apiVersion string, queueID int, removeFromClient, blocklist bool) error {
	path := fmt.Sprintf("/api/%s/queue/%d?removeFromClient=%t&blocklist=%t",
		apiVersion, queueID, removeFromClient, blocklist)
	return c.del(ctx, path)
}

// ManualImportItem represents a file available for manual import.
type ManualImportItem struct {
	ID                int          `json:"id"`
	Path              string       `json:"path"`
	Name              string       `json:"name"`
	Size              int64        `json:"size"`
	Quality           *QualityInfo `json:"quality,omitempty"`
	CustomFormatScore int          `json:"customFormatScore"`
	Rejections        []struct {
		Reason string `json:"reason"`
	} `json:"rejections"`
}

// GetManualImport lists files available for manual import for a download ID.
func (c *Client) GetManualImport(ctx context.Context, apiVersion, downloadID string) ([]ManualImportItem, error) {
	var items []ManualImportItem
	path := fmt.Sprintf("/api/%s/manualimport?downloadId=%s", apiVersion, downloadID)
	if err := c.get(ctx, path, &items); err != nil {
		return nil, fmt.Errorf("get manual import: %w", err)
	}
	return items, nil
}

// SystemStatus represents the status response from any *Arr app.
type SystemStatus struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}

// TestConnection verifies connectivity and returns system status.
func (c *Client) TestConnection(ctx context.Context, apiVersion string) (*SystemStatus, error) {
	var status SystemStatus
	if err := c.get(ctx, "/api/"+apiVersion+"/system/status", &status); err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}
	return &status, nil
}

// HealthCheck represents a single health check entry from an *Arr app.
type HealthCheck struct {
	Source  string `json:"source"`
	Type    string `json:"type"` // "ok", "notice", "warning", "error"
	Message string `json:"message"`
	WikiURL string `json:"wikiUrl"`
}

// GetHealth returns the health checks for the instance.
func (c *Client) GetHealth(ctx context.Context, apiVersion string) ([]HealthCheck, error) {
	var checks []HealthCheck
	if err := c.get(ctx, "/api/"+apiVersion+"/health", &checks); err != nil {
		return nil, fmt.Errorf("get health: %w", err)
	}
	return checks, nil
}

// DiskSpace represents disk space information from an *Arr app.
type DiskSpace struct {
	Path       string `json:"path"`
	Label      string `json:"label"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
}

// GetDiskSpace returns disk space information for the instance.
func (c *Client) GetDiskSpace(ctx context.Context, apiVersion string) ([]DiskSpace, error) {
	var disks []DiskSpace
	if err := c.get(ctx, "/api/"+apiVersion+"/diskspace", &disks); err != nil {
		return nil, fmt.Errorf("get disk space: %w", err)
	}
	return disks, nil
}

// APIVersionFor returns the API version string for a given app type.
// Lidarr, Readarr, and Prowlarr use v1; all others use v3.
func APIVersionFor(appType string) string {
	switch appType {
	case "lidarr", "readarr", "prowlarr":
		return "v1"
	default:
		return "v3"
	}
}

// CommandResponse is the response from a command endpoint.
type CommandResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// QueueRecord represents an item in the download queue.
type QueueRecord struct {
	ID                    int             `json:"id"`
	DownloadID            string          `json:"downloadId"`
	Status                string          `json:"status"`
	Title                 string          `json:"title"`
	Size                  int64           `json:"size"`
	Sizeleft              int64           `json:"sizeleft"`
	TimeleftStr           string          `json:"timeleft"`
	Protocol              string          `json:"protocol"`
	Indexer               string          `json:"indexer"`
	IndexerFlags          int             `json:"indexerFlags"`
	DownloadClient        string          `json:"downloadClient"`
	TrackedDownloadStatus string          `json:"trackedDownloadStatus"`
	TrackedDownloadState  string          `json:"trackedDownloadState"`
	StatusMessages        []StatusMessage `json:"statusMessages"`
	CustomFormatScore     int             `json:"customFormatScore"`
	MovieID               int             `json:"movieId,omitempty"`
	SeriesID              int             `json:"seriesId,omitempty"`
	EpisodeID             int             `json:"episodeId,omitempty"`
	AlbumID               int             `json:"albumId,omitempty"`
	BookID                int             `json:"bookId,omitempty"`
	Quality               *QualityInfo    `json:"quality,omitempty"`
}

// StatusMessage from arr queue items for detecting import issues.
type StatusMessage struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

// QualityInfo holds quality profile data from the queue.
type QualityInfo struct {
	Quality struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"quality"`
	Revision struct {
		Version int `json:"version"`
	} `json:"revision"`
}

// HasImportError checks if this queue item has an import problem.
func (q *QueueRecord) HasImportError() bool {
	return q.TrackedDownloadStatus == "warning" || q.TrackedDownloadState == "importPending"
}

// MediaID returns the media ID for this queue record based on which field is set.
func (q *QueueRecord) MediaID() int {
	if q.MovieID > 0 {
		return q.MovieID
	}
	if q.EpisodeID > 0 {
		return q.EpisodeID
	}
	if q.AlbumID > 0 {
		return q.AlbumID
	}
	if q.BookID > 0 {
		return q.BookID
	}
	if q.SeriesID > 0 {
		return q.SeriesID
	}
	return 0
}

// QueueResponse wraps a paginated queue response.
type QueueResponse struct {
	TotalRecords int           `json:"totalRecords"`
	Records      []QueueRecord `json:"records"`
}

// GetQueue returns the download queue for any arr instance.
func (c *Client) GetQueue(ctx context.Context, apiVersion string) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, "/api/"+apiVersion+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("get queue: %w", err)
	}
	return &resp, nil
}

// IsPrivateIP checks if a URL points to a private/loopback address.
// Used for SSRF protection in test-connection.
func IsPrivateIP(rawURL string) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("parse url: %w", err)
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		// Resolve hostname
		addrs, err := net.LookupHost(host)
		if err != nil {
			return false, fmt.Errorf("resolve host: %w", err)
		}
		for _, addr := range addrs {
			resolved := net.ParseIP(addr)
			if resolved != nil && (resolved.IsLoopback() || resolved.IsPrivate() || resolved.IsLinkLocalUnicast()) {
				return true, nil
			}
		}
		return false, nil
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast(), nil
}
