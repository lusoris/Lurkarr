package database

import (
	"testing"
)

func TestAllAppTypes(t *testing.T) {
	types := AllAppTypes()
	if len(types) != 6 {
		t.Fatalf("AllAppTypes() returned %d types, want 6", len(types))
	}

	// AllAppTypes returns only arr apps, excluding Prowlarr (indexer manager)
	expected := map[AppType]bool{
		AppSonarr: true, AppRadarr: true, AppLidarr: true,
		AppReadarr: true, AppWhisparr: true, AppEros: true,
	}
	for _, at := range types {
		if !expected[at] {
			t.Errorf("unexpected app type in AllAppTypes: %s", at)
		}
	}

	// Verify Prowlarr is NOT in AllAppTypes
	for _, at := range types {
		if at == AppProwlarr {
			t.Errorf("Prowlarr should not be in AllAppTypes (it's an indexer, not an arr)")
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
		{"prowlarr", true}, // Still valid for config/connections, just not lurking
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
