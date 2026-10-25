// Package queue connects to the job queue. The queue is Asynq on Redis, in the queue database.
package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
)

// The client of the API uses short timeouts, so a request does not wait long for the queue.
// Asynq has no option to turn the retries off, so a call can still take a few seconds when Redis refuses
// the connection. The adapter that enqueues also limits the wait with a context.
const (
	clientDialTimeout  = 1 * time.Second
	clientReadTimeout  = 500 * time.Millisecond
	clientWriteTimeout = 500 * time.Millisecond
)

// ClientOpt gives the Redis settings of a queue client, for example the client of the API.
// The timeouts are short.
func ClientOpt(cfg config.Redis) asynq.RedisClientOpt {
	opt := WorkerOpt(cfg)
	opt.DialTimeout = clientDialTimeout
	opt.ReadTimeout = clientReadTimeout
	opt.WriteTimeout = clientWriteTimeout
	return opt
}

// WorkerOpt gives the Redis settings of a worker, and of an inspector that reads the queue.
// It uses the default timeouts of Asynq.
func WorkerOpt(cfg config.Redis) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{Addr: cfg.Addr(), Password: cfg.Password, DB: cfg.QueueDB}
}

// Open connects a queue client to Redis and checks the connection.
func Open(cfg config.Redis) (*asynq.Client, error) {
	client := asynq.NewClient(ClientOpt(cfg))
	if err := client.Ping(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("The application cannot connect to the Redis queue at %s. "+
			"Check REDIS_HOST, REDIS_PORT, REDIS_PASSWORD and REDIS_QUEUE_DB. Make sure that Redis runs. Cause: %w", cfg.Addr(), err)
	}
	return client, nil
}

// NewErrorHandler gives the error handler of a worker. It logs each task that fails, also the task
// that the queue does not try again. Asynq does not log the error of a task handler by itself.
func NewErrorHandler(log *slog.Logger) asynq.ErrorHandler {
	return asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
		retried, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)
		log.ErrorContext(ctx, "The worker did not complete the task.",
			slog.String("type", task.Type()), slog.Int("retried", retried), slog.Int("maxRetry", maxRetry), slog.Any("error", err))
	})
}
