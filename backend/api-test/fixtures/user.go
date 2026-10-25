// Package fixtures builds test data. A builder gives a valid object with default values,
// and each call gives unique values for the unique fields.
package fixtures

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

var userCounter atomic.Uint64

type userValues struct {
	emailAddress string
	displayName  string
	createdAt    time.Time
}

// UserOption changes a value of the User that User builds.
type UserOption func(*userValues)

// WithEmailAddress sets the email address.
func WithEmailAddress(emailAddress string) UserOption {
	return func(v *userValues) { v.emailAddress = emailAddress }
}

// WithDisplayName sets the display name.
func WithDisplayName(displayName string) UserOption {
	return func(v *userValues) { v.displayName = displayName }
}

// WithCreatedAt sets the time of the Registration.
func WithCreatedAt(createdAt time.Time) UserOption {
	return func(v *userValues) { v.createdAt = createdAt }
}

// User builds a valid User. It panics when an option gives a value that the domain does not accept,
// because that is a mistake in the test.
func User(options ...UserOption) domain.User {
	n := userCounter.Add(1)
	values := userValues{
		emailAddress: fmt.Sprintf("user%d@example.com", n),
		displayName:  fmt.Sprintf("User %d", n),
		createdAt:    time.Now(),
	}
	for _, option := range options {
		option(&values)
	}
	user, err := domain.NewUser(values.emailAddress, values.displayName, values.createdAt)
	if err != nil {
		panic(fmt.Sprintf("The fixture User is not valid. Cause: %v", err))
	}
	return user
}

// SaveUser builds a User and saves it through the repository.
func SaveUser(t *testing.T, users app.UserRepository, options ...UserOption) domain.User {
	t.Helper()
	user := User(options...)
	if err := users.Create(t.Context(), user); err != nil {
		t.Fatalf("The fixture User did not save. Cause: %v", err)
	}
	return user
}
