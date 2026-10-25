// Package bootstrap connects the parts of the application: it builds the adapters and the use cases of
// each context, and the HTTP engine. The API command and the E2E tests both use it.
package bootstrap

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/server"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	userhttp "github.com/rudradcruze/go-react-template/backend/internal/user/http"
	"github.com/rudradcruze/go-react-template/backend/internal/user/infra"
	"github.com/rudradcruze/go-react-template/backend/internal/user/jobs"
)

// NewAPI builds the HTTP handler of the API.
func NewAPI(cfg config.Config, log *slog.Logger, db *gorm.DB, cacheClient *redis.Client, queueClient *asynq.Client) http.Handler {
	users := infra.NewUserRepository(db)
	userCache := infra.NewRedisUserCache(cacheClient)
	welcomeEmails := infra.NewWelcomeEmailQueue(queueClient)
	registerUser := app.NewRegisterUser(users, welcomeEmails, log, time.Now)
	getUser := app.NewGetUser(users, userCache, log)
	userHandler := userhttp.NewHandler(registerUser, getUser, log)
	return server.New(cfg, log, userHandler.Routes)
}

// NewWorkerMux builds the handlers of the worker: it routes each task type to its handler.
// The mailer of the application writes the emails to the log.
func NewWorkerMux(log *slog.Logger, db *gorm.DB) *asynq.ServeMux {
	users := infra.NewUserRepository(db)
	sendWelcomeEmail := app.NewSendWelcomeEmail(users, infra.NewLogMailer(log))
	mux := asynq.NewServeMux()
	jobs.NewHandler(sendWelcomeEmail).Register(mux)
	return mux
}
