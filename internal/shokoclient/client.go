// Package shokoclient provides an HTTP client for the Shoko anime server API.
package shokoclient

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

// Client communicates with a Shoko Server instance.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Shoko API client.
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

// VersionInfo is the server component from /api/v3/Init/Version.
type VersionInfo struct {
	Server struct {
		Version string `json:"Version"`
		Tag     string `json:"Tag,omitempty"`
		Commit  string `json:"Commit,omitempty"`
	} `json:"Server"`
}

// CollectionStats is from /api/v3/Dashboard/Stats.
type CollectionStats struct {
	FileCount              int     `json:"FileCount"`
	FileSize               int64   `json:"FileSize"`
	SeriesCount            int     `json:"SeriesCount"`
	GroupCount             int     `json:"GroupCount"`
	FinishedSeries         int     `json:"FinishedSeries"`
	WatchedEpisodes        int     `json:"WatchedEpisodes"`
	WatchedHours           float64 `json:"WatchedHours"`
	PercentDuplicate       float64 `json:"PercentDuplicate"`
	MissingEpisodes        int     `json:"MissingEpisodes"`
	UnrecognizedFiles      int     `json:"UnrecognizedFiles"`
	SeriesWithMissingLinks int     `json:"SeriesWithMissingLinks"`
}

// SeriesSummary is from /api/v3/Dashboard/SeriesSummary.
type SeriesSummary struct {
	Series  int `json:"Series"`
	Special int `json:"Special"`
	Movie   int `json:"Movie"`
	OVA     int `json:"OVA"`
	Web     int `json:"Web"`
	Other   int `json:"Other"`
	Unknown int `json:"Unknown"`
	None    int `json:"None"`
}

func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("apikey", c.apiKey)
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

// TestConnection tests connectivity by hitting the version endpoint.
func (c *Client) TestConnection(ctx context.Context) (*VersionInfo, error) {
	var v VersionInfo
	if err := c.get(ctx, "/api/v3/Init/Version", &v); err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}
	return &v, nil
}

// GetStats returns collection statistics from the dashboard.
func (c *Client) GetStats(ctx context.Context) (*CollectionStats, error) {
	var stats CollectionStats
	if err := c.get(ctx, "/api/v3/Dashboard/Stats", &stats); err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &stats, nil
}

// GetSeriesSummary returns the series type breakdown.
func (c *Client) GetSeriesSummary(ctx context.Context) (*SeriesSummary, error) {
	var summary SeriesSummary
	if err := c.get(ctx, "/api/v3/Dashboard/SeriesSummary", &summary); err != nil {
		return nil, fmt.Errorf("get series summary: %w", err)
	}
	return &summary, nil
}
