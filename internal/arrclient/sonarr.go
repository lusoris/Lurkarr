package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

const sonarrAPI = "/api/v3"

// SonarrSeries represents a series from Sonarr.
type SonarrSeries struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Monitored  bool   `json:"monitored"`
	Statistics struct {
		EpisodeCount     int `json:"episodeCount"`
		EpisodeFileCount int `json:"episodeFileCount"`
	} `json:"statistics"`
}

// SonarrEpisode represents an episode from Sonarr.
type SonarrEpisode struct {
	ID            int    `json:"id"`
	SeriesID      int    `json:"seriesId"`
	Title         string `json:"title"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Monitored     bool   `json:"monitored"`
	HasFile       bool   `json:"hasFile"`
	AirDateUtc    string `json:"airDateUtc"`
}

// SonarrGetMissing fetches episodes without files.
func (c *Client) SonarrGetMissing(ctx context.Context) ([]SonarrEpisode, error) {
	return getWanted[SonarrEpisode](ctx, c, sonarrAPI, "missing", "airDateUtc", "descending", "sonarr get missing")
}

// SonarrGetCutoffUnmet fetches episodes that haven't met quality cutoff.
func (c *Client) SonarrGetCutoffUnmet(ctx context.Context) ([]SonarrEpisode, error) {
	return getWanted[SonarrEpisode](ctx, c, sonarrAPI, "cutoff", "airDateUtc", "descending", "sonarr get cutoff unmet")
}

// SonarrSearchEpisode triggers a search for specific episode IDs.
func (c *Client) SonarrSearchEpisode(ctx context.Context, episodeIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":       "EpisodeSearch",
		"episodeIds": episodeIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, sonarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("sonarr search episode: %w", err)
	}
	return &resp, nil
}

// SonarrSearchSeason triggers a search for all episodes in a season.
func (c *Client) SonarrSearchSeason(ctx context.Context, seriesID, seasonNumber int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":         "SeasonSearch",
		"seriesId":     seriesID,
		"seasonNumber": seasonNumber,
	})
	var resp CommandResponse
	if err := c.post(ctx, sonarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("sonarr search season: %w", err)
	}
	return &resp, nil
}

// SonarrSearchSeries triggers a search for all episodes in a series.
func (c *Client) SonarrSearchSeries(ctx context.Context, seriesID int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     "SeriesSearch",
		"seriesId": seriesID,
	})
	var resp CommandResponse
	if err := c.post(ctx, sonarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("sonarr search series: %w", err)
	}
	return &resp, nil
}

// SonarrGetQueue returns the current download queue.
func (c *Client) SonarrGetQueue(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, sonarrAPI+"/queue")
	if err != nil {
		return nil, fmt.Errorf("sonarr get queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// SonarrGetQueueEnriched returns the queue with embedded series+episode data.
// Used for cross-arr sync to match by TVDB ID + season instead of title.
func (c *Client) SonarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, sonarrAPI+"/queue?includeSeries=true&includeEpisode=true")
	if err != nil {
		return nil, fmt.Errorf("sonarr get enriched queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// SonarrTestConnection tests the Sonarr API connection.
func (c *Client) SonarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v3")
}
