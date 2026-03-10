package queuecleaner

import (
	"testing"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
)

func TestParseReleaseResolution(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.2160p.BluRay.x265-GROUP", "2160p"},
		{"Show.S01E02.1080p.WEB-DL.x264-SPARKS", "1080p"},
		{"Old.Movie.720p.HDTV.XviD-FGT", "720p"},
		{"Some.Title.480p.DVDRip.x264", "480p"},
		{"No.Resolution.Here-GROUP", ""},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.Resolution != tt.expected {
			t.Errorf("ParseRelease(%q).Resolution = %q, want %q", tt.title, info.Resolution, tt.expected)
		}
	}
}

func TestParseReleaseCodec(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.1080p.BluRay.x265-GROUP", "x265"},
		{"Movie.2024.1080p.BluRay.H.265-GROUP", "x265"},
		{"Movie.2024.1080p.BluRay.HEVC-GROUP", "x265"},
		{"Movie.2024.1080p.BluRay.x264-GROUP", "x264"},
		{"Movie.2024.1080p.BluRay.H.264-GROUP", "x264"},
		{"Movie.2024.1080p.BluRay.AV1-GROUP", "AV1"},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.Codec != tt.expected {
			t.Errorf("ParseRelease(%q).Codec = %q, want %q", tt.title, info.Codec, tt.expected)
		}
	}
}

func TestParseReleaseSource(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.1080p.BluRay.x265-GROUP", "BluRay"},
		{"Movie.2024.1080p.Blu-Ray.x265-GROUP", "BluRay"},
		{"Movie.2024.1080p.Remux.x265-GROUP", "Remux"},
		{"Movie.2024.1080p.WEB-DL.x265-GROUP", "WEB-DL"},
		{"Movie.2024.1080p.WEBRip.x265-GROUP", "WEBRip"},
		{"Movie.2024.1080p.HDTV.x264-GROUP", "HDTV"},
		{"Movie.2024.720p.DVDRip.x264-GROUP", "DVD"},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.Source != tt.expected {
			t.Errorf("ParseRelease(%q).Source = %q, want %q", tt.title, info.Source, tt.expected)
		}
	}
}

func TestParseReleaseGroup(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.1080p.BluRay.x265-SPARKS", "SPARKS"},
		{"Movie.2024.1080p.BluRay.x265-FGT", "FGT"},
		{"Movie.2024.1080p.BluRay.x265-GROUP.mkv", "GROUP"},
		{"Movie No Group", ""},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.ReleaseGroup != tt.expected {
			t.Errorf("ParseRelease(%q).ReleaseGroup = %q, want %q", tt.title, info.ReleaseGroup, tt.expected)
		}
	}
}

func TestParseReleaseHDR(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.2160p.BluRay.HDR.x265-GROUP", "HDR"},
		{"Movie.2024.2160p.BluRay.HDR10.x265-GROUP", "HDR10"},
		{"Movie.2024.2160p.BluRay.DV.x265-GROUP", "DV"},
		{"Movie.2024.1080p.BluRay.x265-GROUP", ""},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.HDR != tt.expected {
			t.Errorf("ParseRelease(%q).HDR = %q, want %q", tt.title, info.HDR, tt.expected)
		}
	}
}

func TestParseReleaseProperRepack(t *testing.T) {
	info := ParseRelease("Movie.2024.1080p.BluRay.PROPER.x265-GROUP")
	if !info.Proper {
		t.Error("expected Proper=true for PROPER release")
	}
	if info.Repack {
		t.Error("expected Repack=false for PROPER release")
	}

	info2 := ParseRelease("Movie.2024.1080p.BluRay.REPACK.x265-GROUP")
	if info2.Proper {
		t.Error("expected Proper=false for REPACK release")
	}
	if !info2.Repack {
		t.Error("expected Repack=true for REPACK release")
	}
}

func TestParseReleaseAudio(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Movie.2024.2160p.BluRay.Atmos.x265-GROUP", "Atmos"},
		{"Movie.2024.2160p.BluRay.TrueHD.x265-GROUP", "TrueHD"},
		{"Movie.2024.2160p.BluRay.DTS-HD.MA.x265-GROUP", "DTS-HD.MA"},
		{"Movie.2024.1080p.BluRay.FLAC.x265-GROUP", "FLAC"},
		{"Movie.2024.1080p.WEB-DL.AAC.x264-GROUP", "AAC"},
	}
	for _, tt := range tests {
		info := ParseRelease(tt.title)
		if info.Audio != tt.expected {
			t.Errorf("ParseRelease(%q).Audio = %q, want %q", tt.title, info.Audio, tt.expected)
		}
	}
}

func TestParseReleaseFullTitle(t *testing.T) {
	info := ParseRelease("The.Matrix.1999.2160p.UHD.BluRay.Remux.HDR10.HEVC.Atmos.TrueHD.7.1-FraMeSToR")
	if info.Resolution != "2160p" {
		t.Errorf("Resolution = %q, want 2160p", info.Resolution)
	}
	if info.Source != "Remux" {
		t.Errorf("Source = %q, want Remux", info.Source)
	}
	if info.Codec != "x265" {
		t.Errorf("Codec = %q, want x265", info.Codec)
	}
	if info.HDR != "HDR10" {
		t.Errorf("HDR = %q, want HDR10", info.HDR)
	}
	if info.ReleaseGroup != "FraMeSToR" {
		t.Errorf("ReleaseGroup = %q, want FraMeSToR", info.ReleaseGroup)
	}
}

func TestParseTimeleft(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"01:30:00", 1*time.Hour + 30*time.Minute},
		{"00:05:30", 5*time.Minute + 30*time.Second},
		{"2.03:00:00", 51 * time.Hour},
		{"00:00:00", 0},
		{"", 0},
		{"invalid", 0},
	}
	for _, tt := range tests {
		got := parseTimeleft(tt.input)
		if got != tt.expected {
			t.Errorf("parseTimeleft(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestHasImportFailure(t *testing.T) {
	tests := []struct {
		name     string
		record   arrclient.QueueRecord
		expected bool
	}{
		{
			name: "healthy download",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "ok",
				TrackedDownloadState:  "downloading",
			},
			expected: false,
		},
		{
			name: "import pending with failure message",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Title: "test", Messages: []string{"Import failed - no matching series"}},
				},
			},
			expected: true,
		},
		{
			name: "import failed state",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importFailed",
				StatusMessages: []arrclient.StatusMessage{
					{Title: "test", Messages: []string{"Unable to import: file not found"}},
				},
			},
			expected: true,
		},
		{
			name: "warning but no import failure message",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Title: "test", Messages: []string{"Download is slow"}},
				},
			},
			expected: false,
		},
		{
			name: "sample detection",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Title: "test", Messages: []string{"File is a sample"}},
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasImportFailure(tt.record)
			if got != tt.expected {
				t.Errorf("hasImportFailure() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestImportFailureReason(t *testing.T) {
	record := arrclient.QueueRecord{
		StatusMessages: []arrclient.StatusMessage{
			{Title: "test", Messages: []string{"Import failed - no matching series"}},
		},
	}
	reason := importFailureReason(record)
	if reason != "Import failed - no matching series" {
		t.Errorf("importFailureReason() = %q, want %q", reason, "Import failed - no matching series")
	}

	empty := arrclient.QueueRecord{}
	if r := importFailureReason(empty); r != "unknown_import_failure" {
		t.Errorf("importFailureReason(empty) = %q, want %q", r, "unknown_import_failure")
	}
}
