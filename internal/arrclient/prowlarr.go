package arrclient

import (
	"context"
	"fmt"
)

const prowlarrAPI = "/api/v1"

// ProwlarrIndexer represents an indexer from Prowlarr.
type ProwlarrIndexer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Enable   bool   `json:"enable"`
	Priority int    `json:"priority"`
}

// ProwlarrIndexerStats represents indexer statistics.
type ProwlarrIndexerStats struct {
	IndexerID           int `json:"indexerId"`
	NumberOfQueries     int `json:"numberOfQueries"`
	NumberOfGrabs       int `json:"numberOfGrabs"`
	NumberOfFailures    int `json:"numberOfFailures"`
	AverageResponseTime int `json:"averageResponseTime"`
}

// ProwlarrGetIndexers fetches all configured indexers.
func (c *Client) ProwlarrGetIndexers(ctx context.Context) ([]ProwlarrIndexer, error) {
	var indexers []ProwlarrIndexer
	if err := c.get(ctx, prowlarrAPI+"/indexer", &indexers); err != nil {
		return nil, fmt.Errorf("prowlarr get indexers: %w", err)
	}
	return indexers, nil
}

// ProwlarrGetIndexerStats fetches indexer statistics.
func (c *Client) ProwlarrGetIndexerStats(ctx context.Context) ([]ProwlarrIndexerStats, error) {
	var resp struct {
		Indexers []ProwlarrIndexerStats `json:"indexers"`
	}
	if err := c.get(ctx, prowlarrAPI+"/indexerstats", &resp); err != nil {
		return nil, fmt.Errorf("prowlarr get indexer stats: %w", err)
	}
	return resp.Indexers, nil
}

// ProwlarrTestConnection tests the Prowlarr API connection.
func (c *Client) ProwlarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v1")
}
