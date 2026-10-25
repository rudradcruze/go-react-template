// Package server builds the HTTP engine of the API.
package server

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/problem"
)

// New builds the Gin engine with all routes of the API. Each function in routes adds the routes
// of one context to the group /api/v1.
func New(cfg config.Config, log *slog.Logger, routes ...func(*gin.RouterGroup)) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(requestLog(log), recoverPanic(log))
	engine.NoRoute(func(c *gin.Context) {
		problem.Write(c, http.StatusNotFound, "The resource does not exist. Check the address. Then try again.")
	})
	// The container platform uses this route to see that the process runs. It does not check Postgres or Redis.
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	registerDocs(engine, cfg.Docs)
	api := engine.Group("/api/v1")
	for _, register := range routes {
		register(api)
	}
	return engine
}

func requestLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.LogAttrs(c.Request.Context(), slog.LevelInfo, "The server handled the request.",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}

// recoverPanic sends a general 500 response and writes the cause to the log.
func recoverPanic(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, cause any) {
		log.ErrorContext(c.Request.Context(), "The server could not handle the request.", slog.Any("cause", cause))
		problem.WriteInternalError(c)
	})
}
