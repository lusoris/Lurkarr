package arrclient

import (
	"bytes"
	"context"
	"crypto/tls"
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
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
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

func (c *Client) put(ctx context.Context, path string, body io.Reader) error {
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
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
	ArtistID              int             `json:"artistId,omitempty"`
	BookID                int             `json:"bookId,omitempty"`
	AuthorID              int             `json:"authorId,omitempty"`
	Quality               *QualityInfo    `json:"quality,omitempty"`

	// Enriched fields (populated when queue is fetched with include* params).
	Movie   *QueueMovie   `json:"movie,omitempty"`
	Series  *QueueSeries  `json:"series,omitempty"`
	Episode *QueueEpisode `json:"episode,omitempty"`
	Album   *QueueAlbum   `json:"album,omitempty"`
	Book    *QueueBook    `json:"book,omitempty"`
}

// QueueMovie holds enriched movie data from Radarr/Eros queue responses.
type QueueMovie struct {
	TmdbID int    `json:"tmdbId"`
	ImdbID string `json:"imdbId"`
	Title  string `json:"title"`
	Tags   []int  `json:"tags"`
}

// QueueSeries holds enriched series data from Sonarr/Whisparr queue responses.
type QueueSeries struct {
	TvdbID int    `json:"tvdbId"`
	Title  string `json:"title"`
	Tags   []int  `json:"tags"`
}

// QueueEpisode holds enriched episode data from Sonarr/Whisparr queue responses.
type QueueEpisode struct {
	SeasonNumber  int `json:"seasonNumber"`
	EpisodeNumber int `json:"episodeNumber"`
}

// QueueAlbum holds enriched album data from Lidarr queue responses.
type QueueAlbum struct {
	ForeignAlbumID string `json:"foreignAlbumId"`
	Title          string `json:"title"`
	Tags           []int  `json:"tags"`
}

// QueueBook holds enriched book data from Readarr queue responses.
type QueueBook struct {
	ForeignBookID string `json:"foreignBookId"`
	Title         string `json:"title"`
	Tags          []int  `json:"tags"`
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

// TaggableMediaID returns the ID of the entity that supports tags in the *arr API.
// Radarr/Eros: MovieID, Sonarr/Whisparr: SeriesID, Lidarr: ArtistID, Readarr: AuthorID.
func (q *QueueRecord) TaggableMediaID() int {
	if q.MovieID > 0 {
		return q.MovieID
	}
	if q.SeriesID > 0 {
		return q.SeriesID
	}
	if q.ArtistID > 0 {
		return q.ArtistID
	}
	if q.AuthorID > 0 {
		return q.AuthorID
	}
	return 0
}

// MediaKey returns a cross-instance media identifier from enriched queue data.
// For Radarr/Eros: "tmdb:12345"
// For Sonarr/Whisparr: "tvdb:67890:s02" (series + season)
// For Lidarr: "album:foreignId"
// For Readarr: "book:foreignId"
// Returns empty string if enriched data is not available.
func (q *QueueRecord) MediaKey() string {
	if q.Movie != nil && q.Movie.TmdbID > 0 {
		return fmt.Sprintf("tmdb:%d", q.Movie.TmdbID)
	}
	if q.Series != nil && q.Series.TvdbID > 0 && q.Episode != nil {
		return fmt.Sprintf("tvdb:%d:s%02d", q.Series.TvdbID, q.Episode.SeasonNumber)
	}
	if q.Album != nil && q.Album.ForeignAlbumID != "" {
		return "album:" + q.Album.ForeignAlbumID
	}
	if q.Book != nil && q.Book.ForeignBookID != "" {
		return "book:" + q.Book.ForeignBookID
	}
	return ""
}

// MediaTags returns the tag IDs from the enriched media item, or nil if unavailable.
func (q *QueueRecord) MediaTags() []int {
	if q.Movie != nil {
		return q.Movie.Tags
	}
	if q.Series != nil {
		return q.Series.Tags
	}
	if q.Album != nil {
		return q.Album.Tags
	}
	if q.Book != nil {
		return q.Book.Tags
	}
	return nil
}

// Tag represents a tag from an *arr instance.
type Tag struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

// GetTags returns all tags from the *arr instance.
func (c *Client) GetTags(ctx context.Context, apiVersion string) ([]Tag, error) {
	var tags []Tag
	if err := c.get(ctx, "/api/"+apiVersion+"/tag", &tags); err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}
	return tags, nil
}

// CreateTag creates a new tag and returns it.
func (c *Client) CreateTag(ctx context.Context, apiVersion, label string) (*Tag, error) {
	body, _ := json.Marshal(map[string]string{"label": label})
	var tag Tag
	if err := c.post(ctx, "/api/"+apiVersion+"/tag", bytes.NewReader(body), &tag); err != nil {
		return nil, fmt.Errorf("create tag %q: %w", label, err)
	}
	return &tag, nil
}

// TagMedia applies a tag to a media item using the bulk editor endpoint.
// appType determines the media path (movie/series/artist/author).
func (c *Client) TagMedia(ctx context.Context, apiVersion, appType string, mediaID, tagID int) error {
	var path string
	var idsKey string
	switch appType {
	case "radarr", "eros":
		path = "/api/" + apiVersion + "/movie/editor"
		idsKey = "movieIds"
	case "sonarr", "whisparr":
		path = "/api/" + apiVersion + "/series/editor"
		idsKey = "seriesIds"
	case "lidarr":
		path = "/api/" + apiVersion + "/artist/editor"
		idsKey = "artistIds"
	case "readarr":
		path = "/api/" + apiVersion + "/author/editor"
		idsKey = "authorIds"
	default:
		return fmt.Errorf("unsupported app type for tagging: %s", appType)
	}
	body, _ := json.Marshal(map[string]any{
		idsKey:     []int{mediaID},
		"tags":     []int{tagID},
		"applyTags": "add",
	})
	return c.put(ctx, path, bytes.NewReader(body))
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
