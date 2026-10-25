package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
)

const connectTimeout = 5 * time.Second

// Open connects to Postgres through GORM with the pgx driver, and checks the connection.
// GORM translates database errors, so a repository can test for gorm.ErrDuplicatedKey.
func Open(ctx context.Context, cfg config.Postgres) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		TranslateError:       true,
		DisableAutomaticPing: true,
		Logger:               gormlogger.Discard,
	})
	if err != nil {
		return nil, connectError(cfg, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, connectError(cfg, err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, connectError(cfg, err)
	}
	return db, nil
}

func connectError(cfg config.Postgres, err error) error {
	return fmt.Errorf("The application cannot connect to Postgres at %s. "+
		"Check POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB and POSTGRES_SSLMODE. "+
		"Make sure that Postgres runs. Cause: %w", cfg.Addr(), err)
}
