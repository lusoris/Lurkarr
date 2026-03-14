package queuecleaner

import (
	"strconv"
	"strings"

	"github.com/lusoris/lurkarr/internal/arrclient"
)

// Pack groups queue records that share the same DownloadID. A season pack in
// Sonarr creates one QueueRecord per episode, all referencing the same
// underlying torrent/NZB. The cleaner must treat these as a single unit to
// avoid N× duplicate API calls, inflated strike counts, and premature removals.
type Pack struct {
	DownloadID string
	Records    []arrclient.QueueRecord
}

// IsPack returns true when this group contains more than one queue record.
func (p *Pack) IsPack() bool {
	return len(p.Records) > 1
}

// Representative returns the first record, used for logging and fields that are
// identical across all pack members (Title, Size, Protocol, DownloadClient, etc.).
func (p *Pack) Representative() arrclient.QueueRecord {
	return p.Records[0]
}

// QueueIDs returns all queue record IDs in the pack.
func (p *Pack) QueueIDs() []int {
	ids := make([]int, len(p.Records))
	for i, r := range p.Records {
		ids[i] = r.ID
	}
	return ids
}

// AllImported returns true if every record in the pack has been imported.
func (p *Pack) AllImported() bool {
	for _, r := range p.Records {
		if r.TrackedDownloadState != "imported" {
			return false
		}
	}
	return true
}

// AnyImportPendingOrImported returns true if any record is imported or pending
// import (used for seeding eligibility).
func (p *Pack) AnyImportPendingOrImported() bool {
	for _, r := range p.Records {
		if r.TrackedDownloadState == "imported" || r.TrackedDownloadState == "importPending" {
			return true
		}
	}
	return false
}

// AnyHasImportFailure returns true if any record in the pack has an import
// failure matching the given patterns.
func (p *Pack) AnyHasImportFailure(patterns []string) bool {
	for _, r := range p.Records {
		if hasImportFailure(r, patterns) {
			return true
		}
	}
	return false
}

// ImportFailureRecord returns the first record with an import failure, or the
// representative if none match. Used for logging the failure reason.
func (p *Pack) ImportFailureRecord(patterns []string) arrclient.QueueRecord {
	for _, r := range p.Records {
		if hasImportFailure(r, patterns) {
			return r
		}
	}
	return p.Records[0]
}

// AllFilesDeleted returns true if every imported record's media file has been
// deleted externally. Records without enriched data are skipped (assumed present).
func (p *Pack) AllFilesDeleted() bool {
	checked := 0
	for _, r := range p.Records {
		if r.TrackedDownloadState != "imported" {
			continue
		}
		hasFile, ok := r.MediaHasFile()
		if !ok {
			continue
		}
		checked++
		if hasFile {
			return false
		}
	}
	return checked > 0
}

// AllUnmonitored returns true if every non-imported record's media is
// unmonitored. Records without enriched data are skipped (assumed monitored).
func (p *Pack) AllUnmonitored() bool {
	checked := 0
	for _, r := range p.Records {
		if r.TrackedDownloadState == "imported" {
			continue
		}
		monitored, ok := r.MediaMonitored()
		if !ok {
			continue
		}
		checked++
		if monitored {
			return false
		}
	}
	return checked > 0
}

// AnyMismatched returns true if any record in the pack shows a metadata mismatch.
func (p *Pack) AnyMismatched() bool {
	for _, r := range p.Records {
		if r.TrackedDownloadStatus != "warning" {
			continue
		}
		if r.TrackedDownloadState != "importPending" && r.TrackedDownloadState != "importFailed" {
			continue
		}
		if isMismatchedRelease(r) {
			return true
		}
	}
	return false
}

// GroupByDownloadID groups queue records by their DownloadID, preserving
// insertion order. Records with an empty DownloadID are each placed in their
// own single-record pack.
func GroupByDownloadID(records []arrclient.QueueRecord) []Pack {
	order := make([]string, 0, len(records))
	groups := make(map[string][]arrclient.QueueRecord, len(records))

	emptyIdx := 0
	for _, r := range records {
		key := r.DownloadID
		if key == "" {
			key = "_empty_" + strconv.Itoa(emptyIdx)
			emptyIdx++
		}
		if _, seen := groups[key]; !seen {
			order = append(order, key)
		}
		groups[key] = append(groups[key], r)
	}

	packs := make([]Pack, 0, len(order))
	for _, key := range order {
		did := key
		if strings.HasPrefix(did, "_empty_") {
			did = ""
		}
		packs = append(packs, Pack{
			DownloadID: did,
			Records:    groups[key],
		})
	}
	return packs
}
