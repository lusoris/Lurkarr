package api

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// ValidateNotificationTarget validates notification target configuration.
func ValidateNotificationTarget(target string, settings map[string]string) error {
	if settings == nil {
		return nil // Optional
	}

	switch target {
	case "discord":
		// Requires: discord_webhook_url
		if settings["discord_webhook_url"] == "" && settings["enabled"] == "true" {
			return NewBadRequest("Discord webhook URL required")
		}

	case "telegram":
		// Requires: telegram_token, telegram_chat_id
		if settings["telegram_token"] == "" && settings["enabled"] == "true" {
			return NewBadRequest("Telegram token required")
		}
		if settings["telegram_chat_id"] == "" && settings["enabled"] == "true" {
			return NewBadRequest("Telegram chat ID required")
		}

	case "email":
		// Requires: email_address, smtp_host, smtp_port
		email := settings["email_address"]
		if email == "" && settings["enabled"] == "true" {
			return NewBadRequest("Email address required")
		}
		// Basic email validation
		if email != "" && !isValidEmail(email) {
			return NewBadRequest("Invalid email address format")
		}
	}

	return nil
}

// ValidateSchedule validates schedule configuration.
func ValidateSchedule(schedule *database.Schedule) error {
	if schedule == nil {
		return NewBadRequest("schedule cannot be nil")
	}
	if schedule.Action == "" {
		return NewBadRequest("action required")
	}
	if schedule.AppType == "" {
		return NewBadRequest("app type required")
	}
	if len(schedule.Days) == 0 {
		return NewBadRequest("days required")
	}
	if schedule.Hour < 0 || schedule.Hour > 23 {
		return NewBadRequest("hour must be 0-23")
	}
	if schedule.Minute < 0 || schedule.Minute > 59 {
		return NewBadRequest("minute must be 0-59")
	}

	// Validate action format
	validActions := []string{
		"lurk_missing", "lurk_upgrade", "lurk_all",
		"clean_queue", "enable_instances", "disable_instances",
	}
	actionFound := false
	for _, valid := range validActions {
		if schedule.Action == valid {
			actionFound = true
			break
		}
	}
	if !actionFound {
		return NewBadRequest("invalid action: " + schedule.Action)
	}

	return nil
}

// Helper functions
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Test for ValidateNotificationTarget
func TestValidateNotificationTarget(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		settings  map[string]string
		wantError bool
		errMsg    string
	}{
		{
			"nil settings is ok",
			"discord",
			nil,
			false,
			"",
		},
		{
			"discord with webhook URL",
			"discord",
			map[string]string{
				"enabled":             "true",
				"discord_webhook_url": "https://discordapp.com/api/webhooks/1234/abcd",
			},
			false,
			"",
		},
		{
			"discord missing webhook URL",
			"discord",
			map[string]string{
				"enabled": "true",
			},
			true,
			"webhook URL required",
		},
		{
			"telegram with token and chat ID",
			"telegram",
			map[string]string{
				"enabled":          "true",
				"telegram_token":   "123456789:ABCDef...",
				"telegram_chat_id": "987654321",
			},
			false,
			"",
		},
		{
			"telegram missing token",
			"telegram",
			map[string]string{
				"enabled":          "true",
				"telegram_chat_id": "987654321",
			},
			true,
			"token required",
		},
		{
			"email with valid address",
			"email",
			map[string]string{
				"enabled":       "true",
				"email_address": "user@example.com",
				"smtp_host":     "smtp.example.com",
				"smtp_port":     "587",
			},
			false,
			"",
		},
		{
			"email with invalid address format",
			"email",
			map[string]string{
				"enabled":       "true",
				"email_address": "not-an-email",
			},
			true,
			"Invalid email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotificationTarget(tt.target, tt.settings)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNotificationTarget() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && tt.errMsg != "" && (err == nil || !contains(err.Error(), tt.errMsg)) {
				t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

// Test for ValidateSchedule
func TestValidateSchedule(t *testing.T) {
	tests := []struct {
		name      string
		schedule  *database.Schedule
		wantError bool
		errMsg    string
	}{
		{
			"nil schedule",
			nil,
			true,
			"cannot be nil",
		},
		{
			"missing action",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Days:    []string{"monday"},
				Hour:    10,
				Minute:  0,
			},
			true,
			"action required",
		},
		{
			"missing app type",
			&database.Schedule{
				ID:     uuid.New(),
				Action: "lurk_missing",
				Days:   []string{"monday"},
				Hour:   10,
				Minute: 0,
			},
			true,
			"app type required",
		},
		{
			"missing days",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Action:  "lurk_missing",
				Hour:    10,
				Minute:  0,
			},
			true,
			"days required",
		},
		{
			"invalid hour",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Action:  "lurk_missing",
				Days:    []string{"monday"},
				Hour:    25,
				Minute:  0,
			},
			true,
			"hour must be",
		},
		{
			"invalid minute",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Action:  "lurk_missing",
				Days:    []string{"monday"},
				Hour:    10,
				Minute:  61,
			},
			true,
			"minute must be",
		},
		{
			"invalid action",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Action:  "invalid_action",
				Days:    []string{"monday"},
				Hour:    10,
				Minute:  0,
			},
			true,
			"invalid action",
		},
		{
			"valid schedule - lurk_missing",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "sonarr",
				Action:  "lurk_missing",
				Days:    []string{"monday", "wednesday"},
				Hour:    10,
				Minute:  30,
				Enabled: true,
			},
			false,
			"",
		},
		{
			"valid schedule - clean_queue",
			&database.Schedule{
				ID:      uuid.New(),
				AppType: "radarr",
				Action:  "clean_queue",
				Days:    []string{"sunday"},
				Hour:    0,
				Minute:  0,
				Enabled: true,
			},
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchedule(tt.schedule)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSchedule() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && tt.errMsg != "" && (err == nil || !contains(err.Error(), tt.errMsg)) {
				t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

// NewBadRequest creates a bad request error
func NewBadRequest(msg string) error {
	return &validationError{message: msg}
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
