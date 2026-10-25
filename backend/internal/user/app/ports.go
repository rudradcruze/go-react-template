// Package app holds the use cases of the User context, and the ports that the use cases call.
package app

import (
	"context"

	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// UserRepository stores Users.
type UserRepository interface {
	// Create saves a new User. It returns domain.ErrEmailTaken when another User has the same email address.
	Create(ctx context.Context, user domain.User) error
}
