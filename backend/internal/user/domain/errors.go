package domain

import "errors"

// These are the errors of the rules of the domain. A caller tests for them with errors.Is.
var (
	ErrInvalidEmailAddress = errors.New("The email address is not valid.")
	ErrEmptyDisplayName    = errors.New("The display name is empty.")
	ErrInvalidDisplayName  = errors.New("The display name has a character that is not allowed.")
	ErrDisplayNameTooLong  = errors.New("The display name is too long.")
	ErrEmailTaken          = errors.New("The email address is already registered.")
)
