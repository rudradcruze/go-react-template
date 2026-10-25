package queue

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
)

// NewLogger gives the logger of an Asynq worker. It writes to slog.
func NewLogger(log *slog.Logger) asynq.Logger { return slogLogger{log: log} }

type slogLogger struct{ log *slog.Logger }

func (l slogLogger) Debug(args ...any) { l.log.Debug(fmt.Sprint(args...)) }
func (l slogLogger) Info(args ...any)  { l.log.Info(fmt.Sprint(args...)) }
func (l slogLogger) Warn(args ...any)  { l.log.Warn(fmt.Sprint(args...)) }
func (l slogLogger) Error(args ...any) { l.log.Error(fmt.Sprint(args...)) }

// Fatal logs the message and stops the process, as Asynq expects.
func (l slogLogger) Fatal(args ...any) {
	l.log.Error(fmt.Sprint(args...))
	os.Exit(1)
}
