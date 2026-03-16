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
	Added     string `json:"added"`
}

// RadarrGetMissing fetches movies without files via the wanted/missing endpoint.
func (c *Client) RadarrGetMissing(ctx context.Context) ([]RadarrMovie, error) {
	return getWanted[RadarrMovie](ctx, c, radarrAPI, "missing", "title", "ascending", "radarr get missing")
}

// RadarrGetCutoffUnmet fetches movies that haven't met quality cutoff.
func (c *Client) RadarrGetCutoffUnmet(ctx context.Context) ([]RadarrMovie, error) {
	return getWanted[RadarrMovie](ctx, c, radarrAPI, "cutoff", "title", "ascending", "radarr get cutoff unmet")
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
	records, err := getAllPages[QueueRecord](ctx, c, radarrAPI+"/queue")
	if err != nil {
		return nil, fmt.Errorf("radarr get queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// RadarrGetQueueEnriched returns the queue with embedded movie data.
// Used for cross-arr sync to match by TMDB ID instead of title.
func (c *Client) RadarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, radarrAPI+"/queue?includeMovie=true")
	if err != nil {
		return nil, fmt.Errorf("radarr get enriched queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// RadarrTestConnection tests the Radarr API connection.
func (c *Client) RadarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
