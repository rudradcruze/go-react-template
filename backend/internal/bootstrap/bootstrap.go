// Package bootstrap connects the parts of the application: it builds the adapters and the use cases of
// each context, and the HTTP engine. The API command and the E2E tests both use it.
package bootstrap

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/server"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	userhttp "github.com/rudradcruze/go-react-template/backend/internal/user/http"
	"github.com/rudradcruze/go-react-template/backend/internal/user/infra"
)

// NewAPI builds the HTTP handler of the API.
func NewAPI(cfg config.Config, log *slog.Logger, db *gorm.DB, cacheClient *redis.Client) http.Handler {
	users := infra.NewUserRepository(db)
	userCache := infra.NewRedisUserCache(cacheClient)
	registerUser := app.NewRegisterUser(users, time.Now)
	getUser := app.NewGetUser(users, userCache, log)
	userHandler := userhttp.NewHandler(registerUser, getUser, log)
	return server.New(cfg, log, userHandler.Routes)
}
