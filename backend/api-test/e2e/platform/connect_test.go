//go:build e2e

package platform_test

import (
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testenv"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/cache"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/database"
)

// unusedPort gives a port on which nothing listens.
func unusedPort(t *testing.T) int {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}

func TestDatabase_Open_ConnectsToPostgres(t *testing.T) {
	res := testenv.Begin(t)

	db, err := database.Open(t.Context(), res.Postgres)

	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	var answer int
	require.NoError(t, db.Raw("SELECT 1").Scan(&answer).Error)
	assert.Equal(t, 1, answer)
}

func TestDatabase_Open_FailsWhenPostgresIsNotReachable(t *testing.T) {
	cfg := testenv.Begin(t).Postgres
	cfg.Port = unusedPort(t)

	_, err := database.Open(t.Context(), cfg)

	require.ErrorContains(t, err, "POSTGRES_HOST")
}

func TestDatabase_Open_DoesNotShowThePasswordInTheErrorWhenTheCredentialsAreWrong(t *testing.T) {
	cfg := testenv.Begin(t).Postgres
	cfg.Password = "wrong-password-0123"

	_, err := database.Open(t.Context(), cfg)

	require.ErrorContains(t, err, "POSTGRES_PASSWORD")
	assert.NotContains(t, err.Error(), cfg.Password)
}

func TestCache_Open_ConnectsToRedis(t *testing.T) {
	res := testenv.Begin(t)

	client, err := cache.Open(t.Context(), res.Redis)

	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Set(t.Context(), "key", "value", 0).Err())
	value, err := client.Get(t.Context(), "key").Result()
	require.NoError(t, err)
	assert.Equal(t, "value", value)
}

func TestCache_Open_FailsWhenRedisIsNotReachable(t *testing.T) {
	cfg := testenv.Begin(t).Redis
	cfg.Port = unusedPort(t)

	_, err := cache.Open(t.Context(), cfg)

	require.ErrorContains(t, err, "REDIS_HOST")
}

func TestCache_Open_FailsFastWhenRedisDoesNotAnswer(t *testing.T) {
	// A server that accepts connections and never answers. The cache must not wait for it for many seconds.
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { _ = conn.Close() })
		}
	}()
	cfg := testenv.Begin(t).Redis
	cfg.Host, cfg.Port = "127.0.0.1", listener.Addr().(*net.TCPAddr).Port

	start := time.Now()
	_, err = cache.Open(t.Context(), cfg)

	require.ErrorContains(t, err, "REDIS_HOST")
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestCache_Open_UsesTheCacheDatabaseAndNotTheQueueDatabase(t *testing.T) {
	res := testenv.Begin(t)
	client, err := cache.Open(t.Context(), res.Redis)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Set(t.Context(), "key", "value", 0).Err())
	cacheView := redis.NewClient(&redis.Options{Addr: res.Redis.Addr(), DB: res.Redis.CacheDB})
	queueView := redis.NewClient(&redis.Options{Addr: res.Redis.Addr(), DB: res.Redis.QueueDB})
	t.Cleanup(func() { _ = cacheView.Close(); _ = queueView.Close() })

	inCache, err := cacheView.Exists(t.Context(), "key").Result()
	require.NoError(t, err)
	inQueue, err := queueView.Exists(t.Context(), "key").Result()
	require.NoError(t, err)

	assert.Equal(t, int64(1), inCache)
	assert.Equal(t, int64(0), inQueue)
}
