package queuecleaner

import (
	"testing"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

func TestScoreQueueItemBasic(t *testing.T) {
	item := &arrclient.QueueRecord{
		CustomFormatScore: 10,
		Title:             "Movie.2024.1080p.BluRay.x264-GROUP",
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
		CustomFormatWeight: 2,
		ResolutionWeight:   50,
		SourceWeight:       30,
		RevisionBonus:      50,
		PreferLargerSize:   true,
		SizeWeight:         5,
	}

	score := ScoreQueueItem(item, profile)

	// Expected: 10*2 (CF) + 3*50 (1080p=rank3) + 4*30 (BluRay=rank4) + 0 (no revision) + 10*5 (size) + 50 (progress)
	expected := 20 + 150 + 120 + 50 + 50
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestScoreQueueItemNoQuality(t *testing.T) {
	item := &arrclient.QueueRecord{
		CustomFormatScore: 5,
		Title:             "movie.file",
		Size:              1024 * 1024 * 1024,
		Sizeleft:          0, // 100% done
	}
	profile := &database.ScoringProfile{
		CustomFormatWeight: 1,
		ResolutionWeight:   50,
		SourceWeight:       30,
	}

	score := ScoreQueueItem(item, profile)
	// 5*1 (CF) + 0*50 (no resolution) + (-1)*30 (unknown source) + 100 (progress)
	expected := 5 + 0 - 30 + 100
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
		Title:    "movie.file",
		Size:     200 * 1024 * 1024 * 1024, // 200 GB
		Sizeleft: 0,
	}
	profile := &database.ScoringProfile{
		PreferLargerSize: true,
		SizeWeight:       10,
	}

	score := ScoreQueueItem(item, profile)
	// Size capped at 100 GB: 100*10 + 0*0 (no resolution) + (-1)*0 (no source, weight=0) + 100 (progress)
	expected := 1000 + 100
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestScoreQueueItemRevisionBonus(t *testing.T) {
	item := &arrclient.QueueRecord{
		Title: "Movie.2024.1080p.WEB-DL.PROPER.x264-GROUP",
		Quality: &arrclient.QualityInfo{
			Quality: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 3, Name: "WEB-DL 1080p"},
			Revision: struct {
				Version int `json:"version"`
			}{Version: 2}, // PROPER
		},
	}
	profile := &database.ScoringProfile{
		ResolutionWeight: 50,
		SourceWeight:     30,
		RevisionBonus:    50,
	}

	score := ScoreQueueItem(item, profile)
	// 0 (CF) + 3*50 (1080p) + 3*30 (WEB-DL=rank3) + 50 (revision bonus) + 0 (size/progress)
	expected := 150 + 90 + 50
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestScoreQueueItemResolutionFromTitle(t *testing.T) {
	// Quality name doesn't contain resolution, must parse from title
	item := &arrclient.QueueRecord{
		Title: "Movie.2024.2160p.Remux.HEVC.Atmos-GROUP",
		Quality: &arrclient.QualityInfo{
			Quality: struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: 31, Name: "Remux"},
		},
	}
	profile := &database.ScoringProfile{
		ResolutionWeight: 50,
		SourceWeight:     30,
	}

	score := ScoreQueueItem(item, profile)
	// 0 (CF) + 4*50 (2160p from title) + 5*30 (Remux=rank5 from title) + 0 (no revision/size/progress)
	expected := 200 + 150
	if score != expected {
		t.Errorf("ScoreQueueItem() = %d, want %d", score, expected)
	}
}

func TestResolutionRank(t *testing.T) {
	cases := []struct {
		res  string
		want int
	}{
		{"2160p", 4}, {"1080p", 3}, {"720p", 2}, {"480p", 1}, {"4320p", 5}, {"", 0},
	}
	for _, c := range cases {
		if got := resolutionRank(c.res); got != c.want {
			t.Errorf("resolutionRank(%q) = %d, want %d", c.res, got, c.want)
		}
	}
}

func TestSourceRank(t *testing.T) {
	cases := []struct {
		src  string
		want int
	}{
		{"Remux", 5}, {"BluRay", 4}, {"WEB-DL", 3}, {"WEBRip", 2}, {"HDTV", 1}, {"DVD", 0}, {"CAM", -1}, {"", -1},
	}
	for _, c := range cases {
		if got := sourceRank(c.src); got != c.want {
			t.Errorf("sourceRank(%q) = %d, want %d", c.src, got, c.want)
		}
	}
}

func TestFindDuplicatesHighest(t *testing.T) {
	records := []arrclient.QueueRecord{
		{ID: 1, MovieID: 100, Title: "Movie.A.720p.HDTV.x264", CustomFormatScore: 5, Size: 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 2, MovieID: 100, Title: "Movie.A.1080p.BluRay.x264", CustomFormatScore: 20, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 3, MovieID: 200, Title: "Movie.B.1080p.WEB-DL.x264", CustomFormatScore: 10, Size: 1024 * 1024 * 1024, Sizeleft: 0},
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
		{ID: 1, MovieID: 100, Title: "Movie.480p.CAM", CustomFormatScore: 2, Size: 1024 * 1024 * 1024, Sizeleft: 1024 * 1024 * 1024},
		{ID: 2, MovieID: 100, Title: "Movie.1080p.BluRay.x264", CustomFormatScore: 500, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 3, MovieID: 100, Title: "Movie.2160p.Remux.x265", CustomFormatScore: 1000, Size: 3 * 1024 * 1024 * 1024, Sizeleft: 0},
	}
	profile := &database.ScoringProfile{
		Strategy:           "adequate",
		AdequateThreshold:  400,
		CustomFormatWeight: 1,
	}

	dupes := FindDuplicates(records, profile)
	// Item 2 is first to meet threshold (CF 500 alone > 400). Items 1 and 3 removed.
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
		{ID: 1, EpisodeID: 50, Title: "Ep.S01E01.720p.HDTV", CustomFormatScore: 5, Size: 1024 * 1024 * 1024, Sizeleft: 0},
		{ID: 2, EpisodeID: 50, Title: "Ep.S01E01.1080p.BluRay", CustomFormatScore: 20, Size: 2 * 1024 * 1024 * 1024, Sizeleft: 0},
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
