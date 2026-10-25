package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// GetUser is the use case that reads a User by ID. It reads through the cache.
type GetUser struct {
	users UserRepository
	cache UserCache
	log   *slog.Logger
}

// NewGetUser builds the use case.
func NewGetUser(users UserRepository, cache UserCache, log *slog.Logger) *GetUser {
	return &GetUser{users: users, cache: cache, log: log}
}

// Execute gives the User with this ID. It reads the cache first, and then the database.
// It returns domain.ErrUserNotFound when no User has the ID.
//
// The cache is not the source of truth. When the cache fails to read, Execute logs the error and reads
// the database. It does not write to a cache that just failed, so a Redis failure does not stop the reads.
func (g *GetUser) Execute(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, found, err := g.cache.Get(ctx, id)
	cacheWorks := err == nil
	if err != nil {
		g.log.WarnContext(ctx, "The application did not read the User from the cache.", slog.Any("error", err))
	} else if found {
		return user, nil
	}

	user, err = g.users.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	if cacheWorks {
		if err := g.cache.Set(ctx, user); err != nil {
			g.log.WarnContext(ctx, "The application did not keep the User in the cache.", slog.Any("error", err))
		}
	}
	return user, nil
}
