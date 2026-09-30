package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

type migration struct {
	version  int64
	name     string
	up       string
	down     string
	checksum string
}

// Up initializes a fresh database or applies pending migrations in order.
// Tests can call Up on a connection configured with an isolated search_path.
func Up(ctx context.Context, db *sql.DB) error {
	return apply(ctx, db, 0)
}

// Down rolls back steps migrations in reverse order, including their data.
func Down(ctx context.Context, db *sql.DB, steps int) error {
	if steps < 1 {
		return errors.New("migration rollback steps must be positive")
	}
	return apply(ctx, db, steps)
}

func apply(ctx context.Context, db *sql.DB, steps int) error {
	migrations, err := load()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migrations: %w", err)
	}
	defer tx.Rollback()
	// Serialize startup and rollback before creating or reading the version table.
	// A transaction lock is released on commit, rollback, or connection loss.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(
		hashtextextended(current_database() || ':' || current_schema() || ':myim:migrations', 0))`); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		name TEXT NOT NULL,
		checksum TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	applied, err := validateHistory(ctx, tx, migrations)
	if err != nil {
		return err
	}
	if steps > applied {
		return fmt.Errorf("cannot roll back %d migrations: only %d applied", steps, applied)
	}
	if steps == 0 {
		for _, m := range migrations[applied:] {
			if _, err := tx.ExecContext(ctx, m.up); err != nil {
				return fmt.Errorf("apply migration %06d_%s: %w", m.version, m.name, err)
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO schema_migrations (version, name, checksum)
				VALUES ($1, $2, $3)
			`, m.version, m.name, m.checksum); err != nil {
				return fmt.Errorf("record migration %d: %w", m.version, err)
			}
		}
	} else {
		for i := applied - 1; i >= applied-steps; i-- {
			m := migrations[i]
			if _, err := tx.ExecContext(ctx, m.down); err != nil {
				return fmt.Errorf("rollback migration %06d_%s: %w", m.version, m.name, err)
			}
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM schema_migrations
				WHERE version = $1
			`, m.version); err != nil {
				return fmt.Errorf("remove migration %d: %w", m.version, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

func validateHistory(ctx context.Context, tx *sql.Tx, migrations []migration) (int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT version, name, checksum
		FROM schema_migrations
		ORDER BY version
	`)
	if err != nil {
		return 0, fmt.Errorf("read migration history: %w", err)
	}
	defer rows.Close()
	applied := 0
	for rows.Next() {
		var version int64
		var name, checksum string
		if err := rows.Scan(&version, &name, &checksum); err != nil {
			return 0, fmt.Errorf("scan migration history: %w", err)
		}
		if applied >= len(migrations) || migrations[applied].version != version {
			return 0, fmt.Errorf("migration history is not a registered prefix at version %d", version)
		}
		m := migrations[applied]
		if m.name != name || m.checksum != checksum {
			return 0, fmt.Errorf("applied migration %d differs from registered SQL", version)
		}
		applied++
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("read migration history: %w", err)
	}
	return applied, nil
}

func load() ([]migration, error) {
	entries, err := fs.Glob(migrationFiles, "sql/*.up.sql")
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("no registered migrations")
	}
	var result []migration
	for _, upPath := range entries {
		base := strings.TrimSuffix(strings.TrimPrefix(upPath, "sql/"), ".up.sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 || len(parts[0]) != 6 || parts[1] == "" {
			return nil, fmt.Errorf("invalid migration filename %q", upPath)
		}
		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || version < 1 || fmt.Sprintf("%06d", version) != parts[0] {
			return nil, fmt.Errorf("invalid migration version in %q", upPath)
		}
		up, err := migrationFiles.ReadFile(upPath)
		if err != nil {
			return nil, err
		}
		downPath := "sql/" + base + ".down.sql"
		down, err := migrationFiles.ReadFile(downPath)
		if err != nil {
			return nil, fmt.Errorf("read down migration %q: %w", downPath, err)
		}
		if strings.TrimSpace(string(up)) == "" || strings.TrimSpace(string(down)) == "" {
			return nil, fmt.Errorf("empty migration SQL for %q", base)
		}
		// Normalize line endings so Git checkout settings cannot change history.
		upSQL := strings.ReplaceAll(string(up), "\r\n", "\n")
		downSQL := strings.ReplaceAll(string(down), "\r\n", "\n")
		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(upSQL+"\x00"+downSQL)))
		result = append(result, migration{version: version, name: parts[1], up: upSQL, down: downSQL, checksum: checksum})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].version < result[j].version })
	for i := 1; i < len(result); i++ {
		if result[i-1].version == result[i].version {
			return nil, fmt.Errorf("duplicate migration version %d", result[i].version)
		}
	}
	return result, nil
}
