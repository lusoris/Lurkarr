//go:build integration

package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lusoris/lurkarr/internal/database"
)

// newTestDB spins up a PostgreSQL container and returns a connected database.DB.
func newTestDB(t *testing.T) *database.DB {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("lurkarr_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	db, err := database.New(ctx, connStr, 5)
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}

// =============================================================================
// User CRUD
// =============================================================================

func TestUserCRUD(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Create a user.
	user, err := db.CreateUser(ctx, "admin", "$2a$10$hashedpassword")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("username = %q, want %q", user.Username, "admin")
	}
	if user.ID == uuid.Nil {
		t.Error("user ID should not be nil")
	}
	if !user.IsAdmin {
		t.Error("new user should be admin by default (per migration 004)")
	}

	// Get by username.
	found, err := db.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("found.ID = %v, want %v", found.ID, user.ID)
	}

	// Get by ID.
	byID, err := db.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if byID.Username != "admin" {
		t.Errorf("byID.Username = %q, want %q", byID.Username, "admin")
	}

	// UpdateUsername.
	if err := db.UpdateUsername(ctx, user.ID, "superadmin"); err != nil {
		t.Fatalf("update username: %v", err)
	}
	updated, _ := db.GetUserByID(ctx, user.ID)
	if updated.Username != "superadmin" {
		t.Errorf("username = %q, want %q", updated.Username, "superadmin")
	}

	// UpdatePassword.
	if err := db.UpdatePassword(ctx, user.ID, "$2a$10$newhashedpassword"); err != nil {
		t.Fatalf("update password: %v", err)
	}
	updated, _ = db.GetUserByID(ctx, user.ID)
	if updated.Password != "$2a$10$newhashedpassword" {
		t.Error("password not updated")
	}

	// UpdateUserAdmin.
	if err := db.UpdateUserAdmin(ctx, user.ID, true); err != nil {
		t.Fatalf("update admin: %v", err)
	}
	updated, _ = db.GetUserByID(ctx, user.ID)
	if !updated.IsAdmin {
		t.Error("expected user to be admin")
	}

	// User count.
	count, err := db.UserCount(ctx)
	if err != nil {
		t.Fatalf("user count: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// ListUsers.
	users, err := db.ListUsers(ctx)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("len(users) = %d, want 1", len(users))
	}

	// Delete user.
	if err := db.DeleteUser(ctx, user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	count, _ = db.UserCount(ctx)
	if count != 0 {
		t.Errorf("count after delete = %d, want 0", count)
	}
}

// =============================================================================
// Session Management
// =============================================================================

func TestSessionManagement(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	user, err := db.CreateUser(ctx, "sessuser", "$2a$10$hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create session.
	sess, err := db.CreateSession(ctx, user.ID, 24*time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if sess.UserID != user.ID {
		t.Errorf("session user_id = %v, want %v", sess.UserID, user.ID)
	}
	if time.Until(sess.ExpiresAt) < 23*time.Hour {
		t.Error("session expires too soon")
	}

	// Create session with metadata.
	sess2, err := db.CreateSessionWithMeta(ctx, user.ID, time.Hour, "192.168.1.1", "TestBrowser/1.0")
	if err != nil {
		t.Fatalf("create session with meta: %v", err)
	}
	if sess2.IPAddress != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", sess2.IPAddress, "192.168.1.1")
	}
	if sess2.UserAgent != "TestBrowser/1.0" {
		t.Errorf("ua = %q, want %q", sess2.UserAgent, "TestBrowser/1.0")
	}

	// Get session.
	got, err := db.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("session id mismatch")
	}

	// List user sessions.
	sessions, err := db.ListUserSessions(ctx, user.ID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}

	// DeleteUserSessionsExcept.
	if err := db.DeleteUserSessionsExcept(ctx, user.ID, sess.ID); err != nil {
		t.Fatalf("delete sessions except: %v", err)
	}
	sessions, _ = db.ListUserSessions(ctx, user.ID)
	if len(sessions) != 1 {
		t.Errorf("len(sessions) = %d, want 1 after delete except", len(sessions))
	}
	if sessions[0].ID != sess.ID {
		t.Error("wrong session kept")
	}

	// Delete single session.
	if err := db.DeleteSession(ctx, sess.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	sessions, _ = db.ListUserSessions(ctx, user.ID)
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

// =============================================================================
// TOTP / Recovery Codes
// =============================================================================

func TestTOTPAndRecoveryCodes(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	user, _ := db.CreateUser(ctx, "totpuser", "$2a$10$hash")

	// Set TOTP secret.
	secret := "JBSWY3DPEHPK3PXP"
	if err := db.SetTOTPSecret(ctx, user.ID, &secret); err != nil {
		t.Fatalf("set totp: %v", err)
	}
	updated, _ := db.GetUserByID(ctx, user.ID)
	if updated.TOTPSecret == nil || *updated.TOTPSecret != secret {
		t.Error("totp secret not set correctly")
	}

	// Clear TOTP secret.
	if err := db.SetTOTPSecret(ctx, user.ID, nil); err != nil {
		t.Fatalf("clear totp: %v", err)
	}
	updated, _ = db.GetUserByID(ctx, user.ID)
	if updated.TOTPSecret != nil {
		t.Error("totp secret should be nil after clear")
	}

	// Set recovery codes.
	codes := []string{"code1", "code2", "code3"}
	if err := db.SetRecoveryCodes(ctx, user.ID, codes); err != nil {
		t.Fatalf("set recovery codes: %v", err)
	}
	updated, _ = db.GetUserByID(ctx, user.ID)
	if len(updated.RecoveryCodes) != 3 {
		t.Errorf("recovery codes = %d, want 3", len(updated.RecoveryCodes))
	}
}

// =============================================================================
// External Users (OIDC)
// =============================================================================

func TestGetOrCreateExternalUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First call: creates user.
	u1, err := db.GetOrCreateExternalUser(ctx, "oidc", "ext-123", "alice")
	if err != nil {
		t.Fatalf("create external user: %v", err)
	}
	if u1.Username != "alice" {
		t.Errorf("username = %q, want %q", u1.Username, "alice")
	}
	if u1.AuthProvider != "oidc" {
		t.Errorf("auth_provider = %q, want %q", u1.AuthProvider, "oidc")
	}
	if u1.ExternalID != "ext-123" {
		t.Errorf("external_id = %q, want %q", u1.ExternalID, "ext-123")
	}

	// Second call: returns existing user.
	u2, err := db.GetOrCreateExternalUser(ctx, "oidc", "ext-123", "alice")
	if err != nil {
		t.Fatalf("get external user: %v", err)
	}
	if u2.ID != u1.ID {
		t.Error("second call should return same user")
	}

	// Third call: same external ID but different username → updates username.
	u3, err := db.GetOrCreateExternalUser(ctx, "oidc", "ext-123", "alice-updated")
	if err != nil {
		t.Fatalf("update external user: %v", err)
	}
	if u3.ID != u1.ID {
		t.Error("should still be same user")
	}
	if u3.Username != "alice-updated" {
		t.Errorf("username = %q, want %q", u3.Username, "alice-updated")
	}
}

// =============================================================================
// App Instances
// =============================================================================

func TestAppInstanceCRUD(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Create instance.
	inst, err := db.CreateInstance(ctx, database.AppSonarr, "Sonarr Main",
		"http://sonarr:8989", "abc123key")
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}
	if inst.AppType != database.AppSonarr {
		t.Errorf("app_type = %q, want %q", inst.AppType, database.AppSonarr)
	}
	if inst.Name != "Sonarr Main" {
		t.Errorf("name = %q, want %q", inst.Name, "Sonarr Main")
	}
	if !inst.Enabled {
		t.Error("new instance should be enabled by default")
	}

	// Get instance.
	got, err := db.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if got.APIURL != "http://sonarr:8989" {
		t.Errorf("api_url = %q, want %q", got.APIURL, "http://sonarr:8989")
	}

	// List instances.
	list, err := db.ListInstances(ctx, database.AppSonarr)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("len = %d, want 1", len(list))
	}

	// Update instance.
	if err := db.UpdateInstance(ctx, inst.ID, "Sonarr Updated",
		"http://sonarr:9999", "newkey", false); err != nil {
		t.Fatalf("update instance: %v", err)
	}
	updated, _ := db.GetInstance(ctx, inst.ID)
	if updated.Name != "Sonarr Updated" || updated.Enabled {
		t.Error("update did not apply correctly")
	}

	// ListEnabledInstances should be empty now.
	enabled, err := db.ListEnabledInstances(ctx, database.AppSonarr)
	if err != nil {
		t.Fatalf("list enabled: %v", err)
	}
	if len(enabled) != 0 {
		t.Errorf("len(enabled) = %d, want 0", len(enabled))
	}

	// Create a second instance of a different type.
	_, err = db.CreateInstance(ctx, database.AppRadarr, "Radarr",
		"http://radarr:7878", "key456")
	if err != nil {
		t.Fatalf("create radarr: %v", err)
	}

	// List should only return sonarr instances.
	sonarrList, _ := db.ListInstances(ctx, database.AppSonarr)
	radarrList, _ := db.ListInstances(ctx, database.AppRadarr)
	if len(sonarrList) != 1 || len(radarrList) != 1 {
		t.Errorf("sonarr=%d radarr=%d, want 1,1", len(sonarrList), len(radarrList))
	}

	// Delete instance.
	if err := db.DeleteInstance(ctx, inst.ID); err != nil {
		t.Fatalf("delete instance: %v", err)
	}
	sonarrList, _ = db.ListInstances(ctx, database.AppSonarr)
	if len(sonarrList) != 0 {
		t.Errorf("len after delete = %d, want 0", len(sonarrList))
	}
}

// =============================================================================
// App Settings
// =============================================================================

func TestAppSettings(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Migration seeds default app_settings rows, so we should be able to read them.
	settings, err := db.GetAppSettings(ctx, database.AppSonarr)
	if err != nil {
		t.Fatalf("get app settings: %v", err)
	}
	if settings.AppType != database.AppSonarr {
		t.Errorf("app_type = %q, want %q", settings.AppType, database.AppSonarr)
	}

	// Update settings.
	settings.LurkMissingCount = 25
	settings.MonitoredOnly = true
	settings.SelectionMode = "newest"
	if err := db.UpdateAppSettings(ctx, settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	updated, _ := db.GetAppSettings(ctx, database.AppSonarr)
	if updated.LurkMissingCount != 25 {
		t.Errorf("missing_count = %d, want 25", updated.LurkMissingCount)
	}
	if !updated.MonitoredOnly {
		t.Error("monitored_only not updated")
	}
	if updated.SelectionMode != "newest" {
		t.Error("selection_mode not updated")
	}
}

// =============================================================================
// General Settings
// =============================================================================

func TestGeneralSettings(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Upsert (insert).
	s := &database.GeneralSettings{
		SecretKey:            "my-secret-key",
		ProxyAuthBypass:      false,
		SSLVerify:            true,
		APITimeout:           30,
		StatefulResetHours:   24,
		CommandWaitDelay:     5,
		CommandWaitAttempts:  3,
		MinDownloadQueueSize: 10,
	}
	if err := db.UpsertGeneralSettings(ctx, s); err != nil {
		t.Fatalf("upsert general settings: %v", err)
	}

	got, err := db.GetGeneralSettings(ctx)
	if err != nil {
		t.Fatalf("get general settings: %v", err)
	}
	if got.APITimeout != 30 {
		t.Errorf("api_timeout = %d, want 30", got.APITimeout)
	}
	if got.MinDownloadQueueSize != 10 {
		t.Errorf("min_download_queue_size = %d, want 10", got.MinDownloadQueueSize)
	}

	// Upsert (update).
	s.APITimeout = 60
	if err := db.UpsertGeneralSettings(ctx, s); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	got, _ = db.GetGeneralSettings(ctx)
	if got.APITimeout != 60 {
		t.Errorf("api_timeout = %d, want 60", got.APITimeout)
	}
}

// =============================================================================
// Health Check
// =============================================================================

func TestHealthCheck(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.HealthCheck(ctx); err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

// =============================================================================
// Processed Items
// =============================================================================

func TestProcessedItems(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Create an instance first (needed for FK).
	inst, err := db.CreateInstance(ctx, database.AppSonarr, "Sonarr",
		"http://sonarr:8989", "key")
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}

	// Not yet processed.
	processed, err := db.IsProcessed(ctx, database.AppSonarr, inst.ID, 42, "lurk_missing")
	if err != nil {
		t.Fatalf("is processed: %v", err)
	}
	if processed {
		t.Error("should not be processed yet")
	}

	// Mark as processed.
	if err := db.MarkProcessed(ctx, database.AppSonarr, inst.ID, 42, "lurk_missing"); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	// Now should be processed.
	processed, err = db.IsProcessed(ctx, database.AppSonarr, inst.ID, 42, "lurk_missing")
	if err != nil {
		t.Fatalf("is processed after mark: %v", err)
	}
	if !processed {
		t.Error("should be processed after mark")
	}

	// Different operation should not be processed.
	processed, _ = db.IsProcessed(ctx, database.AppSonarr, inst.ID, 42, "lurk_upgrade")
	if processed {
		t.Error("different operation should not be processed")
	}
}

// =============================================================================
// Lurk History
// =============================================================================

func TestLurkHistory(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	inst, _ := db.CreateInstance(ctx, database.AppSonarr, "Sonarr",
		"http://sonarr:8989", "key")

	// Record history entries.
	for i := range 5 {
		if err := db.AddLurkHistory(ctx, database.AppSonarr, inst.ID, "Sonarr",
			i+1, fmt.Sprintf("Show %d", i), "lurk_missing"); err != nil {
			t.Fatalf("record lurk %d: %v", i, err)
		}
	}

	// List history with pagination.
	q := database.HistoryQuery{
		AppType: string(database.AppSonarr),
		Limit:   3,
		Offset:  0,
	}
	entries, total, err := db.ListLurkHistory(ctx, q)
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}

	// Delete history for sonarr.
	if err := db.DeleteHistory(ctx, database.AppSonarr); err != nil {
		t.Fatalf("delete history: %v", err)
	}
	_, total, _ = db.ListLurkHistory(ctx, q)
	if total != 0 {
		t.Errorf("total after delete = %d, want 0", total)
	}
}

// =============================================================================
// Stats
// =============================================================================

func TestStats(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Stats are per-instance, not seeded — create an instance and increment.
	inst, _ := db.CreateInstance(ctx, database.AppSonarr, "Sonarr",
		"http://sonarr:8989", "key")

	if err := db.IncrementStats(ctx, database.AppSonarr, inst.ID, 10, 3); err != nil {
		t.Fatalf("increment stats: %v", err)
	}

	stats, err := db.GetAllStats(ctx)
	if err != nil {
		t.Fatalf("get all stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("len(stats) = %d, want 1", len(stats))
	}
	if stats[0].Lurked != 10 || stats[0].Upgraded != 3 {
		t.Errorf("stats = lurked:%d upgraded:%d, want 10,3", stats[0].Lurked, stats[0].Upgraded)
	}

	// Increment again — should accumulate.
	if err := db.IncrementStats(ctx, database.AppSonarr, inst.ID, 5, 2); err != nil {
		t.Fatalf("increment stats 2: %v", err)
	}
	stats, _ = db.GetAllStats(ctx)
	if stats[0].Lurked != 15 || stats[0].Upgraded != 5 {
		t.Errorf("accumulated stats = lurked:%d upgraded:%d, want 15,5", stats[0].Lurked, stats[0].Upgraded)
	}

	if err := db.ResetStats(ctx); err != nil {
		t.Fatalf("reset stats: %v", err)
	}
	stats, _ = db.GetAllStats(ctx)
	for _, s := range stats {
		if s.Lurked != 0 || s.Upgraded != 0 {
			t.Errorf("stats not reset for %s", s.AppType)
		}
	}
}

// =============================================================================
// State (Last Reset)
// =============================================================================

func TestState(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	inst, _ := db.CreateInstance(ctx, database.AppSonarr, "Sonarr",
		"http://sonarr:8989", "key")

	// No state yet.
	last, err := db.GetLastReset(ctx, database.AppSonarr, inst.ID)
	if err != nil {
		t.Fatalf("get last reset: %v", err)
	}
	if last != nil {
		t.Error("expected nil for no state")
	}

	// Reset state creates the entry.
	if err := db.ResetState(ctx, database.AppSonarr, inst.ID); err != nil {
		t.Fatalf("reset state: %v", err)
	}
	last, _ = db.GetLastReset(ctx, database.AppSonarr, inst.ID)
	if last == nil {
		t.Error("expected non-nil after reset")
	}
}

// =============================================================================
// Clean Expired Sessions
// =============================================================================

func TestCleanExpiredSessions(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	user, _ := db.CreateUser(ctx, "expuser", "$2a$10$hash")

	// Create an expired session (very short duration).
	_, err := db.CreateSession(ctx, user.ID, -time.Hour) // already expired
	if err != nil {
		t.Fatalf("create expired session: %v", err)
	}

	// Create an active session.
	active, err := db.CreateSession(ctx, user.ID, 24*time.Hour)
	if err != nil {
		t.Fatalf("create active session: %v", err)
	}

	// Clean expired.
	cleaned, err := db.CleanExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("clean expired: %v", err)
	}
	if cleaned != 1 {
		t.Errorf("cleaned = %d, want 1", cleaned)
	}

	// Active session still exists.
	got, err := db.GetSession(ctx, active.ID)
	if err != nil {
		t.Fatalf("get active session: %v", err)
	}
	if got.ID != active.ID {
		t.Error("active session should persist")
	}
}
