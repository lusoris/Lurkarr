// Package seerr provides a client for the Seerr API.
package seerr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MediaRequestStatus represents the status of a media request.
type MediaRequestStatus int

const (
	RequestPending   MediaRequestStatus = 1
	RequestApproved  MediaRequestStatus = 2
	RequestDeclined  MediaRequestStatus = 3
	RequestFailed    MediaRequestStatus = 4
	RequestCompleted MediaRequestStatus = 5
)

// MediaStatus represents the overall availability of a media item.
type MediaStatus int

const (
	MediaUnknown            MediaStatus = 1
	MediaPending            MediaStatus = 2
	MediaProcessing         MediaStatus = 3
	MediaPartiallyAvailable MediaStatus = 4
	MediaAvailable          MediaStatus = 5
	MediaDeleted            MediaStatus = 6
)

// PageInfo contains pagination metadata.
type PageInfo struct {
	Pages    int `json:"pages"`
	PageSize int `json:"pageSize"`
	Results  int `json:"results"`
	Page     int `json:"page"`
}

// RequestsResponse is the paginated response from GET /request.
type RequestsResponse struct {
	PageInfo PageInfo       `json:"pageInfo"`
	Results  []MediaRequest `json:"results"`
}

// MediaRequest represents a single media request.
type MediaRequest struct {
	ID            int                `json:"id"`
	Status        MediaRequestStatus `json:"status"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	Type          string             `json:"type"` // "movie" or "tv"
	Is4K          bool               `json:"is4k"`
	ServerID      *int               `json:"serverId"`
	ProfileID     *int               `json:"profileId"`
	RootFolder    *string            `json:"rootFolder"`
	IsAutoRequest bool               `json:"isAutoRequest"`
	SeasonCount   int                `json:"seasonCount"`
	Seasons       []SeasonRequest    `json:"seasons"`
	Media         Media              `json:"media"`
	RequestedBy   User               `json:"requestedBy"`
	ModifiedBy    *User              `json:"modifiedBy"`
}

// SeasonRequest represents a season within a request.
type SeasonRequest struct {
	ID           int                `json:"id"`
	SeasonNumber int                `json:"seasonNumber"`
	Status       MediaRequestStatus `json:"status"`
}

// Media represents a media item's status and metadata.
type Media struct {
	ID                    int         `json:"id"`
	MediaType             string      `json:"mediaType"`
	TmdbID                int         `json:"tmdbId"`
	TvdbID                *int        `json:"tvdbId"`
	ImdbID                *string     `json:"imdbId"`
	Status                MediaStatus `json:"status"`
	Status4K              MediaStatus `json:"status4k"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
	ExternalServiceID     *int        `json:"externalServiceId"`
	ExternalServiceSlug   *string     `json:"externalServiceSlug"`
	ExternalServiceID4K   *int        `json:"externalServiceId4k"`
	ExternalServiceSlug4K *string     `json:"externalServiceSlug4k"`
	ServiceURL            *string     `json:"serviceUrl"`
	Seasons               []Season    `json:"seasons"`
}

// Season represents a season's availability status.
type Season struct {
	ID           int         `json:"id"`
	SeasonNumber int         `json:"seasonNumber"`
	Status       MediaStatus `json:"status"`
	Status4K     MediaStatus `json:"status4k"`
}

// User represents a Seerr user.
type User struct {
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
}

// RequestCount holds counts by type and status.
type RequestCount struct {
	Total      int `json:"total"`
	Movie      int `json:"movie"`
	TV         int `json:"tv"`
	Pending    int `json:"pending"`
	Approved   int `json:"approved"`
	Declined   int `json:"declined"`
	Processing int `json:"processing"`
	Available  int `json:"available"`
}

// AboutInfo holds server info from GET /settings/about.
type AboutInfo struct {
	Version         string `json:"version"`
	TotalMediaItems int    `json:"totalMediaItems"`
	TotalRequests   int    `json:"totalRequests"`
}

// Client is an HTTP client for the Seerr API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Seerr API client.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetAbout retrieves server info.
func (c *Client) GetAbout(ctx context.Context) (*AboutInfo, error) {
	var info AboutInfo
	if err := c.get(ctx, "/api/v1/settings/about", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetRequestCount retrieves request counts by type and status.
func (c *Client) GetRequestCount(ctx context.Context) (*RequestCount, error) {
	var count RequestCount
	if err := c.get(ctx, "/api/v1/request/count", nil, &count); err != nil {
		return nil, err
	}
	return &count, nil
}

// ListRequests retrieves a paginated list of media requests.
func (c *Client) ListRequests(ctx context.Context, filter string, take, skip int) (*RequestsResponse, error) {
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if take > 0 {
		params.Set("take", strconv.Itoa(take))
	}
	if skip > 0 {
		params.Set("skip", strconv.Itoa(skip))
	}
	params.Set("sort", "added")

	var resp RequestsResponse
	if err := c.get(ctx, "/api/v1/request", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRequest retrieves a single media request by ID.
func (c *Client) GetRequest(ctx context.Context, id int) (*MediaRequest, error) {
	var req MediaRequest
	if err := c.get(ctx, fmt.Sprintf("/api/v1/request/%d", id), nil, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ApproveRequest approves a pending request.
func (c *Client) ApproveRequest(ctx context.Context, id int) error {
	return c.post(ctx, fmt.Sprintf("/api/v1/request/%d/approve", id))
}

// DeclineRequest declines a pending request.
func (c *Client) DeclineRequest(ctx context.Context, id int) error {
	return c.post(ctx, fmt.Sprintf("/api/v1/request/%d/decline", id))
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) post(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
