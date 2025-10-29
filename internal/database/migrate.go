package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/**/*.sql
var migrationFiles embed.FS

// MigrateUp applies all pending migrations.
func MigrateUp(db *sql.DB, driver string) error {
	m, err := newMigrator(db, driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// MigrateDown reverts all applied migrations.
func MigrateDown(db *sql.DB, driver string) error {
	m, err := newMigrator(db, driver)
	if err != nil {
		return err
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// MigrateSteps moves the schema by the specified number of steps. Positive values migrate up, negative values migrate down.
func MigrateSteps(db *sql.DB, driver string, steps int) error {
	if steps == 0 {
		return nil
	}

	m, err := newMigrator(db, driver)
	if err != nil {
		return err
	}

	if err := m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) && isNotExist(pathErr.Err) {
			return nil
		}
		if isNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

// MigrationVersion reports the current migration version and dirty flag.
func MigrationVersion(db *sql.DB, driver string) (version uint, dirty bool, err error) {
	m, err := newMigrator(db, driver)
	if err != nil {
		return 0, false, err
	}

	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

func newMigrator(db *sql.DB, driver string) (*migrate.Migrate, error) {
	normalizedDriver, err := normalizeDriver(driver)
	if err != nil {
		return nil, err
	}

	dbDriver, databaseName, sourcePath, err := migrationDatabaseDriver(db, normalizedDriver)
	if err != nil {
		return nil, err
	}

	sourceDriver, err := iofs.New(migrationFiles, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, databaseName, dbDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return m, nil
}

func migrationDatabaseDriver(db *sql.DB, driver string) (database.Driver, string, string, error) {
	switch driver {
	case "sqlite":
		drv, err := sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to init sqlite migrator: %w", err)
		}
		return drv, "sqlite", "migrations/sqlite", nil
	case "pgx":
		drv, err := pgxv5.WithInstance(db, &pgxv5.Config{})
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to init pgx migrator: %w", err)
		}
		return drv, "pgx5", "migrations/postgres", nil
	default:
		return nil, "", "", fmt.Errorf("unsupported migration driver: %s", driver)
	}
}

func isNotExist(err error) bool {
	if errors.Is(err, fs.ErrNotExist) {
		return true
	}
	return strings.Contains(err.Error(), "file does not exist")
}
