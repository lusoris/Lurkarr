package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Eros (Whisparr v3) is Radarr-based: scenes and movies are individual items.
const erosAPI = "/api/v3"

// ErosMovie represents a scene or movie from Whisparr v3 (Eros).
type ErosMovie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Monitored bool   `json:"monitored"`
	HasFile   bool   `json:"hasFile"`
	ItemType  string `json:"itemType"` // "movie" or "scene"
}

// ErosGetMissing fetches scenes/movies without files.
// Eros has no wanted/missing endpoint, so we fetch all and filter client-side.
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

// ErosGetCutoffUnmet returns nil — Eros has no wanted/cutoff endpoint.
func (c *Client) ErosGetCutoffUnmet(_ context.Context) ([]ErosMovie, error) {
	return nil, nil
}

// ErosSearchMovie triggers a search for scenes/movies by ID.
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

// ErosGetQueueEnriched returns the queue with embedded movie data.
func (c *Client) ErosGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, erosAPI+"/queue?pageSize=1000&includeMovie=true", &resp); err != nil {
		return nil, fmt.Errorf("eros get enriched queue: %w", err)
	}
	return &resp, nil
}

// ErosTestConnection tests the Eros (Whisparr v3) API connection.
func (c *Client) ErosTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
