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
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Monitored  bool   `json:"monitored"`
	Statistics struct {
		TrackCount     int `json:"trackCount"`
		TrackFileCount int `json:"trackFileCount"`
	} `json:"statistics"`
}

// LidarrGetMissing fetches albums without all tracks.
func (c *Client) LidarrGetMissing(ctx context.Context) ([]LidarrAlbum, error) {
	var resp struct {
		TotalRecords int           `json:"totalRecords"`
		Records      []LidarrAlbum `json:"records"`
	}
	if err := c.get(ctx, lidarrAPI+"/wanted/missing?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("lidarr get missing: %w", err)
	}
	return resp.Records, nil
}

// LidarrGetCutoffUnmet fetches albums that haven't met quality cutoff.
func (c *Client) LidarrGetCutoffUnmet(ctx context.Context) ([]LidarrAlbum, error) {
	var resp struct {
		TotalRecords int           `json:"totalRecords"`
		Records      []LidarrAlbum `json:"records"`
	}
	if err := c.get(ctx, lidarrAPI+"/wanted/cutoff?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("lidarr get cutoff unmet: %w", err)
	}
	return resp.Records, nil
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
	var resp QueueResponse
	if err := c.get(ctx, lidarrAPI+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("lidarr get queue: %w", err)
	}
	return &resp, nil
}

// LidarrGetQueueEnriched returns the queue with embedded album data.
func (c *Client) LidarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, lidarrAPI+"/queue?pageSize=1000&includeAlbum=true", &resp); err != nil {
		return nil, fmt.Errorf("lidarr get enriched queue: %w", err)
	}
	return &resp, nil
}

// LidarrTestConnection tests the Lidarr API connection.
func (c *Client) LidarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v1")
}
