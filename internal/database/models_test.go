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
		{"normal key", "abcdef1234567890", "****7890"},
		{"short key", "abc", "****"},
		{"exactly 4", "abcd", "****"},
		{"5 chars", "abcde", "****bcde"},
		{"empty", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &AppInstance{APIKey: tt.apiKey}
			got := inst.MaskedAPIKey()
			if got != tt.want {
				t.Errorf("MaskedAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProwlarrSettingsMaskedAPIKey(t *testing.T) {
	tests := []struct {
		apiKey string
		want   string
	}{
		{"abcdef1234567890", "****7890"},
		{"abc", "****"},
		{"", "****"},
	}
	for _, tt := range tests {
		p := &ProwlarrSettings{APIKey: tt.apiKey}
		got := p.MaskedAPIKey()
		if got != tt.want {
			t.Errorf("ProwlarrSettings.MaskedAPIKey() with %q = %q, want %q", tt.apiKey, got, tt.want)
		}
	}
}

func TestSABnzbdSettingsMaskedAPIKey(t *testing.T) {
	tests := []struct {
		apiKey string
		want   string
	}{
		{"abcdef1234567890", "****7890"},
		{"abc", "****"},
		{"", "****"},
	}
	for _, tt := range tests {
		s := &SABnzbdSettings{APIKey: tt.apiKey}
		got := s.MaskedAPIKey()
		if got != tt.want {
			t.Errorf("SABnzbdSettings.MaskedAPIKey() with %q = %q, want %q", tt.apiKey, got, tt.want)
		}
	}
}
