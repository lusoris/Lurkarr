package blocklist

import (
	"regexp"
	"strings"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ReleaseParser is a function that parses a title into its components.
type ReleaseParser func(title string) ReleaseInfo

// ReleaseInfo holds parsed release attributes needed for blocklist matching.
type ReleaseInfo struct {
	ReleaseGroup string
}

// MatchResult describes why a record matched a blocklist rule.
type MatchResult struct {
	Matched bool
	Rule    database.BlocklistRule
}

// MaxRegexPatternLength is the maximum allowed length for a regex pattern.
// Go's regexp uses RE2 (linear-time), but very long patterns can still
// consume excessive memory during compilation.
const MaxRegexPatternLength = 1024

// Matcher checks queue records against a set of blocklist rules.
type Matcher struct {
	rules    []database.BlocklistRule
	compiled map[string]*regexp.Regexp // pre-compiled regexes keyed by rule ID
	parse    ReleaseParser
}

// NewMatcher creates a Matcher from a set of enabled rules.
// Invalid or overly long regex patterns are silently skipped.
// The parse function extracts release group info from titles.
func NewMatcher(rules []database.BlocklistRule, parse ReleaseParser) *Matcher {
	m := &Matcher{
		rules:    rules,
		compiled: make(map[string]*regexp.Regexp, len(rules)),
		parse:    parse,
	}
	for _, r := range rules {
		if r.PatternType == "title_regex" && len(r.Pattern) <= MaxRegexPatternLength {
			re, err := regexp.Compile(r.Pattern)
			if err == nil {
				m.compiled[r.ID.String()] = re
			}
		}
	}
	return m
}

// Check tests a queue record against all rules.
// Returns the first matching result, or a non-matched result.
func (m *Matcher) Check(record arrclient.QueueRecord) MatchResult {
	var info ReleaseInfo
	if m.parse != nil {
		info = m.parse(record.Title)
	}

	for _, r := range m.rules {
		if m.matches(r, record, info) {
			return MatchResult{Matched: true, Rule: r}
		}
	}
	return MatchResult{}
}

func (m *Matcher) matches(rule database.BlocklistRule, record arrclient.QueueRecord, info ReleaseInfo) bool {
	switch rule.PatternType {
	case "release_group":
		return info.ReleaseGroup != "" &&
			strings.EqualFold(info.ReleaseGroup, rule.Pattern)
	case "title_contains":
		return strings.Contains(
			strings.ToLower(record.Title),
			strings.ToLower(rule.Pattern),
		)
	case "title_regex":
		re, ok := m.compiled[rule.ID.String()]
		if !ok {
			return false
		}
		return re.MatchString(record.Title)
	case "indexer":
		return strings.EqualFold(record.Indexer, rule.Pattern)
	default:
		return false
	}
}
