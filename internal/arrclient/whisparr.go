package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Whisparr v2 uses the same API as Radarr (v3 API).
const whisparrAPI = "/api/v3"

// WhisparrMovie represents content from Whisparr v2.
type WhisparrMovie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Monitored bool   `json:"monitored"`
	HasFile   bool   `json:"hasFile"`
}

// WhisparrGetMissing fetches items without files.
func (c *Client) WhisparrGetMissing(ctx context.Context) ([]WhisparrMovie, error) {
	var movies []WhisparrMovie
	if err := c.get(ctx, whisparrAPI+"/movie", &movies); err != nil {
		return nil, fmt.Errorf("whisparr get movies: %w", err)
	}
	var missing []WhisparrMovie
	for _, m := range movies {
		if m.Monitored && !m.HasFile {
			missing = append(missing, m)
		}
	}
	return missing, nil
}

// WhisparrGetCutoffUnmet fetches items that haven't met quality cutoff.
func (c *Client) WhisparrGetCutoffUnmet(ctx context.Context) ([]WhisparrMovie, error) {
	var resp struct {
		TotalRecords int             `json:"totalRecords"`
		Records      []WhisparrMovie `json:"records"`
	}
	if err := c.get(ctx, whisparrAPI+"/wanted/cutoff?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("whisparr get cutoff unmet: %w", err)
	}
	return resp.Records, nil
}

// WhisparrSearchMovie triggers a search for items.
func (c *Client) WhisparrSearchMovie(ctx context.Context, movieIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     "MoviesSearch",
		"movieIds": movieIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, whisparrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("whisparr search: %w", err)
	}
	return &resp, nil
}

// WhisparrGetQueue returns the current download queue.
func (c *Client) WhisparrGetQueue(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, whisparrAPI+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("whisparr get queue: %w", err)
	}
	return &resp, nil
}

// WhisparrTestConnection tests the Whisparr v2 API connection.
func (c *Client) WhisparrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
