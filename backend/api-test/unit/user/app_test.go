package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/mocks"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

var registrationTime = time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)

func newRegisterUser(t *testing.T) (*app.RegisterUser, *mocks.MockUserRepository) {
	t.Helper()
	repository := mocks.NewMockUserRepository(t)
	return app.NewRegisterUser(repository, func() time.Time { return registrationTime }), repository
}

func TestRegisterUser_Execute_SavesTheNormalizedUserAndReturnsIt(t *testing.T) {
	t.Parallel()
	useCase, repository := newRegisterUser(t)
	var saved domain.User
	repository.EXPECT().
		Create(t.Context(), mock.MatchedBy(func(u domain.User) bool {
			return u.EmailAddress == "ann@example.com" && u.DisplayName == "Ann" && u.CreatedAt.Equal(registrationTime)
		})).
		Run(func(_ context.Context, user domain.User) { saved = user }).
		Return(nil).
		Once()

	got, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: " Ann@Example.COM ", DisplayName: " Ann "})

	require.NoError(t, err)
	assert.Equal(t, saved, got)
	assert.Equal(t, uuid.Version(7), got.ID.Version())
}

func TestRegisterUser_Execute_ReturnsEmailTakenWhenTheRepositoryReportsADuplicate(t *testing.T) {
	t.Parallel()
	useCase, repository := newRegisterUser(t)
	repository.EXPECT().Create(t.Context(), mock.MatchedBy(func(u domain.User) bool {
		return u.EmailAddress == "ann@example.com"
	})).Return(domain.ErrEmailTaken).Once()

	got, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "Ann"})

	require.ErrorIs(t, err, domain.ErrEmailTaken)
	assert.Equal(t, domain.User{}, got)
}

func TestRegisterUser_Execute_DoesNotCallTheRepositoryWhenTheInputBreaksARule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input app.RegisterUserInput
		want  error
	}{
		"email address is not valid": {
			input: app.RegisterUserInput{EmailAddress: "ann", DisplayName: "Ann"},
			want:  domain.ErrInvalidEmailAddress,
		},
		"display name is empty": {
			input: app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "  "},
			want:  domain.ErrEmptyDisplayName,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			useCase, _ := newRegisterUser(t) // The mock has no expected call, so a call fails the test.

			_, err := useCase.Execute(t.Context(), tt.input)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRegisterUser_Execute_ReturnsTheUnexpectedErrorOfTheRepository(t *testing.T) {
	t.Parallel()
	useCase, repository := newRegisterUser(t)
	failure := errors.New("The connection is closed.")
	repository.EXPECT().Create(t.Context(), mock.MatchedBy(func(u domain.User) bool {
		return u.EmailAddress == "ann@example.com" && u.DisplayName == "Ann"
	})).Return(failure).Once()

	_, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "Ann"})

	require.ErrorIs(t, err, failure)
	require.NotErrorIs(t, err, domain.ErrEmailTaken)
}
