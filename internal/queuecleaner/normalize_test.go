package queuecleaner

import (
	"testing"
)

func TestNormalizeCodec(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"x265", "x265"},
		{"H.265", "x265"},
		{"H265", "x265"},
		{"HEVC", "x265"},
		{"x264", "x264"},
		{"H.264", "x264"},
		{"H264", "x264"},
		{"AVC", "x264"},
		{"AV1", "AV1"},
		{"VP9", "VP9"},
		{"XviD", "XVID"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeCodec(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCodec(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSource(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"BluRay", "BluRay"},
		{"Blu-Ray", "BluRay"},
		{"BDRip", "BluRay"},
		{"BRRip", "BluRay"},
		{"Remux", "Remux"},
		{"WEB-DL", "WEB-DL"},
		{"WEBRip", "WEBRip"},
		{"WEB", "WEB-DL"},
		{"HDTV", "HDTV"},
		{"PDTV", "HDTV"},
		{"SDTV", "HDTV"},
		{"DVDRip", "DVD"},
		{"DVD", "DVD"},
		{"CAM", "CAM"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSource(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSource(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeHDR(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"HDR", "HDR"},
		{"HDR10", "HDR10"},
		{"HDR10+", "HDR10+"},
		{"DV", "DV"},
		{"DoVi", "DV"},
		{"Dolby Vision", "DV"},
		{"DolbyVision", "DV"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeHDR(tt.input)
			if got != tt.want {
				t.Errorf("normalizeHDR(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseReleaseEdgeCases(t *testing.T) {
	// Empty title
	info := ParseRelease("")
	if info.Resolution != "" || info.Codec != "" || info.Source != "" {
		t.Error("expected empty ReleaseInfo for empty title")
	}

	// Title with extension excluded from group
	info = ParseRelease("Movie.2024.1080p.BluRay-en.srt")
	if info.ReleaseGroup != "" {
		t.Errorf("expected empty group for srt subtitle suffix, got %q", info.ReleaseGroup)
	}

	// 576p resolution
	info = ParseRelease("Show.S01E01.576p.DVDRip.x264-GROUP")
	if info.Resolution != "576p" {
		t.Errorf("Resolution = %q, want 576p", info.Resolution)
	}
}

func TestParseReleaseREPACK(t *testing.T) {
	info := ParseRelease("Movie.2024.1080p.BluRay.RERIP.x265-GROUP")
	if !info.Repack {
		t.Error("expected Repack=true for RERIP release")
	}
}

func TestParseReleaseRemuxPriority(t *testing.T) {
	// Remux should take priority even after BluRay
	info := ParseRelease("Movie.2024.2160p.UHD.BluRay.Remux.x265-GROUP")
	if info.Source != "Remux" {
		t.Errorf("Source = %q, want Remux (should take priority over BluRay)", info.Source)
	}
}
