// Package bootstrap connects the parts of the application: it builds the adapters and the use cases of
// each context, and the HTTP engine. The API command and the E2E tests both use it.
package bootstrap

import (
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/server"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	userhttp "github.com/rudradcruze/go-react-template/backend/internal/user/http"
	"github.com/rudradcruze/go-react-template/backend/internal/user/infra"
)

// NewAPI builds the HTTP handler of the API.
func NewAPI(cfg config.Config, log *slog.Logger, db *gorm.DB) http.Handler {
	users := infra.NewUserRepository(db)
	registerUser := app.NewRegisterUser(users, time.Now)
	userHandler := userhttp.NewHandler(registerUser, log)
	return server.New(cfg, log, userHandler.Routes)
}
