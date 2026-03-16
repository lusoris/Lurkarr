package blocklist

import (
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

var reGroup = regexp.MustCompile(`(?i)-([A-Za-z0-9]+)(?:\.[a-z]{2,4})?$`)

func testParser(title string) ReleaseInfo {
	if m := reGroup.FindStringSubmatch(title); len(m) > 1 {
		return ReleaseInfo{ReleaseGroup: m[1]}
	}
	return ReleaseInfo{}
}

func rule(patternType, pattern string) database.BlocklistRule {
	return database.BlocklistRule{
		ID:          uuid.New(),
		Pattern:     pattern,
		PatternType: patternType,
		Reason:      "test",
		Enabled:     true,
	}
}

func TestMatcherReleaseGroup(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("release_group", "EVO")}, testParser)
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"match", "Movie.2024.1080p.WEB-DL.x264-EVO", true},
		{"case insensitive", "Movie.2024.1080p.WEB-DL.x264-evo", true},
		{"no match", "Movie.2024.1080p.WEB-DL.x264-SPARKS", false},
		{"no group", "Movie 2024 1080p", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Check(arrclient.QueueRecord{Title: tt.title})
			if result.Matched != tt.want {
				t.Errorf("got matched=%v, want %v for title %q", result.Matched, tt.want, tt.title)
			}
		})
	}
}

func TestMatcherTitleContains(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("title_contains", "cam")}, testParser)
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"match lower", "Movie.2024.CAMRip", true},
		{"match mixed", "Some.camera.footage", true},
		{"no match", "Movie.2024.1080p.BluRay", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Check(arrclient.QueueRecord{Title: tt.title})
			if result.Matched != tt.want {
				t.Errorf("got matched=%v, want %v for title %q", result.Matched, tt.want, tt.title)
			}
		})
	}
}

func TestMatcherTitleRegex(t *testing.T) {
	pattern := `(?i)\b(TS|TELESYNC|CAM)\b`
	m := NewMatcher([]database.BlocklistRule{rule("title_regex", pattern)}, testParser)
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"match TS", "Movie.2024.TS.x264", true},
		{"match CAM", "Movie.2024.CAM.x264", true},
		{"no match", "Movie.2024.BluRay.1080p", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Check(arrclient.QueueRecord{Title: tt.title})
			if result.Matched != tt.want {
				t.Errorf("got matched=%v, want %v for title %q", result.Matched, tt.want, tt.title)
			}
		})
	}
}

func TestMatcherInvalidRegex(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("title_regex", "[invalid")}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: "anything"})
	if result.Matched {
		t.Error("invalid regex should not match")
	}
}

func TestMatcherRegexTooLong(t *testing.T) {
	longPattern := strings.Repeat("a", MaxRegexPatternLength+1)
	m := NewMatcher([]database.BlocklistRule{rule("title_regex", longPattern)}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: strings.Repeat("a", MaxRegexPatternLength+1)})
	if result.Matched {
		t.Error("regex exceeding MaxRegexPatternLength should not be compiled or match")
	}
}

func TestMatcherRegexAtMaxLength(t *testing.T) {
	// A valid regex at exactly the limit should still work.
	pattern := strings.Repeat("a", MaxRegexPatternLength)
	m := NewMatcher([]database.BlocklistRule{rule("title_regex", pattern)}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: strings.Repeat("a", MaxRegexPatternLength)})
	if !result.Matched {
		t.Error("regex at exactly MaxRegexPatternLength should compile and match")
	}
}

func TestMatcherIndexer(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("indexer", "BadIndexer")}, testParser)
	tests := []struct {
		name    string
		indexer string
		want    bool
	}{
		{"match", "BadIndexer", true},
		{"case insensitive", "badindexer", true},
		{"no match", "GoodIndexer", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Check(arrclient.QueueRecord{Title: "something", Indexer: tt.indexer})
			if result.Matched != tt.want {
				t.Errorf("got matched=%v, want %v for indexer %q", result.Matched, tt.want, tt.indexer)
			}
		})
	}
}

func TestMatcherMultipleRules(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{
		rule("release_group", "EVO"),
		rule("indexer", "BadIndexer"),
		rule("title_contains", "sample"),
	}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: "Movie.2024.1080p", Indexer: "BadIndexer"})
	if !result.Matched {
		t.Error("expected match on indexer rule")
	}
	if result.Rule.PatternType != "indexer" {
		t.Errorf("expected indexer rule, got %s", result.Rule.PatternType)
	}
}

func TestMatcherNoRules(t *testing.T) {
	m := NewMatcher(nil, testParser)
	result := m.Check(arrclient.QueueRecord{Title: "anything"})
	if result.Matched {
		t.Error("no rules should not match")
	}
}

func TestMatcherUnknownPatternType(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("unknown_type", "pattern")}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: "anything"})
	if result.Matched {
		t.Error("unknown pattern type should not match")
	}
}

func TestMatcherFilePatternTitle(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("file_pattern", `(?i)\.(exe|scr|bat|cmd|msi)$`)}, testParser)
	result := m.Check(arrclient.QueueRecord{Title: "Movie.2024.1080p.exe"})
	if !result.Matched {
		t.Error("file_pattern should match malware extension in title")
	}
}

func TestMatcherFilePatternStatusMessage(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("file_pattern", `(?i)\.(exe|scr|bat|cmd|msi)$`)}, testParser)
	result := m.Check(arrclient.QueueRecord{
		Title: "Movie.2024.1080p.BluRay-GROUP",
		StatusMessages: []arrclient.StatusMessage{
			{Title: "movie.2024.1080p.bluray.scr", Messages: []string{"Not a valid media file"}},
		},
	})
	if !result.Matched {
		t.Error("file_pattern should match malware extension in status message filename")
	}
}

func TestMatcherFilePatternNoMatch(t *testing.T) {
	m := NewMatcher([]database.BlocklistRule{rule("file_pattern", `(?i)\.(exe|scr|bat|cmd|msi)$`)}, testParser)
	result := m.Check(arrclient.QueueRecord{
		Title: "Movie.2024.1080p.BluRay-GROUP",
		StatusMessages: []arrclient.StatusMessage{
			{Title: "movie.2024.1080p.bluray.mkv", Messages: []string{"Imported"}},
		},
	})
	if result.Matched {
		t.Error("file_pattern should not match legitimate media extension")
	}
}
