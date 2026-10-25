package platform_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/queue"
)

func TestQueue_NewLogger_WritesEachLevelToSlog(t *testing.T) {
	t.Parallel()
	logs := &bytes.Buffer{}
	logger := queue.NewLogger(slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})))

	logger.Debug("one ", 1)
	logger.Info("two")
	logger.Warn("three")
	logger.Error("four")

	for level, message := range map[string]string{"DEBUG": "one 1", "INFO": "two", "WARN": "three", "ERROR": "four"} {
		assert.Contains(t, logs.String(), `"level":"`+level+`","msg":"`+message+`"`)
	}
}
