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
		 RETURNING id, username, password, totp_secret, created_at, updated_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, username, password, totp_secret, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, username, password, totp_secret, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.TOTPSecret, &u.CreatedAt, &u.UpdatedAt)
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

func (db *DB) UserCount(ctx context.Context) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// --- Sessions ---

func (db *DB) CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*Session, error) {
	var s Session
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO sessions (user_id, expires_at) VALUES ($1, $2)
		 RETURNING id, user_id, expires_at, created_at`,
		userID, time.Now().Add(duration),
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &s, nil
}

func (db *DB) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	var s Session
	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1 AND expires_at > now()`,
		id,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
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
		`SELECT app_type, hunt_missing_count, hunt_upgrade_count, hunt_missing_mode,
		        upgrade_mode, sleep_duration, monitored_only, skip_future,
		        hourly_cap, random_selection, debug_mode
		 FROM app_settings WHERE app_type = $1`,
		string(appType),
	).Scan(&s.AppType, &s.HuntMissingCount, &s.HuntUpgradeCount, &s.HuntMissingMode,
		&s.UpgradeMode, &s.SleepDuration, &s.MonitoredOnly, &s.SkipFuture,
		&s.HourlyCap, &s.RandomSelection, &s.DebugMode)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) UpdateAppSettings(ctx context.Context, s *AppSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE app_settings SET
		    hunt_missing_count = $1, hunt_upgrade_count = $2, hunt_missing_mode = $3,
		    upgrade_mode = $4, sleep_duration = $5, monitored_only = $6, skip_future = $7,
		    hourly_cap = $8, random_selection = $9, debug_mode = $10
		 WHERE app_type = $11`,
		s.HuntMissingCount, s.HuntUpgradeCount, s.HuntMissingMode,
		s.UpgradeMode, s.SleepDuration, s.MonitoredOnly, s.SkipFuture,
		s.HourlyCap, s.RandomSelection, s.DebugMode, string(s.AppType),
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
		 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
		string(appType), instanceID, mediaID, operation,
	)
	return err
}

func (db *DB) ResetState(ctx context.Context, appType AppType, instanceID uuid.UUID) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

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

// --- Hunt History ---

func (db *DB) AddHuntHistory(ctx context.Context, appType AppType, instanceID uuid.UUID, instanceName string, mediaID int, mediaTitle, operation string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO hunt_history (app_type, instance_id, instance_name, media_id, media_title, operation)
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

func (db *DB) ListHuntHistory(ctx context.Context, q HistoryQuery) ([]HuntHistory, int, error) {
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
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM hunt_history "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT id, app_type, instance_id, instance_name, media_id, media_title, operation, created_at
		 FROM hunt_history %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argN, argN+1,
	)
	args = append(args, q.Limit, q.Offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []HuntHistory
	for rows.Next() {
		var h HuntHistory
		if err := rows.Scan(&h.ID, &h.AppType, &h.InstanceID, &h.InstanceName, &h.MediaID, &h.MediaTitle, &h.Operation, &h.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, h)
	}

	return results, total, rows.Err()
}

func (db *DB) DeleteHistory(ctx context.Context, appType AppType) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM hunt_history WHERE app_type = $1`, string(appType))
	return err
}

// --- Hunt Stats ---

func (db *DB) GetAllStats(ctx context.Context) ([]HuntStats, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT app_type, hunted, upgraded, updated_at FROM hunt_stats ORDER BY app_type`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[HuntStats])
}

func (db *DB) IncrementStats(ctx context.Context, appType AppType, hunted, upgraded int64) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE hunt_stats SET hunted = hunted + $1, upgraded = upgraded + $2, updated_at = now()
		 WHERE app_type = $3`,
		hunted, upgraded, string(appType),
	)
	return err
}

func (db *DB) ResetStats(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE hunt_stats SET hunted = 0, upgraded = 0, updated_at = now()`,
	)
	return err
}

// --- Hourly Caps ---

func (db *DB) GetCurrentHourHits(ctx context.Context, appType AppType) (int, error) {
	var hits int
	err := db.Pool.QueryRow(ctx,
		`SELECT COALESCE(api_hits, 0) FROM hourly_caps
		 WHERE app_type = $1 AND hour_bucket = date_trunc('hour', now())`,
		string(appType),
	).Scan(&hits)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return hits, err
}

func (db *DB) IncrementHourlyHits(ctx context.Context, appType AppType, count int) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO hourly_caps (app_type, hour_bucket, api_hits)
		 VALUES ($1, date_trunc('hour', now()), $2)
		 ON CONFLICT (app_type, hour_bucket)
		 DO UPDATE SET api_hits = hourly_caps.api_hits + $2`,
		string(appType), count,
	)
	return err
}

func (db *DB) GetAllHourlyCaps(ctx context.Context) ([]HourlyCap, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT app_type, hour_bucket, api_hits FROM hourly_caps
		 WHERE hour_bucket = date_trunc('hour', now()) ORDER BY app_type`,
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

func (db *DB) ListScheduleExecutions(ctx context.Context, limit int) ([]struct {
	ID         int64     `json:"id"`
	ScheduleID uuid.UUID `json:"schedule_id"`
	ExecutedAt time.Time `json:"executed_at"`
	Result     *string   `json:"result"`
}, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, schedule_id, executed_at, result FROM schedule_executions
		 ORDER BY executed_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type ScheduleExec struct {
		ID         int64     `json:"id"`
		ScheduleID uuid.UUID `json:"schedule_id"`
		ExecutedAt time.Time `json:"executed_at"`
		Result     *string   `json:"result"`
	}

	var result []struct {
		ID         int64     `json:"id"`
		ScheduleID uuid.UUID `json:"schedule_id"`
		ExecutedAt time.Time `json:"executed_at"`
		Result     *string   `json:"result"`
	}
	for rows.Next() {
		var e struct {
			ID         int64     `json:"id"`
			ScheduleID uuid.UUID `json:"schedule_id"`
			ExecutedAt time.Time `json:"executed_at"`
			Result     *string   `json:"result"`
		}
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
	defer br.Close()

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
