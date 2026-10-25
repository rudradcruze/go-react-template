//go:build e2e

package platform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testenv"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/database"
)

// writeMigration writes a goose migration that makes a table on Up and drops it on Down.
func writeMigration(t *testing.T, dir, file, table string) {
	t.Helper()
	content := "-- +goose Up\nCREATE TABLE " + table + " (id int);\n\n-- +goose Down\nDROP TABLE " + table + ";\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, file), []byte(content), 0o600))
}

func TestDatabase_MigrateUp_RunsAllMigrations(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	dir := t.TempDir()
	writeMigration(t, dir, "00001_create_first.sql", "first")
	writeMigration(t, dir, "00002_create_second.sql", "second")

	err := database.MigrateUp(t.Context(), db, dir)

	require.NoError(t, err)
	assert.True(t, testenv.TableExists(t, db, "first"))
	assert.True(t, testenv.TableExists(t, db, "second"))
}

func TestDatabase_MigrateUp_IsSafeToRunTwice(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	dir := t.TempDir()
	writeMigration(t, dir, "00001_create_first.sql", "first")
	require.NoError(t, database.MigrateUp(t.Context(), db, dir))

	err := database.MigrateUp(t.Context(), db, dir)

	require.NoError(t, err)
	assert.True(t, testenv.TableExists(t, db, "first"))
}

func TestDatabase_MigrateUp_DoesNothingWithoutMigrationFiles(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)

	err := database.MigrateUp(t.Context(), db, t.TempDir())

	require.NoError(t, err)
}

func TestDatabase_MigrateUp_ReturnsAnErrorForABrokenMigration(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	dir := t.TempDir()
	broken := "-- +goose Up\nCREATE TABEL broken (id int);\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "00001_broken.sql"), []byte(broken), 0o600))

	err := database.MigrateUp(t.Context(), db, dir)

	require.ErrorContains(t, err, "00001_broken.sql")
}

func TestDatabase_MigrateDown_RevertsOnlyTheLastMigration(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	dir := t.TempDir()
	writeMigration(t, dir, "00001_create_first.sql", "first")
	writeMigration(t, dir, "00002_create_second.sql", "second")
	require.NoError(t, database.MigrateUp(t.Context(), db, dir))

	err := database.MigrateDown(t.Context(), db, dir)

	require.NoError(t, err)
	assert.True(t, testenv.TableExists(t, db, "first"))
	assert.False(t, testenv.TableExists(t, db, "second"))
}

func TestDatabase_MigrateUp_ReturnsAnErrorWhenTheFolderDoesNotExist(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	missing := filepath.Join(t.TempDir(), "missing")

	err := database.MigrateUp(t.Context(), db, missing)

	require.ErrorContains(t, err, missing)
}

func TestDatabase_MigrateDown_ReturnsAnErrorWhenTheFolderDoesNotExist(t *testing.T) {
	db := testenv.OpenDB(t, testenv.BeginEmpty(t).Postgres)
	missing := filepath.Join(t.TempDir(), "missing")

	err := database.MigrateDown(t.Context(), db, missing)

	require.ErrorContains(t, err, missing)
}
