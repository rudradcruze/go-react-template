// Package domain holds the rules of the User context. It imports no framework, database or cache.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User is a person who has registered with the application.
type User struct {
	ID           uuid.UUID
	EmailAddress EmailAddress
	DisplayName  DisplayName
	CreatedAt    time.Time
}

// NewUser registers a User at the given time. It makes the ID, a UUID of version 7.
func NewUser(emailAddress, displayName string, now time.Time) (User, error) {
	email, err := NewEmailAddress(emailAddress)
	if err != nil {
		return User{}, err
	}
	name, err := NewDisplayName(displayName)
	if err != nil {
		return User{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return User{}, fmt.Errorf("The ID of the User is not available. Cause: %w", err)
	}
	return User{
		ID:           id,
		EmailAddress: email,
		DisplayName:  name,
		// Postgres keeps microseconds. Truncate here, so the User is the same after it is saved.
		CreatedAt: now.UTC().Truncate(time.Microsecond),
	}, nil
}
