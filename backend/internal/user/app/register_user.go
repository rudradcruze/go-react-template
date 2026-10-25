package app

import (
	"context"
	"fmt"
	"log/slog"
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
	users         UserRepository
	welcomeEmails WelcomeEmailQueue
	log           *slog.Logger
	now           func() time.Time
}

// NewRegisterUser builds the use case. now gives the time of the Registration.
func NewRegisterUser(users UserRepository, welcomeEmails WelcomeEmailQueue, log *slog.Logger, now func() time.Time) *RegisterUser {
	return &RegisterUser{users: users, welcomeEmails: welcomeEmails, log: log, now: now}
}

// Execute registers a User. It returns an error of the domain when the input breaks a rule,
// and domain.ErrEmailTaken when the email address is already registered.
//
// After the database saves the User, Execute queues the Welcome email. The email is not essential:
// when the queue fails, Execute logs the error and the Registration still succeeds.
func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (domain.User, error) {
	user, err := domain.NewUser(input.EmailAddress, input.DisplayName, r.now())
	if err != nil {
		return domain.User{}, err
	}
	if err := r.users.Create(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("The application did not save the User. Cause: %w", err)
	}
	if err := r.welcomeEmails.EnqueueWelcomeEmail(ctx, user.ID); err != nil {
		r.log.ErrorContext(ctx, "The application did not queue the Welcome email.",
			slog.String("userId", user.ID.String()), slog.Any("error", err))
	}
	return user, nil
}
