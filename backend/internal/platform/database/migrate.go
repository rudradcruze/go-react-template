// Package database connects to Postgres and runs the migrations.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

// MigrateUp applies all new migrations in dir. It does nothing when dir has no migration files.
func MigrateUp(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("The migrations in %s did not run. %w", dir, nameFailedMigration(err))
	}
	return nil
}

// MigrateDown reverts the last migration that is applied.
// It does nothing when no migration is applied, or when dir has no migration files.
func MigrateDown(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	if _, err := provider.Down(ctx); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		return fmt.Errorf("The migration in %s did not revert. %w", dir, nameFailedMigration(err))
	}
	return nil
}

// newProvider returns a nil provider when dir has no migration files.
// A path that is not a folder is an error, because a wrong path must not look like "no migrations".
func newProvider(db *sql.DB, dir string) (*goose.Provider, error) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("The path %s is not a folder that the application can read. Check the path.", dir)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(dir))
	if errors.Is(err, goose.ErrNoMigrations) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("The migrations in %s are not valid. Cause: %w", dir, err)
	}
	return provider, nil
}

// nameFailedMigration adds the file name of the failed migration to the error.
func nameFailedMigration(err error) error {
	var partial *goose.PartialError
	if errors.As(err, &partial) && partial.Failed != nil {
		return fmt.Errorf("The migration %s failed. Cause: %w", partial.Failed.Source.Path, partial.Err)
	}
	return fmt.Errorf("Cause: %w", err)
}
