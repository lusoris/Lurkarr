package queuecleaner

import (
	"regexp"
	"strings"
)

// ReleaseInfo holds parsed attributes from a release title.
type ReleaseInfo struct {
	Resolution   string // e.g. "2160p", "1080p", "720p", "480p"
	Codec        string // e.g. "x265", "x264", "AV1", "HEVC", "VP9"
	Source       string // e.g. "BluRay", "WEB-DL", "WEBRip", "HDTV", "Remux"
	Audio        string // e.g. "Atmos", "TrueHD", "DTS-HD MA", "FLAC", "DD5.1", "AAC"
	ReleaseGroup string // e.g. "SPARKS", "FGT"
	HDR          string // e.g. "HDR", "HDR10", "HDR10+", "DV" (Dolby Vision)
	Proper       bool
	Repack       bool
}

var (
	reResolution = regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|576p|4320p)\b`)
	reCodec      = regexp.MustCompile(`(?i)\b(x265|h\.?265|HEVC|x264|h\.?264|AVC|AV1|VP9|MPEG-?2|XviD|DivX)\b`)
	reSource     = regexp.MustCompile(`(?i)\b(Remux|Blu-?Ray|BDRip|BRRip|WEB-DL|WEBRip|WEB|HDTV|DVDRip|DVD|PDTV|SDTV|HDRip|CAM|TS|TELESYNC|SCR|SCREENER)\b`)
	reAudio      = regexp.MustCompile(`(?i)(Atmos|TrueHD|DTS-HD[\. ]?MA|DTS-HD|DTS|DD[P\+]?[\. ]?(?:5\.1|7\.1|2\.0)?|AC3|AAC(?:\d\.\d)?|FLAC|EAC3|LPCM|PCM|Opus|MP3)`)
	reHDR        = regexp.MustCompile(`(?i)\b(HDR10\+|HDR10|HDR|DV|DoVi|Dolby[. ]?Vision)\b`)
	reGroup      = regexp.MustCompile(`(?i)-([A-Za-z0-9]+)(?:\.[a-z]{2,4})?$`)
	reProper     = regexp.MustCompile(`(?i)\b(PROPER)\b`)
	reRepack     = regexp.MustCompile(`(?i)\b(REPACK|RERIP)\b`)
)

// ParseRelease extracts quality attributes from a release title string.
func ParseRelease(title string) ReleaseInfo {
	var info ReleaseInfo

	if m := reResolution.FindStringSubmatch(title); len(m) > 1 {
		info.Resolution = strings.ToLower(m[1])
	}

	if m := reCodec.FindStringSubmatch(title); len(m) > 1 {
		info.Codec = normalizeCodec(m[1])
	}

	if m := reSource.FindStringSubmatch(title); len(m) > 1 {
		info.Source = normalizeSource(m[1])
		// Remux takes priority — it may appear after another source keyword (e.g. "BluRay.Remux")
		if info.Source != "Remux" && strings.Contains(strings.ToLower(title), "remux") {
			info.Source = "Remux"
		}
	}

	if m := reAudio.FindStringSubmatch(title); len(m) > 1 {
		info.Audio = m[1]
	}

	if m := reHDR.FindStringSubmatch(title); len(m) > 1 {
		info.HDR = normalizeHDR(m[1])
	}

	if m := reGroup.FindStringSubmatch(title); len(m) > 1 {
		g := m[1]
		// Exclude common file extensions and false positives
		lower := strings.ToLower(g)
		if lower != "en" && lower != "srt" && lower != "nfo" && lower != "txt" {
			info.ReleaseGroup = g
		}
	}

	info.Proper = reProper.MatchString(title)
	info.Repack = reRepack.MatchString(title)

	return info
}

func normalizeCodec(s string) string {
	upper := strings.ToUpper(s)
	switch {
	case strings.Contains(upper, "265") || upper == "HEVC":
		return "x265"
	case strings.Contains(upper, "264") || upper == "AVC":
		return "x264"
	default:
		return strings.ToUpper(s)
	}
}

func normalizeSource(s string) string {
	upper := strings.ToUpper(strings.ReplaceAll(s, "-", ""))
	switch {
	case strings.HasPrefix(upper, "BLU") || upper == "BDRIP" || upper == "BRRIP":
		return "BluRay"
	case upper == "REMUX":
		return "Remux"
	case upper == "WEBDL":
		return "WEB-DL"
	case upper == "WEBRIP":
		return "WEBRip"
	case upper == "WEB":
		return "WEB-DL"
	case upper == "HDTV" || upper == "PDTV" || upper == "SDTV":
		return "HDTV"
	case strings.Contains(upper, "DVD"):
		return "DVD"
	default:
		return s
	}
}

func normalizeHDR(s string) string {
	upper := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	switch {
	case strings.Contains(upper, "HDR10+"):
		return "HDR10+"
	case strings.Contains(upper, "HDR10"):
		return "HDR10"
	case upper == "HDR":
		return "HDR"
	case strings.Contains(upper, "DV") || strings.Contains(upper, "DOVI") || strings.Contains(upper, "DOLBY"):
		return "DV"
	default:
		return s
	}
}
