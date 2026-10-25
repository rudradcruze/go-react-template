// Command api starts the HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/logger"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/server"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		// The settings are not known yet, so use a logger with the default level.
		logger.New(slog.LevelInfo, os.Stderr).Error("The settings are not valid.", slog.Any("error", err))
		return err
	}
	log := logger.New(cfg.LogLevel, os.Stdout)

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.APIPort),
		Handler:           server.New(cfg, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("The server did not stop correctly.", slog.Any("error", err))
		}
	}()

	log.Info("The API starts.", slog.Int("port", cfg.APIPort), slog.Bool("docs", cfg.Docs.Enabled))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("The server stopped with an error.", slog.Any("error", err))
		return err
	}
	return nil
}
