package downloadclient

import "testing"

func TestUtorrentStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected string
	}{
		{"error", 128, "error"},
		{"error with started", 128 | 1, "error"},
		{"checking", 2, "checking"},
		{"paused", 16, "paused"},
		{"paused and started", 16 | 1, "paused"},
		{"downloading", 1, "downloading"},
		{"queued", 32, "queued"},
		{"stopped", 64, "stopped"},
		{"unknown", 0, "unknown"},
		{"unknown 4", 4, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utorrentStatusString(tt.status)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
