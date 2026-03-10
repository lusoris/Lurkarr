package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Eros (Whisparr v3) uses the v3 API but with different endpoint behavior.
const erosAPI = "/api/v3"

// ErosMovie represents content from Whisparr v3 (Eros).
type ErosMovie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Monitored bool   `json:"monitored"`
	HasFile   bool   `json:"hasFile"`
}

// ErosGetMissing fetches items without files.
func (c *Client) ErosGetMissing(ctx context.Context) ([]ErosMovie, error) {
	var movies []ErosMovie
	if err := c.get(ctx, erosAPI+"/movie", &movies); err != nil {
		return nil, fmt.Errorf("eros get movies: %w", err)
	}
	var missing []ErosMovie
	for _, m := range movies {
		if m.Monitored && !m.HasFile {
			missing = append(missing, m)
		}
	}
	return missing, nil
}

// ErosGetCutoffUnmet fetches items that haven't met quality cutoff.
func (c *Client) ErosGetCutoffUnmet(ctx context.Context) ([]ErosMovie, error) {
	var resp struct {
		TotalRecords int         `json:"totalRecords"`
		Records      []ErosMovie `json:"records"`
	}
	if err := c.get(ctx, erosAPI+"/wanted/cutoff?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("eros get cutoff unmet: %w", err)
	}
	return resp.Records, nil
}

// ErosSearchMovie triggers a search for items.
func (c *Client) ErosSearchMovie(ctx context.Context, movieIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     "MoviesSearch",
		"movieIds": movieIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, erosAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("eros search: %w", err)
	}
	return &resp, nil
}

// ErosGetQueue returns the current download queue.
func (c *Client) ErosGetQueue(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, erosAPI+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("eros get queue: %w", err)
	}
	return &resp, nil
}

// ErosTestConnection tests the Eros (Whisparr v3) API connection.
func (c *Client) ErosTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
