package seerr

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// RoutingDecision represents the outcome of evaluating a Seerr request.
type RoutingDecision struct {
	Action  string // "approve", "decline", "skip"
	Reason  string
	GroupID *uuid.UUID
}

// RoutingStore provides the database queries needed for request routing.
type RoutingStore interface {
	FindMediaPresenceByExternalID(ctx context.Context, externalID string) ([]database.MediaPresenceResult, error)
	CreateCrossInstanceAction(ctx context.Context, action database.CrossInstanceAction) error
}

// RequestRouter evaluates Seerr requests against cross-instance data to decide
// whether to approve, decline, or skip them.
type RequestRouter struct {
	DB RoutingStore
}

// Evaluate checks whether a pending Seerr request should be approved or declined
// based on cross-instance media presence and quality hierarchy rules.
func (r *RequestRouter) Evaluate(ctx context.Context, req MediaRequest) RoutingDecision {
	if r.DB == nil {
		return RoutingDecision{Action: "approve", Reason: "no routing store configured"}
	}

	// Build canonical external ID from Seerr media metadata.
	externalID := BuildExternalID(req)
	if externalID == "" {
		return RoutingDecision{Action: "approve", Reason: "no external ID available"}
	}

	presences, err := r.DB.FindMediaPresenceByExternalID(ctx, externalID)
	if err != nil {
		slog.Error("seerr router: failed to look up media presence", "external_id", externalID, "error", err)
		return RoutingDecision{Action: "approve", Reason: "lookup error, defaulting to approve"}
	}

	if len(presences) == 0 {
		return RoutingDecision{Action: "approve", Reason: "no cross-instance data"}
	}

	// Check each group for quality hierarchy conflicts.
	for _, p := range presences {
		if p.GroupMode != "quality_hierarchy" {
			continue
		}

		// In quality hierarchy mode, if a higher-ranked (lower number = better)
		// instance already has the file, decline the request.
		bestRank := 0
		var bestInst *database.PresenceInstance
		for i := range p.Instances {
			inst := &p.Instances[i]
			if inst.HasFile && inst.QualityRank > 0 {
				if bestRank == 0 || inst.QualityRank < bestRank {
					bestRank = inst.QualityRank
					bestInst = inst
				}
			}
		}

		if bestInst != nil {
			// A request is 4K if is4k flag is set; we check if the best file
			// is from a higher-tier instance (rank 1 = best quality).
			// If the requestor is asking for lower quality than what exists, decline.
			// For now: if any instance with rank 1 has the file, and the Seerr
			// request is NOT 4K, it's a duplicate of lower quality → decline.
			if !req.Is4K && bestRank == 1 {
				gid := p.GroupID
				return RoutingDecision{
					Action:  "decline",
					Reason:  fmt.Sprintf("already exists in higher quality instance %q (rank %d)", bestInst.Name, bestRank),
					GroupID: &gid,
				}
			}

			// If media already exists in 2+ instances, flag as duplicate.
			instancesWithFile := 0
			for _, inst := range p.Instances {
				if inst.HasFile {
					instancesWithFile++
				}
			}
			if instancesWithFile >= 2 {
				gid := p.GroupID
				return RoutingDecision{
					Action:  "decline",
					Reason:  fmt.Sprintf("media already in %d instances within group", instancesWithFile),
					GroupID: &gid,
				}
			}
		}
	}

	return RoutingDecision{Action: "approve", Reason: "no quality conflict detected"}
}

// LogAction records a routing decision in the audit log.
func (r *RequestRouter) LogAction(ctx context.Context, req MediaRequest, decision RoutingDecision) {
	if r.DB == nil || decision.GroupID == nil {
		return
	}

	externalID := BuildExternalID(req)
	title := req.Media.MediaType
	if req.Type != "" {
		title = req.Type
	}

	reqID := req.ID
	action := database.CrossInstanceAction{
		GroupID:        *decision.GroupID,
		ExternalID:     externalID,
		Title:          fmt.Sprintf("[%s] Seerr request #%d", title, req.ID),
		Action:         decision.Action,
		Reason:         decision.Reason,
		SeerrRequestID: &reqID,
	}

	if err := r.DB.CreateCrossInstanceAction(ctx, action); err != nil {
		slog.Error("seerr router: failed to log action", "error", err)
	}
}

// BuildExternalID converts a Seerr request's media IDs into the canonical format
// used by the cross-instance scanner (tmdb:X or tvdb:X).
func BuildExternalID(req MediaRequest) string {
	switch req.Type {
	case "movie":
		if req.Media.TmdbID > 0 {
			return fmt.Sprintf("tmdb:%d", req.Media.TmdbID)
		}
	case "tv":
		if req.Media.TvdbID != nil && *req.Media.TvdbID > 0 {
			return fmt.Sprintf("tvdb:%d", *req.Media.TvdbID)
		}
		if req.Media.TmdbID > 0 {
			return fmt.Sprintf("tmdb:%d", req.Media.TmdbID)
		}
	}
	return ""
}
