package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version int64
	name    string
	sql     string
}

type migrator struct {
	log *slog.Logger
	db  *sql.DB
}

func newMigrator(log *slog.Logger, db *sql.DB) *migrator {
	return &migrator{log: log, db: db}
}

func (m *migrator) run(ctx context.Context) error {
	if err := m.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Apply pending migrations in a transaction
	for _, migration := range migrations {
		if !applied[migration.version] {
			// Start transaction for each migration
			tx, err := m.db.BeginTx(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to start transaction: %w", err)
			}

			// Execute migration
			if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to apply migration %d: %w", migration.version, err)
			}

			// Record migration
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO schema_migration (version) VALUES ($1)",
				migration.version,
			); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %d: %w", migration.version, err)
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", migration.version, err)
			}

			m.log.Info("Applied migration", slog.Int64("version", migration.version), slog.String("name", migration.name))
		}
	}

	return nil
}

func (m *migrator) initialize(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migration (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.ExecContext(ctx, query)
	return err
}

func (m *migrator) getAppliedMigrations(ctx context.Context) (map[int64]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migration ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			version, name, err := parseMigrationFileName(entry.Name())
			if err != nil {
				return nil, err
			}

			content, err := migrationsFS.ReadFile(filepath.Join("migrations", entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
			}

			migrations = append(migrations, migration{
				version: version,
				name:    name,
				sql:     string(content),
			})
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

// parseMigrationFileName extracts version and name from filename like "000001_create_table.sql"
func parseMigrationFileName(filename string) (int64, string, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration version: %s", parts[0])
	}

	name := strings.TrimSuffix(parts[1], ".sql")
	return version, name, nil
}
