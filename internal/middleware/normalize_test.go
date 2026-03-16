package middleware

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/api/instances", "/api/instances"},
		{"/api/instances/123", "/api/instances/:id"},
		{"/api/instances/550e8400-e29b-41d4-a716-446655440000", "/api/instances/:id"},
		{"/api/schedules/42/runs", "/api/schedules/:id/runs"},
		{"/", "/"},
		{"", "/"},
		{"/api", "/api"},
		{"/api/instances/123/history/456", "/api/instances/:id/history/:id"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := normalizePath(tt.in)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"/api/instances", []string{"api", "instances"}},
		{"/", nil},
		{"", nil},
		{"/a/b/c", []string{"a", "b", "c"}},
		{"/api//double", []string{"api", "double"}}, // double slash
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := splitPath(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitPath(%q) = %v (len %d), want %v (len %d)", tt.in, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitPath(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, "/"},
		{[]string{"api"}, "/api"},
		{[]string{"api", "instances"}, "/api/instances"},
		{[]string{"api", ":id", "runs"}, "/api/:id/runs"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := joinPath(tt.in)
			if got != tt.want {
				t.Errorf("joinPath(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"999999", true},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"", false},
		{"api", false},
		{"instances", false},
		{"abc-def", false},
		{"12a", false},
		{"550e8400-e29b-41d4-a716", false}, // truncated UUID
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := isID(tt.in)
			if got != tt.want {
				t.Errorf("isID(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIPRateLimiter_Stop(t *testing.T) {
	rl := NewIPRateLimiter(10, 20)
	// Allow should work before stop.
	if !rl.Allow("1.2.3.4") {
		t.Error("expected Allow to return true")
	}
	rl.Stop()
	// After Stop, the cleanup goroutine should have exited (no panic/hang).
}
