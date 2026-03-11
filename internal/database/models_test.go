package database

import (
	"testing"
)

func TestAllAppTypes(t *testing.T) {
	types := AllAppTypes()
	if len(types) != 7 {
		t.Fatalf("AllAppTypes() returned %d types, want 7", len(types))
	}

	expected := map[AppType]bool{
		AppSonarr: true, AppRadarr: true, AppLidarr: true,
		AppReadarr: true, AppWhisparr: true, AppEros: true, AppProwlarr: true,
	}
	for _, at := range types {
		if !expected[at] {
			t.Errorf("unexpected app type: %s", at)
		}
	}
}

func TestValidAppType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sonarr", true},
		{"radarr", true},
		{"lidarr", true},
		{"readarr", true},
		{"whisparr", true},
		{"eros", true},
		{"prowlarr", true},
		{"invalid", false},
		{"Sonarr", false},
		{"", false},
		{"swaparr", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ValidAppType(tt.input)
			if got != tt.want {
				t.Errorf("ValidAppType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAppInstanceMaskedAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{"long key", "abcdef123456", "****3456"},
		{"short key", "ab", "****"},
		{"empty key", "", "****"},
		{"exactly 4", "abcd", "****"},
		{"5 chars", "abcde", "****bcde"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppInstance{APIKey: tt.apiKey}
			if got := a.MaskedAPIKey(); got != tt.want {
				t.Errorf("MaskedAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProwlarrSettingsMaskedAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{"long key", "abcdef123456", "****3456"},
		{"short key", "ab", "****"},
		{"empty key", "", "****"},
		{"exactly 4", "abcd", "****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProwlarrSettings{APIKey: tt.apiKey}
			if got := p.MaskedAPIKey(); got != tt.want {
				t.Errorf("MaskedAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSABnzbdSettingsMaskedAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{"long key", "abcdef123456", "****3456"},
		{"short key", "ab", "****"},
		{"empty key", "", "****"},
		{"exactly 4", "abcd", "****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SABnzbdSettings{APIKey: tt.apiKey}
			if got := s.MaskedAPIKey(); got != tt.want {
				t.Errorf("MaskedAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
