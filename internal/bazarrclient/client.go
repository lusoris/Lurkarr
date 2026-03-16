// Package bazarrclient provides an HTTP client for the Bazarr subtitle manager API.
package bazarrclient

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

// Client communicates with a Bazarr instance.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Bazarr API client.
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

// SystemStatus is the response from /api/system/status.
type SystemStatus struct {
	Version   string `json:"bazarr_version"`
	StartTime string `json:"start_time"`
}

// HealthItem is a single health check result.
type HealthItem struct {
	Object  string `json:"object"`
	Issue   string `json:"issue"`
	Message string `json:"message,omitempty"`
}

// WantedResponse is the paginated response for wanted episodes/movies.
type WantedResponse struct {
	Data     []WantedItem `json:"data"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// WantedItem represents an episode or movie missing subtitles.
type WantedItem struct {
	Title           string   `json:"title"`
	SeriesTitle     string   `json:"seriesTitle,omitempty"` // episodes only
	EpisodeNumber   string   `json:"episode_number,omitempty"`
	SeasonNumber    int      `json:"season,omitempty"`
	Missing         []string `json:"missing_subtitles"`
	SonarrSeriesID  int      `json:"sonarrSeriesId,omitempty"`
	SonarrEpisodeID int      `json:"sonarrEpisodeId,omitempty"`
	RadarrID        int      `json:"radarrId,omitempty"`
}

// HistoryItem represents a subtitle download from history.
type HistoryItem struct {
	Action      int    `json:"action"`
	Title       string `json:"title"`
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
	Provider    string `json:"provider"`
	Language    string `json:"language"`
	Score       string `json:"score,omitempty"`
}

// HistoryResponse is the paginated history response.
type HistoryResponse struct {
	Data     []HistoryItem `json:"data"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
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

// TestConnection tests connectivity by hitting the system status endpoint.
func (c *Client) TestConnection(ctx context.Context) (*SystemStatus, error) {
	var status SystemStatus
	if err := c.get(ctx, "/api/system/status", &status); err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}
	return &status, nil
}

// GetHealth returns Bazarr health check items.
func (c *Client) GetHealth(ctx context.Context) ([]HealthItem, error) {
	var resp struct {
		Data []HealthItem `json:"data"`
	}
	if err := c.get(ctx, "/api/system/health", &resp); err != nil {
		return nil, fmt.Errorf("get health: %w", err)
	}
	return resp.Data, nil
}

// GetWantedEpisodes returns episodes missing subtitles.
func (c *Client) GetWantedEpisodes(ctx context.Context) (*WantedResponse, error) {
	var resp WantedResponse
	if err := c.get(ctx, "/api/episodes/wanted?page=1&length=20", &resp); err != nil {
		return nil, fmt.Errorf("get wanted episodes: %w", err)
	}
	return &resp, nil
}

// GetWantedMovies returns movies missing subtitles.
func (c *Client) GetWantedMovies(ctx context.Context) (*WantedResponse, error) {
	var resp WantedResponse
	if err := c.get(ctx, "/api/movies/wanted?page=1&length=20", &resp); err != nil {
		return nil, fmt.Errorf("get wanted movies: %w", err)
	}
	return &resp, nil
}

// GetEpisodeHistory returns recent episode subtitle history.
func (c *Client) GetEpisodeHistory(ctx context.Context) (*HistoryResponse, error) {
	var resp HistoryResponse
	if err := c.get(ctx, "/api/history/episodes?page=1&length=20", &resp); err != nil {
		return nil, fmt.Errorf("get episode history: %w", err)
	}
	return &resp, nil
}

// GetMovieHistory returns recent movie subtitle history.
func (c *Client) GetMovieHistory(ctx context.Context) (*HistoryResponse, error) {
	var resp HistoryResponse
	if err := c.get(ctx, "/api/history/movies?page=1&length=20", &resp); err != nil {
		return nil, fmt.Errorf("get movie history: %w", err)
	}
	return &resp, nil
}
