package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

const readarrAPI = "/api/v1"

// ReadarrBook represents a book from Readarr.
type ReadarrBook struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Monitored   bool   `json:"monitored"`
	ReleaseDate string `json:"releaseDate"`
	Statistics  struct {
		BookFileCount int `json:"bookFileCount"`
	} `json:"statistics"`
}

// ReadarrGetMissing fetches books without files.
func (c *Client) ReadarrGetMissing(ctx context.Context) ([]ReadarrBook, error) {
	var resp struct {
		TotalRecords int           `json:"totalRecords"`
		Records      []ReadarrBook `json:"records"`
	}
	if err := c.get(ctx, readarrAPI+"/wanted/missing?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("readarr get missing: %w", err)
	}
	return resp.Records, nil
}

// ReadarrGetCutoffUnmet fetches books that haven't met quality cutoff.
func (c *Client) ReadarrGetCutoffUnmet(ctx context.Context) ([]ReadarrBook, error) {
	var resp struct {
		TotalRecords int           `json:"totalRecords"`
		Records      []ReadarrBook `json:"records"`
	}
	if err := c.get(ctx, readarrAPI+"/wanted/cutoff?sortKey=title&sortDirection=ascending&pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("readarr get cutoff unmet: %w", err)
	}
	return resp.Records, nil
}

// ReadarrSearchBook triggers a search for books.
func (c *Client) ReadarrSearchBook(ctx context.Context, bookIDs []int) (*CommandResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"name":    "BookSearch",
		"bookIds": bookIDs,
	})
	var resp CommandResponse
	if err := c.post(ctx, readarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("readarr search book: %w", err)
	}
	return &resp, nil
}

// ReadarrGetQueue returns the current download queue.
func (c *Client) ReadarrGetQueue(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, readarrAPI+"/queue?pageSize=1000", &resp); err != nil {
		return nil, fmt.Errorf("readarr get queue: %w", err)
	}
	return &resp, nil
}

// ReadarrGetQueueEnriched returns the queue with embedded book data.
func (c *Client) ReadarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	var resp QueueResponse
	if err := c.get(ctx, readarrAPI+"/queue?pageSize=1000&includeBook=true", &resp); err != nil {
		return nil, fmt.Errorf("readarr get enriched queue: %w", err)
	}
	return &resp, nil
}

// ReadarrTestConnection tests the Readarr API connection.
func (c *Client) ReadarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v1")
}
