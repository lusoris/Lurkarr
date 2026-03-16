package downloadclient

import (
	"testing"
	"time"

	gort "github.com/autobrr/go-rtorrent"
)

func TestRtorrentProgress(t *testing.T) {
	tests := []struct {
		name     string
		status   gort.Status
		expected float64
	}{
		{"zero size", gort.Status{Size: 0, CompletedBytes: 0}, 0},
		{"half done", gort.Status{Size: 1000, CompletedBytes: 500}, 0.5},
		{"complete", gort.Status{Size: 1000, CompletedBytes: 1000}, 1.0},
		{"quarter", gort.Status{Size: 400, CompletedBytes: 100}, 0.25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rtorrentProgress(tt.status)
			if got != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, got)
			}
		})
	}
}

func TestRtorrentETA(t *testing.T) {
	tests := []struct {
		name     string
		status   gort.Status
		expected int64
	}{
		{"no speed", gort.Status{Size: 1000, CompletedBytes: 0, DownRate: 0}, 0},
		{"complete", gort.Status{Size: 1000, CompletedBytes: 1000, DownRate: 100}, 0},
		{"partial", gort.Status{Size: 1000, CompletedBytes: 500, DownRate: 100}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rtorrentETA(tt.status)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestRtorrentCompletedAt(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		torrent := gort.Torrent{Finished: time.Time{}}
		if got := rtorrentCompletedAt(torrent); got != 0 {
			t.Errorf("expected 0, got %d", got)
		}
	})

	t.Run("finished", func(t *testing.T) {
		fin := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		torrent := gort.Torrent{Finished: fin}
		if got := rtorrentCompletedAt(torrent); got != fin.Unix() {
			t.Errorf("expected %d, got %d", fin.Unix(), got)
		}
	})
}

func TestRtorrentStatusString(t *testing.T) {
	tests := []struct {
		name     string
		torrent  gort.Torrent
		status   gort.Status
		active   bool
		expected string
	}{
		{"completed seeding", gort.Torrent{}, gort.Status{Completed: true, UpRate: 100}, true, "seeding"},
		{"completed idle", gort.Torrent{}, gort.Status{Completed: true}, false, "seeding"},
		{"paused", gort.Torrent{}, gort.Status{}, false, "paused"},
		{"downloading active", gort.Torrent{}, gort.Status{DownRate: 100}, true, "downloading"},
		{"downloading no speed", gort.Torrent{}, gort.Status{}, true, "downloading"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rtorrentStatusString(tt.torrent, tt.status, tt.active)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
