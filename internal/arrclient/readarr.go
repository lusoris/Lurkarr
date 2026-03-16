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
	return getWanted[ReadarrBook](ctx, c, readarrAPI, "missing", "title", "ascending", "readarr get missing")
}

// ReadarrGetCutoffUnmet fetches books that haven't met quality cutoff.
func (c *Client) ReadarrGetCutoffUnmet(ctx context.Context) ([]ReadarrBook, error) {
	return getWanted[ReadarrBook](ctx, c, readarrAPI, "cutoff", "title", "ascending", "readarr get cutoff unmet")
}

// ReadarrSearchBook triggers a search for books.
func (c *Client) ReadarrSearchBook(ctx context.Context, bookIDs []int) (*CommandResponse, error) {
	body, err := json.Marshal(map[string]any{
		"name":    "BookSearch",
		"bookIds": bookIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal readarr search: %w", err)
	}
	var resp CommandResponse
	if err := c.post(ctx, readarrAPI+"/command", bytes.NewReader(body), &resp); err != nil {
		return nil, fmt.Errorf("readarr search book: %w", err)
	}
	return &resp, nil
}

// ReadarrGetQueue returns the current download queue.
func (c *Client) ReadarrGetQueue(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, readarrAPI+"/queue")
	if err != nil {
		return nil, fmt.Errorf("readarr get queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// ReadarrGetQueueEnriched returns the queue with embedded book data.
func (c *Client) ReadarrGetQueueEnriched(ctx context.Context) (*QueueResponse, error) {
	records, err := getAllPages[QueueRecord](ctx, c, readarrAPI+"/queue?includeBook=true")
	if err != nil {
		return nil, fmt.Errorf("readarr get enriched queue: %w", err)
	}
	return &QueueResponse{TotalRecords: len(records), Records: records}, nil
}

// ReadarrTestConnection tests the Readarr API connection.
func (c *Client) ReadarrTestConnection(ctx context.Context) (*SystemStatus, error) {
	return c.TestConnection(ctx, "v1")
}
