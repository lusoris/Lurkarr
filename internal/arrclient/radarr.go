package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

const radarrAPI = "/api/v3"

// RadarrMovie represents a movie from Radarr.
type RadarrMovie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Monitored bool   `json:"monitored"`
	HasFile   bool   `json:"hasFile"`
}

// RadarrGetMissing fetches movies without files.
func (c *Client) RadarrGetMissing(ctx context.Context) ([]RadarrMovie, error) {
	var movies []RadarrMovie
	if err := c.get(ctx, radarrAPI+"/movie", &movies); err != nil {
		return nil, fmt.Errorf("radarr get movies: %w", err)
	}
	var missing []RadarrMovie
	for _, m := range movies {
		if m.Monitored && !m.HasFile {
			missing = append(missing, m)
		}
	}
	return missing, nil
}

// RadarrGetCutoffUnmet fetches movies that haven't met quality cutoff.
func (c *Client) RadarrGetCutoffUnmet(ctx context.Context) ([]RadarrMovie, error) {
	var resp struct {
		TotalRecords int           `json:"totalRecords"`
		Records      []RadarrMovie `json:"records"`
	}
	if err := c.get(ctx, radarrAPI+"/wanted/cutoff?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("radarr get cutoff unmet: %w", err)
	}
	return resp.Records, nil
}

// RadarrSearchMovie triggers a search for a movie.
func (c *Client) RadarrSearchMovie(ctx context.Context, movieIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     "MoviesSearch",
		"movieIds": movieIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, radarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("radarr search movie: %w", err)
	}
	return &resp, nil
}

// RadarrGetQueue returns the current download queue.
func (c *Client) RadarrGetQueue(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, radarrAPI+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("radarr get queue: %w", err)
	}
	return &resp, nil
}

// RadarrTestConnection tests the Radarr API connection.
func (c *Client) RadarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
