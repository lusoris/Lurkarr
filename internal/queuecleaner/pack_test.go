package queuecleaner

import (
	"testing"

	"github.com/lusoris/lurkarr/internal/arrclient"
)

func TestGroupByDownloadID(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, DownloadID: "abc123", Title: "Pack.S01.1080p"},
		{ID: 2, DownloadID: "abc123", Title: "Pack.S01.1080p"},
		{ID: 3, DownloadID: "abc123", Title: "Pack.S01.1080p"},
		{ID: 4, DownloadID: "def456", Title: "Single.Movie.2024"},
		{ID: 5, DownloadID: "", Title: "No.DownloadID.1"},
		{ID: 6, DownloadID: "", Title: "No.DownloadID.2"},
	}

	packs := GroupByDownloadID(records)

	if len(packs) != 4 {
		t.Fatalf("expected 4 packs, got %d", len(packs))
	}

	// Pack 1: abc123 with 3 records
	if packs[0].DownloadID != "abc123" || len(packs[0].Records) != 3 {
		t.Errorf("pack 0: got downloadID=%q records=%d, want abc123/3",
			packs[0].DownloadID, len(packs[0].Records))
	}
	if !packs[0].IsPack() {
		t.Error("pack 0 should be a pack")
	}

	// Pack 2: def456 with 1 record
	if packs[1].DownloadID != "def456" || len(packs[1].Records) != 1 {
		t.Errorf("pack 1: got downloadID=%q records=%d, want def456/1",
			packs[1].DownloadID, len(packs[1].Records))
	}
	if packs[1].IsPack() {
		t.Error("pack 1 should not be a pack")
	}

	// Packs 3-4: empty DownloadIDs, 1 record each
	for i := 2; i < 4; i++ {
		if packs[i].DownloadID != "" || len(packs[i].Records) != 1 {
			t.Errorf("pack %d: got downloadID=%q records=%d, want empty/1",
				i, packs[i].DownloadID, len(packs[i].Records))
		}
	}
}

func TestPack_QueueIDs(t *testing.T) {
	p := Pack{
		Records: []arrclient.QueueRecord{
			{ID: 10}, {ID: 20}, {ID: 30},
		},
	}
	ids := p.QueueIDs()
	if len(ids) != 3 || ids[0] != 10 || ids[1] != 20 || ids[2] != 30 {
		t.Errorf("QueueIDs = %v, want [10 20 30]", ids)
	}
}

func TestPack_Representative(t *testing.T) {
	p := Pack{
		Records: []arrclient.QueueRecord{
			{ID: 1, Title: "First"},
			{ID: 2, Title: "Second"},
		},
	}
	rep := p.Representative()
	if rep.ID != 1 || rep.Title != "First" {
		t.Errorf("Representative = {ID:%d Title:%q}, want {ID:1 Title:First}", rep.ID, rep.Title)
	}
}

func TestPack_AllImported(t *testing.T) {
	tests := []struct {
		name   string
		states []string
		want   bool
	}{
		{"all imported", []string{"imported", "imported"}, true},
		{"mixed", []string{"imported", "downloading"}, false},
		{"none imported", []string{"downloading", "downloading"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var records []arrclient.QueueRecord
			for _, s := range tt.states {
				records = append(records, arrclient.QueueRecord{TrackedDownloadState: s})
			}
			p := Pack{Records: records}
			if got := p.AllImported(); got != tt.want {
				t.Errorf("AllImported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPack_AllFilesDeleted(t *testing.T) {
	withMovie := func(hasFile bool) arrclient.QueueRecord {
		return arrclient.QueueRecord{
			TrackedDownloadState: "imported",
			Movie:               &arrclient.QueueMovie{HasFile: hasFile},
		}
	}

	tests := []struct {
		name    string
		records []arrclient.QueueRecord
		want    bool
	}{
		{
			"all deleted",
			[]arrclient.QueueRecord{withMovie(false), withMovie(false)},
			true,
		},
		{
			"one still has file",
			[]arrclient.QueueRecord{withMovie(false), withMovie(true)},
			false,
		},
		{
			"no enriched data",
			[]arrclient.QueueRecord{{TrackedDownloadState: "imported"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pack{Records: tt.records}
			if got := p.AllFilesDeleted(); got != tt.want {
				t.Errorf("AllFilesDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPack_AllUnmonitored(t *testing.T) {
	withMovie := func(monitored bool) arrclient.QueueRecord {
		return arrclient.QueueRecord{
			TrackedDownloadState: "downloading",
			Movie:               &arrclient.QueueMovie{Monitored: monitored},
		}
	}
	imported := arrclient.QueueRecord{TrackedDownloadState: "imported"}

	tests := []struct {
		name    string
		records []arrclient.QueueRecord
		want    bool
	}{
		{
			"all unmonitored",
			[]arrclient.QueueRecord{withMovie(false), withMovie(false)},
			true,
		},
		{
			"one monitored",
			[]arrclient.QueueRecord{withMovie(false), withMovie(true)},
			false,
		},
		{
			"imported skipped, rest unmonitored",
			[]arrclient.QueueRecord{imported, withMovie(false)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pack{Records: tt.records}
			if got := p.AllUnmonitored(); got != tt.want {
				t.Errorf("AllUnmonitored() = %v, want %v", got, tt.want)
			}
		})
	}
}
