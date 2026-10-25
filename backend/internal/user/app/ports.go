// Package app holds the use cases of the User context, and the ports that the use cases call.
package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// UserRepository stores Users.
type UserRepository interface {
	// Create saves a new User. It returns domain.ErrEmailTaken when another User has the same email address.
	Create(ctx context.Context, user domain.User) error
	// FindByID gives the User with this ID. It returns domain.ErrUserNotFound when no User has the ID.
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// UserCache keeps Users for a short time, so a read does not need the database.
// The cache is not the source of truth. Postgres has the data.
type UserCache interface {
	// Get gives the User with this ID. found is false when the cache does not have the User.
	Get(ctx context.Context, id uuid.UUID) (user domain.User, found bool, err error)
	// Set keeps the User in the cache.
	Set(ctx context.Context, user domain.User) error
}
