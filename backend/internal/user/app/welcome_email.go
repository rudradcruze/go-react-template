package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// WelcomeEmailTaskType is the type of the task that sends the Welcome email.
	WelcomeEmailTaskType = "user:send_welcome_email"
	// WelcomeEmailMaxRetry is the number of times that the queue tries again after a failure.
	// After the last try, the queue archives the task.
	WelcomeEmailMaxRetry = 3
	// WelcomeEmailTimeout is the time that one try can use. A mailer that does not answer must not hold a worker for long.
	WelcomeEmailTimeout = 1 * time.Minute
)

// WelcomeEmailPayload is the JSON payload of the task. It has only the ID, so the worker sends the current data of the User.
type WelcomeEmailPayload struct {
	UserID uuid.UUID `json:"userId"`
}

// SendWelcomeEmail is the use case that sends the Welcome email. The worker runs it.
type SendWelcomeEmail struct {
	users  UserRepository
	mailer Mailer
}

// NewSendWelcomeEmail builds the use case.
func NewSendWelcomeEmail(users UserRepository, mailer Mailer) *SendWelcomeEmail {
	return &SendWelcomeEmail{users: users, mailer: mailer}
}

// Execute sends the Welcome email to the User with this ID.
// It returns domain.ErrUserNotFound when no User has the ID.
func (s *SendWelcomeEmail) Execute(ctx context.Context, userID uuid.UUID) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	return s.mailer.Send(ctx, Email{
		To:      user.EmailAddress,
		Subject: "Welcome to go-react-template",
		Body:    fmt.Sprintf("Hello %s.\n\nYour Registration is complete. You can use go-react-template now.", user.DisplayName),
	})
}
