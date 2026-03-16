package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

const lidarrAPI = "/api/v1"

// LidarrAlbum represents an album from Lidarr.
type LidarrAlbum struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Monitored   bool   `json:"monitored"`
	ReleaseDate string `json:"releaseDate"`
	Statistics  struct {
		TrackCount     int `json:"trackCount"`
		TrackFileCount int `json:"trackFileCount"`
	} `json:"statistics"`
}

// LidarrGetMissing fetches albums without all tracks.
func (c *Client) LidarrGetMissing(ctx context.Context) ([]LidarrAlbum, error) {
	return getWanted[LidarrAlbum](ctx, c, lidarrAPI, "missing", "title", "ascending", "lidarr get missing")
}

// LidarrGetCutoffUnmet fetches albums that haven't met quality cutoff.
func (c *Client) LidarrGetCutoffUnmet(ctx context.Context) ([]LidarrAlbum, error) {
	return getWanted[LidarrAlbum](ctx, c, lidarrAPI, "cutoff", "title", "ascending", "lidarr get cutoff unmet")
}

// LidarrSearchAlbum triggers a search for albums.
func (c *Client) LidarrSearchAlbum(ctx context.Context, albumIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     "AlbumSearch",
		"albumIds": albumIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, lidarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("lidarr search album: %w", err)
	}
	return &resp, nil
}

// LidarrGetQueue returns the current download queue.
func (c *Client) LidarrGetQueue(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, lidarrAPI+"/queue")
	if err != nil {
		return nil, fmt.Errorf("lidarr get queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// LidarrGetQueueEnriched returns the queue with embedded album data.
func (c *Client) LidarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, lidarrAPI+"/queue?includeAlbum=true")
	if err != nil {
		return nil, fmt.Errorf("lidarr get enriched queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// LidarrTestConnection tests the Lidarr API connection.
func (c *Client) LidarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v1")
}
