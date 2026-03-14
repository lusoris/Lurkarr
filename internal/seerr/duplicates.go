package seerr

import (
	"context"
	"fmt"
	"log/slog"
)

// DuplicateFlag represents a Seerr request that was flagged as a potential duplicate.
type DuplicateFlag struct {
	RequestID   int    `json:"request_id"`
	MediaTitle  string `json:"media_title"`
	ExternalID  string `json:"external_id"`
	RequestType string `json:"request_type"` // "movie" or "tv"
	Is4K        bool   `json:"is4k"`
	RequestedBy string `json:"requested_by"`
	Reason      string `json:"reason"`
}

// ScanResult summarizes the duplicate scan outcome.
type DupScanResult struct {
	TotalScanned int             `json:"total_scanned"`
	Duplicates   []DuplicateFlag `json:"duplicates"`
}

// ScanForDuplicates fetches all approved/pending Seerr requests and checks
// each against cross-instance media data to find duplicates.
func (r *RequestRouter) ScanForDuplicates(ctx context.Context, client *Client) (*DupScanResult, error) {
	if r.DB == nil {
		return &DupScanResult{}, nil
	}

	result := &DupScanResult{}

	// Fetch all requests in pages of 50.
	for skip := 0; ; skip += 50 {
		resp, err := client.ListRequests(ctx, "all", 50, skip)
		if err != nil {
			return nil, fmt.Errorf("list requests (skip=%d): %w", skip, err)
		}

		for _, req := range resp.Results {
			result.TotalScanned++

			decision := r.Evaluate(ctx, req)
			if decision.Action == "decline" {
				result.Duplicates = append(result.Duplicates, DuplicateFlag{
					RequestID:   req.ID,
					MediaTitle:  mediaTitle(req),
					ExternalID:  buildExternalID(req),
					RequestType: req.Type,
					Is4K:        req.Is4K,
					RequestedBy: req.RequestedBy.DisplayName,
					Reason:      decision.Reason,
				})
			}
		}

		// Stop when we've fetched all pages.
		if skip+len(resp.Results) >= resp.PageInfo.Results {
			break
		}
	}

	slog.Info("seerr: duplicate scan complete",
		"scanned", result.TotalScanned,
		"duplicates", len(result.Duplicates))

	return result, nil
}

func mediaTitle(req MediaRequest) string {
	if req.Media.MediaType != "" {
		return fmt.Sprintf("%s (ID: %d)", req.Media.MediaType, req.Media.TmdbID)
	}
	return fmt.Sprintf("%s #%d", req.Type, req.ID)
}
