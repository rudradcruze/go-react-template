package user_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/fixtures"
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

func newGetUser(t *testing.T) (*app.GetUser, *mocks.MockUserRepository, *mocks.MockUserCache, *bytes.Buffer) {
	t.Helper()
	repository := mocks.NewMockUserRepository(t)
	cache := mocks.NewMockUserCache(t)
	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	return app.NewGetUser(repository, cache, logger), repository, cache, logs
}

func TestGetUser_Execute_ReturnsTheCachedUserWithoutTheDatabase(t *testing.T) {
	t.Parallel()
	useCase, _, cache, _ := newGetUser(t) // The repository has no expected call, so a call fails the test.
	user := fixtures.User()
	cache.EXPECT().Get(t.Context(), user.ID).Return(user, true, nil).Once()

	got, err := useCase.Execute(t.Context(), user.ID)

	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestGetUser_Execute_ReadsTheDatabaseAndKeepsTheUserInTheCacheOnAMiss(t *testing.T) {
	t.Parallel()
	useCase, repository, cache, _ := newGetUser(t)
	user := fixtures.User()
	cache.EXPECT().Get(t.Context(), user.ID).Return(domain.User{}, false, nil).Once()
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	cache.EXPECT().Set(t.Context(), user).Return(nil).Once()

	got, err := useCase.Execute(t.Context(), user.ID)

	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestGetUser_Execute_ReadsTheDatabaseAndDoesNotWriteTheCacheWhenTheCacheReadFails(t *testing.T) {
	t.Parallel()
	useCase, repository, cache, logs := newGetUser(t) // The cache has no expected Set call. A cache that fails to read fails to write too.
	user := fixtures.User()
	cache.EXPECT().Get(t.Context(), user.ID).Return(domain.User{}, false, errors.New("The Redis server is not reachable.")).Once()
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()

	got, err := useCase.Execute(t.Context(), user.ID)

	require.NoError(t, err)
	assert.Equal(t, user, got)
	assert.Equal(t, 1, strings.Count(logs.String(), `"level":"WARN"`))
	assert.Contains(t, logs.String(), "The Redis server is not reachable.")
}

func TestGetUser_Execute_ReturnsTheUserAndLogsTheErrorWhenTheCacheWriteFails(t *testing.T) {
	t.Parallel()
	useCase, repository, cache, logs := newGetUser(t)
	user := fixtures.User()
	cache.EXPECT().Get(t.Context(), user.ID).Return(domain.User{}, false, nil).Once()
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	cache.EXPECT().Set(t.Context(), user).Return(errors.New("The Redis server is not reachable.")).Once()

	got, err := useCase.Execute(t.Context(), user.ID)

	require.NoError(t, err)
	assert.Equal(t, user, got)
	assert.Contains(t, logs.String(), `"level":"WARN"`)
	assert.Contains(t, logs.String(), "The Redis server is not reachable.")
}

func TestGetUser_Execute_ReturnsNotFoundAndDoesNotCacheWhenNoUserHasTheID(t *testing.T) {
	t.Parallel()
	useCase, repository, cache, _ := newGetUser(t) // The cache has no expected Set call, so a Set call fails the test.
	id := uuid.Must(uuid.NewV7())
	cache.EXPECT().Get(t.Context(), id).Return(domain.User{}, false, nil).Once()
	repository.EXPECT().FindByID(t.Context(), id).Return(domain.User{}, domain.ErrUserNotFound).Once()

	got, err := useCase.Execute(t.Context(), id)

	require.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Equal(t, domain.User{}, got)
}

func TestGetUser_Execute_ReturnsTheUnexpectedErrorOfTheDatabase(t *testing.T) {
	t.Parallel()
	useCase, repository, cache, _ := newGetUser(t)
	id := uuid.Must(uuid.NewV7())
	failure := errors.New("The connection is closed.")
	cache.EXPECT().Get(t.Context(), id).Return(domain.User{}, false, nil).Once()
	repository.EXPECT().FindByID(t.Context(), id).Return(domain.User{}, failure).Once()

	_, err := useCase.Execute(t.Context(), id)

	require.ErrorIs(t, err, failure)
	require.NotErrorIs(t, err, domain.ErrUserNotFound)
}
