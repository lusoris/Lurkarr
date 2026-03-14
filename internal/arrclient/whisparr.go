package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Whisparr v2 is Sonarr-based: studios are "series", scenes/movies are "episodes".
const whisparrAPI = "/api/v3"

// WhisparrEpisode represents a scene or movie from Whisparr v2.
type WhisparrEpisode struct {
	ID           int    `json:"id"`
	SeriesID     int    `json:"seriesId"`
	Title        string `json:"title"`
	SeasonNumber int    `json:"seasonNumber"`
	Monitored    bool   `json:"monitored"`
	HasFile      bool   `json:"hasFile"`
	ReleaseDate  string `json:"releaseDate"`
}

// WhisparrGetMissing fetches scenes/movies without files via the wanted/missing endpoint.
func (c *Client) WhisparrGetMissing(ctx context.Context) ([]WhisparrEpisode, error) {
	var resp struct {
		TotalRecords int               `json:"totalRecords"`
		Records      []WhisparrEpisode `json:"records"`
	}
	if err := c.get(ctx, whisparrAPI+"/wanted/missing?sortKey=releaseDate&sortDirection=descending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("whisparr get missing: %w", err)
	}
	return resp.Records, nil
}

// WhisparrGetCutoffUnmet fetches scenes/movies that haven't met quality cutoff.
func (c *Client) WhisparrGetCutoffUnmet(ctx context.Context) ([]WhisparrEpisode, error) {
	var resp struct {
		TotalRecords int               `json:"totalRecords"`
		Records      []WhisparrEpisode `json:"records"`
	}
	if err := c.get(ctx, whisparrAPI+"/wanted/cutoff?sortKey=releaseDate&sortDirection=descending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("whisparr get cutoff unmet: %w", err)
	}
	return resp.Records, nil
}

// WhisparrSearchEpisode triggers a search for specific episode (scene/movie) IDs.
func (c *Client) WhisparrSearchEpisode(ctx context.Context, episodeIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":       "EpisodeSearch",
		"episodeIds": episodeIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, whisparrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("whisparr search episode: %w", err)
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

// WhisparrGetQueueEnriched returns the queue with embedded series+episode data.
func (c *Client) WhisparrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, whisparrAPI+"/queue?pageSize=1000&includeSeries=true&includeEpisode=true", &resp); err != nil {
		return nil, fmt.Errorf("whisparr get enriched queue: %w", err)
	}
	return &resp, nil
}

// WhisparrTestConnection tests the Whisparr v2 API connection.
func (c *Client) WhisparrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
