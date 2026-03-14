package queuecleaner

import (
	"regexp"
	"strings"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ScoreQueueItem computes a weighted score for a queue record based on the scoring profile.
// Higher score = better quality download. Scoring priority:
//  1. Custom format score (authoritative, from the *arr quality profile)
//  2. Resolution rank (2160p > 1080p > 720p > 480p)
//  3. Source rank (Remux > BluRay > WEB-DL > WEBRip > HDTV > DVD)
//  4. Revision bonus (PROPER/REPACK)
//  5. Size tiebreaker
//  6. Progress tiebreaker (more complete = prefer keeping)
func ScoreQueueItem(item *arrclient.QueueRecord, profile *database.ScoringProfile) int {
	score := 0

	// 1. Custom format score (from the arr itself)
	score += item.CustomFormatScore * profile.CustomFormatWeight

	// 2. Resolution rank — extract from quality name first, fall back to title
	resolution := extractResolution(item)
	score += resolutionRank(resolution) * profile.ResolutionWeight

	// 3. Source rank — parsed from release title
	info := ParseRelease(item.Title)
	score += sourceRank(info.Source) * profile.SourceWeight

	// 4. Revision bonus (PROPER/REPACK get a flat bonus)
	if item.Quality != nil && item.Quality.Revision.Version > 1 {
		score += profile.RevisionBonus
	}

	// 5. Size tiebreaker
	if profile.PreferLargerSize && item.Size > 0 {
		sizeGB := int(item.Size / (1024 * 1024 * 1024))
		if sizeGB > 100 {
			sizeGB = 100
		}
		score += sizeGB * profile.SizeWeight
	}

	// 6. Download progress as tiebreaker (more complete = prefer keeping)
	if item.Size > 0 {
		progress := int(((item.Size - item.Sizeleft) * 100) / item.Size)
		score += progress // 0-100 points
	}

	return score
}

// resolutionRank maps resolution strings to numeric rank values.
func resolutionRank(res string) int {
	switch strings.ToLower(res) {
	case "4320p":
		return 5
	case "2160p":
		return 4
	case "1080p":
		return 3
	case "720p":
		return 2
	case "576p", "480p":
		return 1
	default:
		return 0
	}
}

// sourceRank maps source type strings to numeric rank values.
func sourceRank(src string) int {
	switch src {
	case "Remux":
		return 5
	case "BluRay":
		return 4
	case "WEB-DL":
		return 3
	case "WEBRip", "WEB":
		return 2
	case "HDTV":
		return 1
	case "DVD", "DVDRip":
		return 0
	default:
		return -1 // CAM, TS, unknown
	}
}

var reResolutionFromName = regexp.MustCompile(`(?i)(4320p|2160p|1080p|720p|576p|480p)`)

// extractResolution gets the resolution from the quality name first (e.g. "Bluray-1080p")
// and falls back to parsing the release title.
func extractResolution(item *arrclient.QueueRecord) string {
	if item.Quality != nil {
		if m := reResolutionFromName.FindString(item.Quality.Quality.Name); m != "" {
			return strings.ToLower(m)
		}
	}
	return ParseRelease(item.Title).Resolution
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
