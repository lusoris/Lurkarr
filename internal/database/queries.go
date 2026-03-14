package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --- Users ---

func (db *DB) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO users (username, password) VALUES ($1, $2)
		 RETURNING id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.RecoveryCodes, &u.AuthProvider, &u.ExternalID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.RecoveryCodes, &u.AuthProvider, &u.ExternalID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.RecoveryCodes, &u.AuthProvider, &u.ExternalID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (db *DB) UpdateUsername(ctx context.Context, id uuid.UUID, username string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET username = $1, updated_at = now() WHERE id = $2`,
		username, id,
	)
	return err
}

func (db *DB) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET password = $1, updated_at = now() WHERE id = $2`,
		passwordHash, id,
	)
	return err
}

func (db *DB) SetTOTPSecret(ctx context.Context, id uuid.UUID, secret *string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET totp_secret = $1, updated_at = now() WHERE id = $2`,
		secret, id,
	)
	return err
}

func (db *DB) UpdateUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET is_admin = $1, updated_at = now() WHERE id = $2`,
		isAdmin, id,
	)
	return err
}

func (db *DB) UserCount(ctx context.Context) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// GetOrCreateExternalUser finds a user by auth_provider + external_id, or creates one.
func (db *DB) GetOrCreateExternalUser(ctx context.Context, provider, externalID, username string) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at
		 FROM users WHERE auth_provider = $1 AND external_id = $2`,
		provider, externalID,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.RecoveryCodes, &u.AuthProvider, &u.ExternalID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		// Update username if changed at the provider.
		if u.Username != username {
			_, _ = db.Pool.Exec(ctx,
				`UPDATE users SET username = $1, updated_at = now() WHERE id = $2`,
				username, u.ID,
			)
			u.Username = username
		}
		return &u, nil
	}

	// Create user — external users get a placeholder password (cannot login locally).
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO users (username, password, auth_provider, external_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at`,
		username, "!oidc-no-local-password", provider, externalID,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.RecoveryCodes, &u.AuthProvider, &u.ExternalID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create external user: %w", err)
	}
	return &u, nil
}

// --- Sessions ---

func (db *DB) CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*Session, error) {
	return db.CreateSessionWithMeta(ctx, userID, duration, "", "")
}

func (db *DB) CreateSessionWithMeta(ctx context.Context, userID uuid.UUID, duration time.Duration, ipAddress, userAgent string) (*Session, error) {
	var s Session
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO sessions (user_id, expires_at, ip_address, user_agent) VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, expires_at, created_at, ip_address, user_agent`,
		userID, time.Now().Add(duration), ipAddress, userAgent,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &s.IPAddress, &s.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &s, nil
}

func (db *DB) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	var s Session
	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at, created_at, ip_address, user_agent FROM sessions WHERE id = $1 AND expires_at > now()`,
		id,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &s.IPAddress, &s.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (db *DB) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (db *DB) CleanExpiredSessions(ctx context.Context) (int64, error) {
	ct, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ListUserSessions returns all active sessions for a user.
func (db *DB) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, user_id, expires_at, created_at, ip_address, user_agent
		 FROM sessions WHERE user_id = $1 AND expires_at > now() ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Session])
}

// DeleteUserSessions deletes all sessions for a user.
func (db *DB) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// DeleteUserSessionsExcept deletes all sessions for a user except the given session.
func (db *DB) DeleteUserSessionsExcept(ctx context.Context, userID, keep uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1 AND id != $2`, userID, keep)
	return err
}

// --- User Admin ---

// ListUsers returns all users.
func (db *DB) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, username, password, totp_secret, recovery_codes, auth_provider, external_id, is_admin, created_at, updated_at
		 FROM users ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[User])
}

// DeleteUser deletes a user by ID.
func (db *DB) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// SetRecoveryCodes stores hashed recovery codes for a user.
func (db *DB) SetRecoveryCodes(ctx context.Context, id uuid.UUID, codes []string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET recovery_codes = $1, updated_at = now() WHERE id = $2`,
		codes, id,
	)
	return err
}

// --- App Instances ---

func (db *DB) ListInstances(ctx context.Context, appType AppType) ([]AppInstance, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, name, api_url, api_key, enabled, created_at
		 FROM app_instances WHERE app_type = $1 ORDER BY name`,
		string(appType),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[AppInstance])
}

func (db *DB) GetInstance(ctx context.Context, id uuid.UUID) (*AppInstance, error) {
	var i AppInstance
	err := db.Pool.QueryRow(ctx,
		`SELECT id, app_type, name, api_url, api_key, enabled, created_at
		 FROM app_instances WHERE id = $1`,
		id,
	).Scan(&i.ID, &i.AppType, &i.Name, &i.APIURL, &i.APIKey, &i.Enabled, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (db *DB) CreateInstance(ctx context.Context, appType AppType, name, apiURL, apiKey string) (*AppInstance, error) {
	var i AppInstance
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO app_instances (app_type, name, api_url, api_key)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, app_type, name, api_url, api_key, enabled, created_at`,
		string(appType), name, apiURL, apiKey,
	).Scan(&i.ID, &i.AppType, &i.Name, &i.APIURL, &i.APIKey, &i.Enabled, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (db *DB) UpdateInstance(ctx context.Context, id uuid.UUID, name, apiURL, apiKey string, enabled bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE app_instances SET name = $1, api_url = $2, api_key = $3, enabled = $4 WHERE id = $5`,
		name, apiURL, apiKey, enabled, id,
	)
	return err
}

func (db *DB) DeleteInstance(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM app_instances WHERE id = $1`, id)
	return err
}

func (db *DB) ListEnabledInstances(ctx context.Context, appType AppType) ([]AppInstance, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, name, api_url, api_key, enabled, created_at
		 FROM app_instances WHERE app_type = $1 AND enabled = true ORDER BY name`,
		string(appType),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[AppInstance])
}

// --- App Settings ---

func (db *DB) GetAppSettings(ctx context.Context, appType AppType) (*AppSettings, error) {
	var s AppSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT app_type, lurk_missing_count, lurk_upgrade_count, lurk_missing_mode,
		        upgrade_mode, sleep_duration, monitored_only, skip_future,
		        hourly_cap, selection_mode, max_search_failures, debug_mode
		 FROM app_settings WHERE app_type = $1`,
		string(appType),
	).Scan(&s.AppType, &s.LurkMissingCount, &s.LurkUpgradeCount, &s.LurkMissingMode,
		&s.UpgradeMode, &s.SleepDuration, &s.MonitoredOnly, &s.SkipFuture,
		&s.HourlyCap, &s.SelectionMode, &s.MaxSearchFailures, &s.DebugMode)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) UpdateAppSettings(ctx context.Context, s *AppSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE app_settings SET
		    lurk_missing_count = $1, lurk_upgrade_count = $2, lurk_missing_mode = $3,
		    upgrade_mode = $4, sleep_duration = $5, monitored_only = $6, skip_future = $7,
		    hourly_cap = $8, selection_mode = $9, max_search_failures = $10, debug_mode = $11
		 WHERE app_type = $12`,
		s.LurkMissingCount, s.LurkUpgradeCount, s.LurkMissingMode,
		s.UpgradeMode, s.SleepDuration, s.MonitoredOnly, s.SkipFuture,
		s.HourlyCap, s.SelectionMode, s.MaxSearchFailures, s.DebugMode, string(s.AppType),
	)
	return err
}

// --- General Settings ---

func (db *DB) GetGeneralSettings(ctx context.Context) (*GeneralSettings, error) {
	var s GeneralSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT secret_key, proxy_auth_bypass, ssl_verify, api_timeout,
		        stateful_reset_hours, command_wait_delay, command_wait_attempts,
		        min_download_queue_size
		 FROM general_settings WHERE id = 1`,
	).Scan(&s.SecretKey, &s.ProxyAuthBypass, &s.SSLVerify, &s.APITimeout,
		&s.StatefulResetHours, &s.CommandWaitDelay, &s.CommandWaitAttempts,
		&s.MinDownloadQueueSize)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) UpsertGeneralSettings(ctx context.Context, s *GeneralSettings) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO general_settings (id, secret_key, proxy_auth_bypass, ssl_verify, api_timeout,
		    stateful_reset_hours, command_wait_delay, command_wait_attempts, min_download_queue_size)
		 VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (id) DO UPDATE SET
		    secret_key = EXCLUDED.secret_key,
		    proxy_auth_bypass = EXCLUDED.proxy_auth_bypass,
		    ssl_verify = EXCLUDED.ssl_verify,
		    api_timeout = EXCLUDED.api_timeout,
		    stateful_reset_hours = EXCLUDED.stateful_reset_hours,
		    command_wait_delay = EXCLUDED.command_wait_delay,
		    command_wait_attempts = EXCLUDED.command_wait_attempts,
		    min_download_queue_size = EXCLUDED.min_download_queue_size`,
		s.SecretKey, s.ProxyAuthBypass, s.SSLVerify, s.APITimeout,
		s.StatefulResetHours, s.CommandWaitDelay, s.CommandWaitAttempts,
		s.MinDownloadQueueSize,
	)
	return err
}

// --- Processed Items ---

func (db *DB) IsProcessed(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID int, operation string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_items
		 WHERE app_type = $1 AND instance_id = $2 AND media_id = $3 AND operation = $4)`,
		string(appType), instanceID, mediaID, operation,
	).Scan(&exists)
	return exists, err
}

func (db *DB) MarkProcessed(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID int, operation string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO processed_items (app_type, instance_id, media_id, operation)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (app_type, instance_id, media_id, operation) DO UPDATE SET processed_at = now()`,
		string(appType), instanceID, mediaID, operation,
	)
	return err
}

// GetProcessedTimes returns the last-processed timestamp for each media ID.
// Used by the "least_recent" selection mode to prioritize items that haven't
// been searched recently. Items not in the map have never been processed.
func (db *DB) GetProcessedTimes(ctx context.Context, appType AppType, instanceID uuid.UUID, operation string) (map[int]time.Time, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT media_id, processed_at FROM processed_items
		 WHERE app_type = $1 AND instance_id = $2 AND operation = $3`,
		string(appType), instanceID, operation,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int]time.Time)
	for rows.Next() {
		var id int
		var t time.Time
		if err := rows.Scan(&id, &t); err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, rows.Err()
}

// --- Search Failure Tracking ---

// RecordSearchFailure increments the failure count for a media item, or inserts
// a new record with count=1 if none exists.
func (db *DB) RecordSearchFailure(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID int) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO search_failures (app_type, instance_id, media_id, fail_count, last_failed)
		 VALUES ($1, $2, $3, 1, NOW())
		 ON CONFLICT (app_type, instance_id, media_id)
		 DO UPDATE SET fail_count = search_failures.fail_count + 1, last_failed = NOW()`,
		string(appType), instanceID, mediaID,
	)
	return err
}

// ClearSearchFailure removes the failure record for a media item on successful search.
func (db *DB) ClearSearchFailure(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID int) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM search_failures WHERE app_type = $1 AND instance_id = $2 AND media_id = $3`,
		string(appType), instanceID, mediaID,
	)
	return err
}

// GetSearchFailureCounts returns a map of mediaID → fail_count for all tracked
// failures on the given instance. Used by the lurking engine to batch-filter
// items that have exceeded the failure limit.
func (db *DB) GetSearchFailureCounts(ctx context.Context, appType AppType, instanceID uuid.UUID) (map[int]int, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT media_id, fail_count FROM search_failures
		 WHERE app_type = $1 AND instance_id = $2`,
		string(appType), instanceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int]int)
	for rows.Next() {
		var id, count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

// IsSearchFailureLimitReached checks whether a single media item has reached
// the maximum number of consecutive search failures.
func (db *DB) IsSearchFailureLimitReached(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID, maxFailures int) (bool, error) {
	var count int
	err := db.Pool.QueryRow(ctx,
		`SELECT fail_count FROM search_failures
		 WHERE app_type = $1 AND instance_id = $2 AND media_id = $3`,
		string(appType), instanceID, mediaID,
	).Scan(&count)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return count >= maxFailures, nil
}

// PruneSearchFailures deletes failure records older than the given duration.
func (db *DB) PruneSearchFailures(ctx context.Context, olderThan time.Duration) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM search_failures WHERE last_failed < NOW() - $1::interval`,
		olderThan.String(),
	)
	return err
}

func (db *DB) ResetState(ctx context.Context, appType AppType, instanceID uuid.UUID) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`DELETE FROM processed_items WHERE app_type = $1 AND instance_id = $2`,
		string(appType), instanceID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO state_resets (app_type, instance_id, last_reset)
		 VALUES ($1, $2, now())
		 ON CONFLICT (app_type, instance_id) DO UPDATE SET last_reset = now()`,
		string(appType), instanceID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) GetLastReset(ctx context.Context, appType AppType, instanceID uuid.UUID) (*time.Time, error) {
	var t time.Time
	err := db.Pool.QueryRow(ctx,
		`SELECT last_reset FROM state_resets WHERE app_type = $1 AND instance_id = $2`,
		string(appType), instanceID,
	).Scan(&t)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// --- Lurk History ---

func (db *DB) AddLurkHistory(ctx context.Context, appType AppType, instanceID uuid.UUID, instanceName string, mediaID int, mediaTitle, operation string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO lurk_history (app_type, instance_id, instance_name, media_id, media_title, operation)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		string(appType), instanceID, instanceName, mediaID, mediaTitle, operation,
	)
	return err
}

type HistoryQuery struct {
	AppType string
	Search  string
	Limit   int
	Offset  int
}

func (db *DB) ListLurkHistory(ctx context.Context, q HistoryQuery) ([]LurkHistory, int, error) {
	args := []any{}
	where := "WHERE 1=1"
	argN := 1

	if q.AppType != "" {
		where += fmt.Sprintf(" AND app_type = $%d", argN)
		args = append(args, q.AppType)
		argN++
	}
	if q.Search != "" {
		where += fmt.Sprintf(" AND to_tsvector('english', media_title) @@ plainto_tsquery('english', $%d)", argN)
		args = append(args, q.Search)
		argN++
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM lurk_history "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT id, app_type, instance_id, instance_name, media_id, media_title, operation, created_at
		 FROM lurk_history %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argN, argN+1,
	)
	args = append(args, q.Limit, q.Offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []LurkHistory
	for rows.Next() {
		var h LurkHistory
		if err := rows.Scan(&h.ID, &h.AppType, &h.InstanceID, &h.InstanceName, &h.MediaID, &h.MediaTitle, &h.Operation, &h.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, h)
	}

	return results, total, rows.Err()
}

func (db *DB) DeleteHistory(ctx context.Context, appType AppType) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM lurk_history WHERE app_type = $1`, string(appType))
	return err
}

// --- Lurk Stats ---

func (db *DB) GetAllStats(ctx context.Context) ([]LurkStats, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT app_type, instance_id, lurked, upgraded, updated_at FROM lurk_stats ORDER BY app_type, instance_id`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[LurkStats])
}

func (db *DB) IncrementStats(ctx context.Context, appType AppType, instanceID uuid.UUID, lurked, upgraded int64) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO lurk_stats (app_type, instance_id, lurked, upgraded, updated_at)
		 VALUES ($1, $2, $3, $4, now())
		 ON CONFLICT (app_type, instance_id) DO UPDATE SET
		   lurked = lurk_stats.lurked + $3,
		   upgraded = lurk_stats.upgraded + $4,
		   updated_at = now()`,
		string(appType), instanceID, lurked, upgraded,
	)
	return err
}

func (db *DB) ResetStats(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE lurk_stats SET lurked = 0, upgraded = 0, updated_at = now()`,
	)
	return err
}

// --- Hourly Caps ---

func (db *DB) GetCurrentHourHits(ctx context.Context, appType AppType, instanceID uuid.UUID) (int, error) {
	var hits int
	err := db.Pool.QueryRow(ctx,
		`SELECT COALESCE(api_hits, 0) FROM hourly_caps
		 WHERE app_type = $1 AND instance_id = $2 AND hour_bucket = date_trunc('hour', now())`,
		string(appType), instanceID,
	).Scan(&hits)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return hits, err
}

func (db *DB) IncrementHourlyHits(ctx context.Context, appType AppType, instanceID uuid.UUID, count int) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO hourly_caps (app_type, instance_id, hour_bucket, api_hits)
		 VALUES ($1, $2, date_trunc('hour', now()), $3)
		 ON CONFLICT (app_type, instance_id, hour_bucket)
		 DO UPDATE SET api_hits = hourly_caps.api_hits + $3`,
		string(appType), instanceID, count,
	)
	return err
}

func (db *DB) GetAllHourlyCaps(ctx context.Context) ([]HourlyCap, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT app_type, instance_id, hour_bucket, api_hits FROM hourly_caps
		 WHERE hour_bucket = date_trunc('hour', now()) ORDER BY app_type, instance_id`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[HourlyCap])
}

// CleanupOldHourlyCaps removes hourly_caps entries older than 7 days.
func (db *DB) CleanupOldHourlyCaps(ctx context.Context) (int64, error) {
	tag, err := db.Pool.Exec(ctx,
		`DELETE FROM hourly_caps WHERE hour_bucket < now() - interval '7 days'`,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// --- Schedules ---

func (db *DB) ListSchedules(ctx context.Context) ([]Schedule, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, action, days, hour, minute, enabled, created_at
		 FROM schedules ORDER BY hour, minute`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Schedule
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.AppType, &s.Action, &s.Days, &s.Hour, &s.Minute, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (db *DB) CreateSchedule(ctx context.Context, s *Schedule) error {
	return db.Pool.QueryRow(ctx,
		`INSERT INTO schedules (app_type, action, days, hour, minute, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		s.AppType, s.Action, s.Days, s.Hour, s.Minute, s.Enabled,
	).Scan(&s.ID, &s.CreatedAt)
}

func (db *DB) UpdateSchedule(ctx context.Context, s *Schedule) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE schedules SET app_type = $1, action = $2, days = $3, hour = $4, minute = $5, enabled = $6
		 WHERE id = $7`,
		s.AppType, s.Action, s.Days, s.Hour, s.Minute, s.Enabled, s.ID,
	)
	return err
}

func (db *DB) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return err
}

func (db *DB) AddScheduleExecution(ctx context.Context, scheduleID uuid.UUID, result string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO schedule_executions (schedule_id, result) VALUES ($1, $2)`,
		scheduleID, result,
	)
	return err
}

func (db *DB) ListScheduleExecutions(ctx context.Context, limit int) ([]ScheduleExecution, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, schedule_id, executed_at, result FROM schedule_executions
		 ORDER BY executed_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ScheduleExecution
	for rows.Next() {
		var e ScheduleExecution
		if err := rows.Scan(&e.ID, &e.ScheduleID, &e.ExecutedAt, &e.Result); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// --- Logs ---

func (db *DB) InsertLogs(ctx context.Context, entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, e := range entries {
		batch.Queue(
			`INSERT INTO logs (app_type, level, message) VALUES ($1, $2, $3)`,
			e.AppType, e.Level, e.Message,
		)
	}

	br := db.Pool.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	for range entries {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type LogQuery struct {
	AppType string
	Level   string
	Limit   int
	Before  int64 // cursor: log ID
}

func (db *DB) QueryLogs(ctx context.Context, q LogQuery) ([]LogEntry, error) {
	args := []any{}
	where := "WHERE 1=1"
	argN := 1

	if q.AppType != "" {
		where += fmt.Sprintf(" AND app_type = $%d", argN)
		args = append(args, q.AppType)
		argN++
	}
	if q.Level != "" {
		where += fmt.Sprintf(" AND level = $%d", argN)
		args = append(args, q.Level)
		argN++
	}
	if q.Before > 0 {
		where += fmt.Sprintf(" AND id < $%d", argN)
		args = append(args, q.Before)
		argN++
	}

	limit := q.Limit
	if limit <= 0 || limit > 500 {
		limit = 500
	}

	query := fmt.Sprintf(
		`SELECT id, app_type, level, message, created_at FROM logs %s ORDER BY id DESC LIMIT $%d`,
		where, argN,
	)
	args = append(args, limit)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.ID, &e.AppType, &e.Level, &e.Message, &e.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

func (db *DB) PruneLogs(ctx context.Context, retentionDays int) (int64, error) {
	ct, err := db.Pool.Exec(ctx,
		`DELETE FROM logs WHERE created_at < now() - make_interval(days => $1)`,
		retentionDays,
	)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// --- WebAuthn Credentials ---

func (db *DB) CreateWebAuthnCredential(ctx context.Context, c *WebAuthnCredential) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO webauthn_credentials (user_id, name, credential_id, public_key, attestation_type, transport, aaguid, sign_count)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.UserID, c.Name, c.CredentialID, c.PublicKey, c.AttestationType, c.Transport, c.AAGUID, c.SignCount,
	)
	return err
}

func (db *DB) ListWebAuthnCredentials(ctx context.Context, userID uuid.UUID) ([]WebAuthnCredential, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, user_id, name, credential_id, public_key, attestation_type, transport, aaguid, sign_count, created_at
		 FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []WebAuthnCredential
	for rows.Next() {
		var c WebAuthnCredential
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CredentialID, &c.PublicKey,
			&c.AttestationType, &c.Transport, &c.AAGUID, &c.SignCount, &c.CreatedAt); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

func (db *DB) GetWebAuthnCredentialByID(ctx context.Context, credID []byte) (*WebAuthnCredential, error) {
	var c WebAuthnCredential
	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, name, credential_id, public_key, attestation_type, transport, aaguid, sign_count, created_at
		 FROM webauthn_credentials WHERE credential_id = $1`, credID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.CredentialID, &c.PublicKey,
		&c.AttestationType, &c.Transport, &c.AAGUID, &c.SignCount, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) DeleteWebAuthnCredential(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM webauthn_credentials WHERE id = $1`, id)
	return err
}

func (db *DB) UpdateWebAuthnSignCount(ctx context.Context, credentialID []byte, signCount int64) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE webauthn_credentials SET sign_count = $1 WHERE credential_id = $2`,
		signCount, credentialID)
	return err
}

func (db *DB) RenameWebAuthnCredential(ctx context.Context, id uuid.UUID, name string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE webauthn_credentials SET name = $1 WHERE id = $2`,
		name, id)
	return err
}

// --- Cleanup ---

func (db *DB) AutoResetExpiredStates(ctx context.Context, resetHours int) (int64, error) {
	ct, err := db.Pool.Exec(ctx,
		`DELETE FROM processed_items WHERE (app_type, instance_id) IN (
			SELECT sr.app_type, sr.instance_id FROM state_resets sr
			WHERE sr.last_reset < now() - make_interval(hours => $1)
		)`,
		resetHours,
	)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// --- CSRF Key ---

// GetCSRFKey returns the persisted CSRF key, or empty string if not set.
func (db *DB) GetCSRFKey(ctx context.Context) (string, error) {
	var key string
	err := db.Pool.QueryRow(ctx,
		`SELECT csrf_key FROM general_settings WHERE id = 1`,
	).Scan(&key)
	if err != nil {
		return "", err
	}
	return key, nil
}

// SetCSRFKey persists a CSRF key in general_settings.
func (db *DB) SetCSRFKey(ctx context.Context, key string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE general_settings SET csrf_key = $1 WHERE id = 1`,
		key,
	)
	return err
}

// --- Instance Groups ---

// CreateInstanceGroup creates a new instance group for the given app type.
func (db *DB) CreateInstanceGroup(ctx context.Context, appType AppType, name string) (*InstanceGroup, error) {
	var g InstanceGroup
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO instance_groups (app_type, name) VALUES ($1, $2)
		 RETURNING id, app_type, name, mode, created_at`,
		appType, name,
	).Scan(&g.ID, &g.AppType, &g.Name, &g.Mode, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create instance group: %w", err)
	}
	return &g, nil
}

// GetInstanceGroup returns a single instance group by ID with its members.
func (db *DB) GetInstanceGroup(ctx context.Context, id uuid.UUID) (*InstanceGroup, error) {
	var g InstanceGroup
	err := db.Pool.QueryRow(ctx,
		`SELECT id, app_type, name, mode, created_at FROM instance_groups WHERE id = $1`,
		id,
	).Scan(&g.ID, &g.AppType, &g.Name, &g.Mode, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get instance group: %w", err)
	}

	rows, err := db.Pool.Query(ctx,
		`SELECT m.group_id, m.instance_id, i.name, m.quality_rank, m.is_independent
		 FROM instance_group_members m
		 JOIN app_instances i ON i.id = m.instance_id
		 WHERE m.group_id = $1
		 ORDER BY m.quality_rank`, id)
	if err != nil {
		return nil, fmt.Errorf("get instance group members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m InstanceGroupMember
		if err := rows.Scan(&m.GroupID, &m.InstanceID, &m.InstanceName, &m.QualityRank, &m.IsIndependent); err != nil {
			return nil, fmt.Errorf("scan instance group member: %w", err)
		}
		g.Members = append(g.Members, m)
	}
	return &g, nil
}

// ListInstanceGroups returns all groups for a given app type with members.
func (db *DB) ListInstanceGroups(ctx context.Context, appType AppType) ([]InstanceGroup, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, name, mode, created_at FROM instance_groups
		 WHERE app_type = $1 ORDER BY name`, appType)
	if err != nil {
		return nil, fmt.Errorf("list instance groups: %w", err)
	}
	defer rows.Close()

	var groups []InstanceGroup
	for rows.Next() {
		var g InstanceGroup
		if err := rows.Scan(&g.ID, &g.AppType, &g.Name, &g.Mode, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan instance group: %w", err)
		}
		groups = append(groups, g)
	}

	// Load members for each group in a single query
	if len(groups) > 0 {
		groupIDs := make([]uuid.UUID, len(groups))
		groupMap := make(map[uuid.UUID]*InstanceGroup, len(groups))
		for i := range groups {
			groupIDs[i] = groups[i].ID
			groupMap[groups[i].ID] = &groups[i]
		}

		mRows, err := db.Pool.Query(ctx,
			`SELECT m.group_id, m.instance_id, i.name, m.quality_rank, m.is_independent
			 FROM instance_group_members m
			 JOIN app_instances i ON i.id = m.instance_id
			 WHERE m.group_id = ANY($1)
			 ORDER BY m.quality_rank`, groupIDs)
		if err != nil {
			return nil, fmt.Errorf("list instance group members: %w", err)
		}
		defer mRows.Close()

		for mRows.Next() {
			var m InstanceGroupMember
			if err := mRows.Scan(&m.GroupID, &m.InstanceID, &m.InstanceName, &m.QualityRank, &m.IsIndependent); err != nil {
				return nil, fmt.Errorf("scan instance group member: %w", err)
			}
			if g, ok := groupMap[m.GroupID]; ok {
				g.Members = append(g.Members, m)
			}
		}
	}
	return groups, nil
}

// UpdateInstanceGroup renames an instance group.
func (db *DB) UpdateInstanceGroup(ctx context.Context, id uuid.UUID, name string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE instance_groups SET name = $1 WHERE id = $2`,
		name, id)
	if err != nil {
		return fmt.Errorf("update instance group: %w", err)
	}
	return nil
}

// DeleteInstanceGroup deletes a group and its member associations.
func (db *DB) DeleteInstanceGroup(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM instance_groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete instance group: %w", err)
	}
	return nil
}

// SetGroupMembers replaces all members of a group with the provided list.
// Each member specifies an instance ID and quality rank.
func (db *DB) SetGroupMembers(ctx context.Context, groupID uuid.UUID, members []InstanceGroupMember) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	_, err = tx.Exec(ctx, `DELETE FROM instance_group_members WHERE group_id = $1`, groupID)
	if err != nil {
		return fmt.Errorf("clear group members: %w", err)
	}

	for _, m := range members {
		_, err = tx.Exec(ctx,
			`INSERT INTO instance_group_members (group_id, instance_id, quality_rank, is_independent)
			 VALUES ($1, $2, $3, $4)`,
			groupID, m.InstanceID, m.QualityRank, m.IsIndependent)
		if err != nil {
			return fmt.Errorf("insert group member: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetInstanceGroupForInstance returns the group an instance belongs to, or nil.
func (db *DB) GetInstanceGroupForInstance(ctx context.Context, instanceID uuid.UUID) (*InstanceGroup, error) {
	var groupID uuid.UUID
	err := db.Pool.QueryRow(ctx,
		`SELECT group_id FROM instance_group_members WHERE instance_id = $1`,
		instanceID,
	).Scan(&groupID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get group for instance: %w", err)
	}
	return db.GetInstanceGroup(ctx, groupID)
}

// UpdateInstanceGroupMode updates the mode of an instance group.
func (db *DB) UpdateInstanceGroupMode(ctx context.Context, id uuid.UUID, mode string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE instance_groups SET mode = $1 WHERE id = $2`, mode, id)
	if err != nil {
		return fmt.Errorf("update instance group mode: %w", err)
	}
	return nil
}

// --- Cross-Instance Media ---

// UpsertCrossInstanceMedia stores a detected media overlap, returning the record.
func (db *DB) UpsertCrossInstanceMedia(ctx context.Context, groupID uuid.UUID, externalID, title string) (*CrossInstanceMedia, error) {
	var m CrossInstanceMedia
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO cross_instance_media (group_id, external_id, title, detected_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (group_id, external_id)
		 DO UPDATE SET title = EXCLUDED.title, detected_at = now()
		 RETURNING id, group_id, external_id, title, detected_at`,
		groupID, externalID, title,
	).Scan(&m.ID, &m.GroupID, &m.ExternalID, &m.Title, &m.DetectedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert cross instance media: %w", err)
	}
	return &m, nil
}

// SetCrossInstancePresence replaces presence records for a media item.
func (db *DB) SetCrossInstancePresence(ctx context.Context, mediaID uuid.UUID, presence []CrossInstancePresence) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	_, err = tx.Exec(ctx, `DELETE FROM cross_instance_presence WHERE media_id = $1`, mediaID)
	if err != nil {
		return fmt.Errorf("clear presence: %w", err)
	}

	for _, p := range presence {
		_, err = tx.Exec(ctx,
			`INSERT INTO cross_instance_presence (media_id, instance_id, monitored, has_file)
			 VALUES ($1, $2, $3, $4)`,
			mediaID, p.InstanceID, p.Monitored, p.HasFile)
		if err != nil {
			return fmt.Errorf("insert presence: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ListCrossInstanceMedia returns all detected overlaps for a group.
func (db *DB) ListCrossInstanceMedia(ctx context.Context, groupID uuid.UUID) ([]CrossInstanceMedia, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, group_id, external_id, title, detected_at
		 FROM cross_instance_media WHERE group_id = $1
		 ORDER BY detected_at DESC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list cross instance media: %w", err)
	}
	defer rows.Close()

	var items []CrossInstanceMedia
	for rows.Next() {
		var m CrossInstanceMedia
		if err := rows.Scan(&m.ID, &m.GroupID, &m.ExternalID, &m.Title, &m.DetectedAt); err != nil {
			return nil, fmt.Errorf("scan cross instance media: %w", err)
		}
		items = append(items, m)
	}

	// Load presence for all items
	if len(items) > 0 {
		mediaIDs := make([]uuid.UUID, len(items))
		mediaMap := make(map[uuid.UUID]*CrossInstanceMedia, len(items))
		for i := range items {
			mediaIDs[i] = items[i].ID
			mediaMap[items[i].ID] = &items[i]
		}

		pRows, err := db.Pool.Query(ctx,
			`SELECT p.media_id, p.instance_id, i.name, p.monitored, p.has_file
			 FROM cross_instance_presence p
			 JOIN app_instances i ON i.id = p.instance_id
			 WHERE p.media_id = ANY($1)`, mediaIDs)
		if err != nil {
			return nil, fmt.Errorf("list cross instance presence: %w", err)
		}
		defer pRows.Close()

		for pRows.Next() {
			var p CrossInstancePresence
			if err := pRows.Scan(&p.MediaID, &p.InstanceID, &p.InstanceName, &p.Monitored, &p.HasFile); err != nil {
				return nil, fmt.Errorf("scan presence: %w", err)
			}
			if m, ok := mediaMap[p.MediaID]; ok {
				m.Presence = append(m.Presence, p)
			}
		}
	}
	return items, nil
}

// DeleteCrossInstanceMediaByGroup removes all overlap data for a group.
func (db *DB) DeleteCrossInstanceMediaByGroup(ctx context.Context, groupID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM cross_instance_media WHERE group_id = $1`, groupID)
	if err != nil {
		return fmt.Errorf("delete cross instance media: %w", err)
	}
	return nil
}

// --- Cross-Instance Routing ---

// MediaPresenceResult contains a matched media item with its group context.
type MediaPresenceResult struct {
	GroupID    uuid.UUID
	GroupMode  string
	ExternalID string
	Title      string
	Instances  []PresenceInstance
}

// PresenceInstance contains instance-level presence details with quality rank.
type PresenceInstance struct {
	InstanceID  uuid.UUID
	Name        string
	QualityRank int
	Monitored   bool
	HasFile     bool
}

// FindMediaPresenceByExternalID looks up cross-instance media presence by external ID,
// including group mode and member quality ranks for routing decisions.
func (db *DB) FindMediaPresenceByExternalID(ctx context.Context, externalID string) ([]MediaPresenceResult, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT
			cim.group_id, ig.mode, cim.external_id, cim.title,
			cip.instance_id, COALESCE(ai.name, ''), COALESCE(igm.quality_rank, 0),
			cip.monitored, cip.has_file
		FROM cross_instance_media cim
		JOIN instance_groups ig ON ig.id = cim.group_id
		JOIN cross_instance_presence cip ON cip.media_id = cim.id
		LEFT JOIN app_instances ai ON ai.id = cip.instance_id
		LEFT JOIN instance_group_members igm ON igm.group_id = cim.group_id AND igm.instance_id = cip.instance_id
		WHERE cim.external_id = $1
		ORDER BY cim.group_id, COALESCE(igm.quality_rank, 999)
	`, externalID)
	if err != nil {
		return nil, fmt.Errorf("find media presence: %w", err)
	}
	defer rows.Close()

	groupMap := make(map[uuid.UUID]*MediaPresenceResult)
	var order []uuid.UUID

	for rows.Next() {
		var groupID uuid.UUID
		var mode, extID, title, instName string
		var instID uuid.UUID
		var rank int
		var monitored, hasFile bool

		if err := rows.Scan(&groupID, &mode, &extID, &title, &instID, &instName, &rank, &monitored, &hasFile); err != nil {
			return nil, fmt.Errorf("scan media presence: %w", err)
		}

		result, ok := groupMap[groupID]
		if !ok {
			result = &MediaPresenceResult{
				GroupID:    groupID,
				GroupMode:  mode,
				ExternalID: extID,
				Title:      title,
			}
			groupMap[groupID] = result
			order = append(order, groupID)
		}
		result.Instances = append(result.Instances, PresenceInstance{
			InstanceID:  instID,
			Name:        instName,
			QualityRank: rank,
			Monitored:   monitored,
			HasFile:     hasFile,
		})
	}

	results := make([]MediaPresenceResult, 0, len(order))
	for _, id := range order {
		results = append(results, *groupMap[id])
	}
	return results, nil
}

// CreateCrossInstanceAction logs a routing or dedup action.
func (db *DB) CreateCrossInstanceAction(ctx context.Context, action CrossInstanceAction) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO cross_instance_actions (group_id, external_id, title, action, reason, seerr_request_id, source_instance_id, target_instance_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		action.GroupID, action.ExternalID, action.Title, action.Action, action.Reason,
		action.SeerrRequestID, action.SourceInstanceID, action.TargetInstanceID,
	)
	if err != nil {
		return fmt.Errorf("create cross instance action: %w", err)
	}
	return nil
}

// ListCrossInstanceActions returns recent routing actions, newest first.
func (db *DB) ListCrossInstanceActions(ctx context.Context, limit int) ([]CrossInstanceAction, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Pool.Query(ctx, `
		SELECT id, group_id, external_id, title, action, reason, seerr_request_id, source_instance_id, target_instance_id, executed_at
		FROM cross_instance_actions
		ORDER BY executed_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list cross instance actions: %w", err)
	}
	defer rows.Close()

	var actions []CrossInstanceAction
	for rows.Next() {
		var a CrossInstanceAction
		if err := rows.Scan(&a.ID, &a.GroupID, &a.ExternalID, &a.Title, &a.Action, &a.Reason,
			&a.SeerrRequestID, &a.SourceInstanceID, &a.TargetInstanceID, &a.ExecutedAt); err != nil {
			return nil, fmt.Errorf("scan cross instance action: %w", err)
		}
		actions = append(actions, a)
	}
	return actions, nil
}
