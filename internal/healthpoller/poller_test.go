package healthpoller

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

type mockStore struct {
	instances map[database.AppType][]database.AppInstance
	settings  *database.GeneralSettings
	settErr   error
	listErr   error
}

func (m *mockStore) ListEnabledInstances(_ context.Context, appType database.AppType) ([]database.AppInstance, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.instances[appType], nil
}

func (m *mockStore) GetGeneralSettings(_ context.Context) (*database.GeneralSettings, error) {
	if m.settErr != nil {
		return nil, m.settErr
	}
	return m.settings, nil
}

func TestNew(t *testing.T) {
	store := &mockStore{}
	p := New(store)
	if p == nil {
		t.Fatal("expected non-nil Poller")
	}
	if p.db != store {
		t.Fatal("expected Poller.db to be the given store")
	}
}

func TestStartStop(t *testing.T) {
	store := &mockStore{
		instances: make(map[database.AppType][]database.AppInstance),
		settings:  &database.GeneralSettings{APITimeout: 5},
	}
	p := New(store)
	ctx := context.Background()
	p.Start(ctx)
	// Should not panic.
	p.Stop()
}

func TestPollAll_NoInstances(t *testing.T) {
	store := &mockStore{
		instances: make(map[database.AppType][]database.AppInstance),
		settings:  &database.GeneralSettings{APITimeout: 5},
	}
	p := New(store)
	// Should not panic when there are no instances.
	p.pollAll(context.Background())
}

func TestPollAll_SettingsError(t *testing.T) {
	store := &mockStore{
		settErr: context.DeadlineExceeded,
	}
	p := New(store)
	// Should return gracefully on settings error.
	p.pollAll(context.Background())
}

func TestPollAll_ListError(t *testing.T) {
	store := &mockStore{
		settings: &database.GeneralSettings{APITimeout: 5},
		listErr:  context.DeadlineExceeded,
	}
	p := New(store)
	// Should handle list errors per app type gracefully.
	p.pollAll(context.Background())
}

func TestPollAll_UnreachableInstance(t *testing.T) {
	store := &mockStore{
		instances: map[database.AppType][]database.AppInstance{
			database.AppSonarr: {
				{
					ID:      uuid.New(),
					AppType: database.AppSonarr,
					Name:    "test-sonarr",
					APIURL:  "http://127.0.0.1:1", // unreachable
					APIKey:  "testkey",
					Enabled: true,
				},
			},
		},
		settings: &database.GeneralSettings{APITimeout: 1},
	}
	p := New(store)
	// Should handle unreachable instances without panicking.
	p.pollAll(context.Background())
}

func TestPollAll_DefaultTimeout(t *testing.T) {
	store := &mockStore{
		instances: make(map[database.AppType][]database.AppInstance),
		settings:  &database.GeneralSettings{APITimeout: 0},
	}
	p := New(store)
	// Should use default 30s timeout when APITimeout is 0.
	p.pollAll(context.Background())
}
