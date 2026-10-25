package app

import (
	"context"
	"fmt"
	"time"

	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// RegisterUserInput is the data of a Registration, as the caller gives it.
type RegisterUserInput struct {
	EmailAddress string
	DisplayName  string
}

// RegisterUser is the use case of a Registration.
type RegisterUser struct {
	users UserRepository
	now   func() time.Time
}

// NewRegisterUser builds the use case. now gives the time of the Registration.
func NewRegisterUser(users UserRepository, now func() time.Time) *RegisterUser {
	return &RegisterUser{users: users, now: now}
}

// Execute registers a User. It returns an error of the domain when the input breaks a rule,
// and domain.ErrEmailTaken when the email address is already registered.
func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (domain.User, error) {
	user, err := domain.NewUser(input.EmailAddress, input.DisplayName, r.now())
	if err != nil {
		return domain.User{}, err
	}
	if err := r.users.Create(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("The application did not save the User. Cause: %w", err)
	}
	return user, nil
}
