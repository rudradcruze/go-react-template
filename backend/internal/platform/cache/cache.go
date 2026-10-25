// Package cache connects to the Redis database of the cache.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
)

const connectTimeout = 5 * time.Second

// The cache is optional: when Redis fails, the application reads the database. So a call to Redis
// must fail fast, and it must not retry. A Redis that does not answer then slows a request by under 1 second.
const (
	dialTimeout  = 1 * time.Second
	readTimeout  = 500 * time.Millisecond
	writeTimeout = 500 * time.Millisecond
)

// Open connects to Redis, in the cache database, and checks the connection.
func Open(ctx context.Context, cfg config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.CacheDB,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		MaxRetries:   -1, // -1 turns the retries off.
	})
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("The application cannot connect to Redis at %s. "+
			"Check REDIS_HOST, REDIS_PORT and REDIS_PASSWORD. Make sure that Redis runs. Cause: %w", cfg.Addr(), err)
	}
	return client, nil
}
