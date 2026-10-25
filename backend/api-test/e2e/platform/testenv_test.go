//go:build e2e

package platform_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testenv"
)

func TestTestenv_Begin_GivesADatabaseThatExistsOnlyDuringTheTest(t *testing.T) {
	var name string

	t.Run("test with a database", func(t *testing.T) {
		name = testenv.Begin(t).Postgres.DB

		assert.True(t, testenv.DatabaseExists(t, name))
	})

	require.NotEmpty(t, name)
	assert.False(t, testenv.DatabaseExists(t, name))
}

func TestTestenv_Begin_GivesEachTestADatabaseOfItsOwn(t *testing.T) {
	var firstName string

	t.Run("first test writes", func(t *testing.T) {
		res := testenv.Begin(t)
		firstName = res.Postgres.DB
		db := testenv.OpenDB(t, res.Postgres)
		_, err := db.ExecContext(t.Context(), "CREATE TABLE isolation_probe (id int)")
		require.NoError(t, err)
	})

	t.Run("second test reads", func(t *testing.T) {
		res := testenv.Begin(t)
		db := testenv.OpenDB(t, res.Postgres)

		assert.NotEqual(t, firstName, res.Postgres.DB)
		assert.False(t, testenv.TableExists(t, db, "isolation_probe"), "the table of the first test is in the database of the second test")
	})
}

func TestTestenv_Begin_EmptiesRedisBeforeEachTest(t *testing.T) {
	t.Run("first test writes", func(t *testing.T) {
		res := testenv.Begin(t)
		client := testenv.OpenRedis(t, res.Redis)
		require.NoError(t, client.Set(t.Context(), "probe", "1", 0).Err())
	})

	t.Run("second test reads", func(t *testing.T) {
		res := testenv.Begin(t)
		client := testenv.OpenRedis(t, res.Redis)

		keys, err := client.Keys(t.Context(), "*").Result()

		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestTestenv_Begin_GivesTheMigratedSchemaOfTheTemplate(t *testing.T) {
	db := testenv.OpenDB(t, testenv.Begin(t).Postgres)

	assert.True(t, testenv.TableExists(t, db, "testenv_probe"), "the template database has no migrated tables")
}
