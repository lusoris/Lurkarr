package seerr

import (
	"context"
)

// DBSettingsFunc adapts a database GetSeerrSettings function to SettingsProvider.
type DBSettingsFunc func(ctx context.Context) (url, apiKey string, enabled bool, syncMinutes int, autoApprove bool, err error)

// GetSeerrSettings implements SettingsProvider.
func (f DBSettingsFunc) GetSeerrSettings(ctx context.Context) (*Settings, error) {
	url, apiKey, enabled, syncMinutes, autoApprove, err := f(ctx)
	if err != nil {
		return nil, err
	}
	return &Settings{
		URL:                 url,
		APIKey:              apiKey,
		Enabled:             enabled,
		SyncIntervalMinutes: syncMinutes,
		AutoApprove:         autoApprove,
	}, nil
}
