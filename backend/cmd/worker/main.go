// Command worker runs the background tasks of the application. It reads the tasks from the Redis queue.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"github.com/rudradcruze/go-react-template/backend/internal/bootstrap"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/database"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/logger"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/queue"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadWorker()
	if err != nil {
		logger.New(slog.LevelInfo, os.Stderr).Error("The settings are not valid.", slog.Any("error", err))
		return err
	}
	log := logger.New(cfg.LogLevel, os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.Postgres)
	if err != nil {
		log.Error("The worker cannot start.", slog.Any("error", err))
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("The worker cannot start.", slog.Any("error", err))
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	server := asynq.NewServer(queue.WorkerOpt(cfg.Redis), asynq.Config{
		Concurrency:  cfg.Concurrency,
		Logger:       queue.NewLogger(log),
		ErrorHandler: queue.NewErrorHandler(log),
	})
	// Start does not check the connection. Check it now, so a wrong Redis setting stops the worker at once.
	if err := server.Ping(); err != nil {
		log.Error("The worker cannot start. The application cannot connect to the Redis queue. "+
			"Check REDIS_HOST, REDIS_PORT, REDIS_PASSWORD and REDIS_QUEUE_DB. Make sure that Redis runs.", slog.Any("error", err))
		return err
	}
	if err := server.Start(bootstrap.NewWorkerMux(log, db)); err != nil {
		log.Error("The worker cannot start.", slog.Any("error", err))
		return err
	}
	log.Info("The worker starts.", slog.Int("concurrency", cfg.Concurrency))

	<-ctx.Done()
	// Shutdown waits until the running tasks end, for the time that asynq.Config.ShutdownTimeout gives (8 seconds by default).
	server.Shutdown()
	log.Info("The worker stops.")
	return nil
}
