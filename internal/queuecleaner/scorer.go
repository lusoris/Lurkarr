package queuecleaner

import (
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ScoreQueueItem computes a weighted score for a queue record based on the scoring profile.
// Higher score = better quality download.
func ScoreQueueItem(item *arrclient.QueueRecord, profile *database.ScoringProfile) int {
	score := 0

	// Custom format score (from the arr itself)
	score += item.CustomFormatScore * profile.CustomFormatWeight

	// Quality tier score (higher quality ID = generally better)
	if profile.PreferHigherQuality && item.Quality != nil {
		score += item.Quality.Quality.ID * 100
	}

	// Size score (larger = more data = potentially better quality)
	if profile.PreferLargerSize && item.Size > 0 {
		// Normalise to GB, cap at 100
		sizeGB := int(item.Size / (1024 * 1024 * 1024))
		if sizeGB > 100 {
			sizeGB = 100
		}
		score += sizeGB * profile.SizeWeight
	}

	// Download progress as tiebreaker (more complete = prefer keeping)
	if item.Size > 0 {
		progress := int(((item.Size - item.Sizeleft) * 100) / item.Size)
		score += progress // 0-100 points
	}

	return score
}

// FindDuplicates groups queue records by media ID and identifies lower-scored duplicates.
// Strategy "highest": keep the best-scoring item, remove all others.
// Strategy "adequate": keep the first item scoring above the threshold, remove all others.
func FindDuplicates(records []arrclient.QueueRecord, profile *database.ScoringProfile) []DuplicateResult {
	type scored struct {
		record arrclient.QueueRecord
		score  int
	}

	// Group by media ID
	groups := make(map[int][]scored)
	for _, r := range records {
		mid := r.MediaID()
		if mid == 0 {
			continue
		}
		groups[mid] = append(groups[mid], scored{record: r, score: ScoreQueueItem(&r, profile)})
	}

	var results []DuplicateResult
	for mediaID, items := range groups {
		if len(items) < 2 {
			continue
		}

		keepIdx := 0
		if profile.Strategy == "adequate" {
			// Keep the first item that meets the threshold
			for i, item := range items {
				if item.score >= profile.AdequateThreshold {
					keepIdx = i
					break
				}
			}
		} else {
			// Default: keep highest score
			for i := 1; i < len(items); i++ {
				if items[i].score > items[keepIdx].score {
					keepIdx = i
				}
			}
		}

		// Everything except the kept item is a duplicate to remove
		for i, item := range items {
			if i == keepIdx {
				continue
			}
			results = append(results, DuplicateResult{
				MediaID:       mediaID,
				RemoveQueueID: item.record.ID,
				RemoveTitle:   item.record.Title,
				RemoveScore:   item.score,
				KeepQueueID:   items[keepIdx].record.ID,
				KeepTitle:     items[keepIdx].record.Title,
				KeepScore:     items[keepIdx].score,
			})
		}
	}
	return results
}

// DuplicateResult describes a duplicate that should be removed.
type DuplicateResult struct {
	MediaID       int
	RemoveQueueID int
	RemoveTitle   string
	RemoveScore   int
	KeepQueueID   int
	KeepTitle     string
	KeepScore     int
}
