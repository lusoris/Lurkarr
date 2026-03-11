package queuecleaner

import (
	"testing"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

func TestScoreQueueItemBasic(t *testing.T) {
	item := &arrclient.QueueRecord{
		CustomFormatScore: 10,
		Size:              10 * 1024 * 1024 * 1024, // 10 GB
		Sizeleft:          5 * 1024 * 1024 * 1024,  // 5 GB left (50% done)
		Quality: &arrclient.QualityInfo{
			Quality: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 7, Name: "Bluray-1080p"},
		},
	}
	profile := &database.ScoringProfile{
		CustomFormatWeight:  2,
		PreferHigherQuality: true,
		PreferLargerSize:    true,
		SizeWeight:          5,
	}

	score := ScoreQueueItem(item, profile)

	// Expected: 10*2 (custom format) + 7*100 (quality) + 10*5 (size GB) + 50 (progress)
	expected := 20 + 700 + 50 + 50
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestScoreQueueItemNoQuality(t *testing.T) {
	item := &arrclient.QueueRecord{
		CustomFormatScore: 5,
		Size:              1024 * 1024 * 1024,
		Sizeleft:          0, // 100% done
	}
	profile := &database.ScoringProfile{
		CustomFormatWeight:  1,
		PreferHigherQuality: true,
		PreferLargerSize:    false,
	}

	score := ScoreQueueItem(item, profile)
	// 5*1 (custom format) + 0 (no quality) + 0 (not prefer larger) + 100 (progress)
	expected := 5 + 100
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestScoreQueueItemZero(t *testing.T) {
	item := &arrclient.QueueRecord{}
	profile := &database.ScoringProfile{}
	score := ScoreQueueItem(item, profile)
	if score != 0 {
		t.Errorf("ScoreQueueItem(empty) = %d, want 0", score)
	}
}

func TestScoreQueueItemSizeCap(t *testing.T) {
	item := &arrclient.QueueRecord{
		Size:     200 * 1024 * 1024 * 1024, // 200 GB
		Sizeleft: 0,
	}
	profile := &database.ScoringProfile{
		PreferLargerSize: true,
		SizeWeight:       10,
	}

	score := ScoreQueueItem(item, profile)
	// Size capped at 100 GB: 100*10 + 100 (progress)
	expected := 1000 + 100
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestFindDuplicatesHighest(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, MovieID: 100, Title: "Movie A Low", CustomFormatScore: 5, Size: 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 2, MovieID: 100, Title: "Movie A High", CustomFormatScore: 20, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 3, MovieID: 200, Title: "Movie B Only", CustomFormatScore: 10, Size: 1024 * 1024 * 1024, Sizeleft: 0},
	}
	profile := &database.ScoringProfile{
		Strategy:           "highest",
		CustomFormatWeight: 1,
	}

	dupes := FindDuplicates(records, profile)
	if len(dupes) != 1 {
		t.Fatalf("got %d duplicates, want 1", len(dupes))
	}
	if dupes[0].RemoveQueueID != 1 {
		t.Errorf("RemoveQueueID = %d, want 1", dupes[0].RemoveQueueID)
	}
	if dupes[0].KeepQueueID != 2 {
		t.Errorf("KeepQueueID = %d, want 2", dupes[0].KeepQueueID)
	}
	if dupes[0].MediaID != 100 {
		t.Errorf("MediaID = %d, want 100", dupes[0].MediaID)
	}
}

func TestFindDuplicatesAdequate(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, MovieID: 100, Title: "Below Threshold", CustomFormatScore: 2, Size: 1024 * 1024 * 1024, Sizeleft: 1024 * 1024 * 1024},
		{ID: 2, MovieID: 100, Title: "Above Threshold", CustomFormatScore: 500, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 3, MovieID: 100, Title: "Even Higher", CustomFormatScore: 1000, Size: 3 * 1024 * 1024 * 1024, Sizeleft: 0},
	}
	profile := &database.ScoringProfile{
		Strategy:           "adequate",
		AdequateThreshold:  400,
		CustomFormatWeight: 1,
	}

	dupes := FindDuplicates(records, profile)
	// Item 1 score: 2*1 + 0 (no progress since sizeleft=size) = 2 (below 400)
	// Item 2 score: 500*1 + 100 (progress) = 600 (above 400) — first to meet threshold
	// Item 3 score: 1000*1 + 100 (progress) = 1100
	// So item 2 is kept, items 1 and 3 are removed
	if len(dupes) != 2 {
		t.Fatalf("got %d duplicates, want 2", len(dupes))
	}
	for _, d := range dupes {
		if d.KeepQueueID != 2 {
			t.Errorf("expected KeepQueueID=2, got %d", d.KeepQueueID)
		}
	}
}

func TestFindDuplicatesNoDupes(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, MovieID: 100, Title: "Movie A"},
		{ID: 2, MovieID: 200, Title: "Movie B"},
		{ID: 3, EpisodeID: 300, Title: "Episode C"},
	}
	profile := &database.ScoringProfile{Strategy: "highest"}

	dupes := FindDuplicates(records, profile)
	if len(dupes) != 0 {
		t.Errorf("got %d duplicates, want 0", len(dupes))
	}
}

func TestFindDuplicatesSkipZeroMediaID(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, Title: "No Media ID"},
		{ID: 2, Title: "Also No Media ID"},
	}
	profile := &database.ScoringProfile{Strategy: "highest"}

	dupes := FindDuplicates(records, profile)
	if len(dupes) != 0 {
		t.Errorf("got %d duplicates, want 0 (no media ID)", len(dupes))
	}
}

func TestFindDuplicatesMultipleMediaTypes(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, EpisodeID: 50, Title: "Ep A", CustomFormatScore: 5, Size: 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 2, EpisodeID: 50, Title: "Ep A Better", CustomFormatScore: 20, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 3, AlbumID: 70, Title: "Album Solo"},
	}
	profile := &database.ScoringProfile{Strategy: "highest", CustomFormatWeight: 1}

	dupes := FindDuplicates(records, profile)
	if len(dupes) != 1 {
		t.Fatalf("got %d duplicates, want 1", len(dupes))
	}
	if dupes[0].RemoveQueueID != 1 {
		t.Errorf("should remove lower-scored episode")
	}
}
