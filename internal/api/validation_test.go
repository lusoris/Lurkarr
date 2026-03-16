package api

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// ValidateAppInstance validates app instance connection settings before saving.
func ValidateAppInstance(inst *database.AppInstance) error {
	if inst == nil {
		return errors.New("instance cannot be nil")
	}
	if inst.Name == "" {
		return errors.New("instance name required")
	}
	if inst.APIURL == "" {
		return errors.New("API URL required")
	}
	if inst.APIKey == "" {
		return errors.New("API key required")
	}
	return nil
}

// ValidateDownloadClientInstance validates download client settings.
func ValidateDownloadClientInstance(dc *database.DownloadClientInstance) error {
	if dc == nil {
		return errors.New("download client cannot be nil")
	}
	if dc.Name == "" {
		return errors.New("client name required")
	}
	if dc.ClientType == "" {
		return errors.New("client type required")
	}
	if dc.URL == "" {
		return errors.New("client URL required")
	}
	return nil
}

func TestValidateAppInstance(t *testing.T) {
	tests := []struct {
		name      string
		instance  *database.AppInstance
		wantError bool
		errMsg    string
	}{
		{
			name:      "nil instance",
			instance:  nil,
			wantError: true,
			errMsg:    "cannot be nil",
		},
		{
			name: "missing name",
			instance: &database.AppInstance{
				ID:      uuid.New(),
				AppType: database.AppSonarr,
				APIURL:  "http://sonarr:8989",
				APIKey:  "test-key",
			},
			wantError: true,
			errMsg:    "name required",
		},
		{
			name: "missing API URL",
			instance: &database.AppInstance{
				ID:      uuid.New(),
				AppType: database.AppSonarr,
				Name:    "Sonarr",
				APIKey:  "test-key",
			},
			wantError: true,
			errMsg:    "URL required",
		},
		{
			name: "missing API key",
			instance: &database.AppInstance{
				ID:      uuid.New(),
				AppType: database.AppSonarr,
				Name:    "Sonarr",
				APIURL:  "http://sonarr:8989",
			},
			wantError: true,
			errMsg:    "API key required",
		},
		{
			name: "valid instance",
			instance: &database.AppInstance{
				ID:      uuid.New(),
				AppType: database.AppSonarr,
				Name:    "Sonarr",
				APIURL:  "http://sonarr:8989",
				APIKey:  "test-key-123",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppInstance(tt.instance)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAppInstance() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && tt.errMsg != "" && (err == nil || !contains(err.Error(), tt.errMsg)) {
				t.Errorf("ValidateAppInstance() error message should contain %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestValidateDownloadClientInstance(t *testing.T) {
	tests := []struct {
		name      string
		client    *database.DownloadClientInstance
		wantError bool
		errMsg    string
	}{
		{
			name:      "nil client",
			client:    nil,
			wantError: true,
			errMsg:    "cannot be nil",
		},
		{
			name: "missing name",
			client: &database.DownloadClientInstance{
				ID:         uuid.New(),
				ClientType: "qbittorrent",
				URL:        "http://qbit:6881",
			},
			wantError: true,
			errMsg:    "name required",
		},
		{
			name: "missing type",
			client: &database.DownloadClientInstance{
				ID:   uuid.New(),
				Name: "qBittorrent",
				URL:  "http://qbit:6881",
			},
			wantError: true,
			errMsg:    "type required",
		},
		{
			name: "missing URL",
			client: &database.DownloadClientInstance{
				ID:         uuid.New(),
				Name:       "qBittorrent",
				ClientType: "qbittorrent",
			},
			wantError: true,
			errMsg:    "URL required",
		},
		{
			name: "valid client",
			client: &database.DownloadClientInstance{
				ID:         uuid.New(),
				Name:       "qBittorrent",
				ClientType: "qbittorrent",
				URL:        "http://qbittorrent:6881",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDownloadClientInstance(tt.client)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDownloadClientInstance() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && tt.errMsg != "" && (err == nil || !contains(err.Error(), tt.errMsg)) {
				t.Errorf("ValidateDownloadClientInstance() error message should contain %q, got %v", tt.errMsg, err)
			}
		})
	}
}

// Helper to check if error string contains substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
