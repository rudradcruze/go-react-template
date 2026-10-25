package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxDisplayNameLength is the maximum length of a display name, in characters.
const MaxDisplayNameLength = 100

// DisplayName is the name that the application shows for a User.
type DisplayName string

// NewDisplayName removes the spaces around a name and checks it.
// A name must have a character that a person can see, no control character, and a maximum length.
func NewDisplayName(raw string) (DisplayName, error) {
	name := strings.TrimSpace(raw)
	visible := false
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrInvalidDisplayName
		}
		// A zero-width joiner is valid inside a name, for example in an emoji, but it is not visible alone.
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			visible = true
		}
	}
	if !visible {
		return "", ErrEmptyDisplayName
	}
	if utf8.RuneCountInString(name) > MaxDisplayNameLength {
		return "", ErrDisplayNameTooLong
	}
	return DisplayName(name), nil
}
