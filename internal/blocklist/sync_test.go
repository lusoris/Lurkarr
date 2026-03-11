package blocklist

import (
	"strings"
	"testing"
)

func TestParseBlocklist(t *testing.T) {
	input := `# Community Blocklist
# Lines starting with # are comments

EVO
SPARKS
group:FGT
regex:(?i)\bCAM\b
indexer:BadIndexer
contains:HDTS

# Empty lines are skipped

YIFY
`

	rules, err := ParseBlocklist(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := []struct {
		pattern     string
		patternType string
	}{
		{"EVO", "release_group"},
		{"SPARKS", "release_group"},
		{"FGT", "release_group"},
		{`(?i)\bCAM\b`, "title_regex"},
		{"BadIndexer", "indexer"},
		{"HDTS", "title_contains"},
		{"YIFY", "release_group"},
	}

	if len(rules) != len(expected) {
		t.Fatalf("got %d rules, want %d", len(rules), len(expected))
	}

	for i, exp := range expected {
		if rules[i].Pattern != exp.pattern {
			t.Errorf("rule %d: got pattern %q, want %q", i, rules[i].Pattern, exp.pattern)
		}
		if rules[i].PatternType != exp.patternType {
			t.Errorf("rule %d: got type %q, want %q", i, rules[i].PatternType, exp.patternType)
		}
		if rules[i].Reason != "community blocklist" {
			t.Errorf("rule %d: got reason %q, want %q", i, rules[i].Reason, "community blocklist")
		}
	}
}

func TestParseBlocklistEmpty(t *testing.T) {
	rules, err := ParseBlocklist(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Errorf("got %d rules for empty input, want 0", len(rules))
	}
}

func TestParseBlocklistOnlyComments(t *testing.T) {
	input := "# comment 1\n# comment 2\n"
	rules, err := ParseBlocklist(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Errorf("got %d rules for comments-only input, want 0", len(rules))
	}
}

func TestParseBlocklistEmptyPrefix(t *testing.T) {
	input := "group:\ncontains:\n"
	rules, err := ParseBlocklist(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Errorf("got %d rules for empty-prefix input, want 0", len(rules))
	}
}
