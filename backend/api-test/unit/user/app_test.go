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

// welcomeEmailOfAnn is the Welcome email that the application sends to Ann. The text is a literal on purpose:
// a test must not build the expected text in the same way as the code.
var welcomeEmailOfAnn = app.Email{
	To:      "ann@example.com",
	Subject: "Welcome to go-react-template",
	Body:    "Hello Ann.\n\nYour Registration is complete. You can use go-react-template now.",
}

var registrationTime = time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)

func newRegisterUser(t *testing.T) (*app.RegisterUser, *mocks.MockUserRepository, *mocks.MockWelcomeEmailQueue, *bytes.Buffer) {
	t.Helper()
	repository := mocks.NewMockUserRepository(t)
	queue := mocks.NewMockWelcomeEmailQueue(t)
	logs := &bytes.Buffer{}
	useCase := app.NewRegisterUser(repository, queue, slog.New(slog.NewJSONHandler(logs, nil)), func() time.Time { return registrationTime })
	return useCase, repository, queue, logs
}

func TestRegisterUser_Execute_SavesTheNormalizedUserAndReturnsIt(t *testing.T) {
	t.Parallel()
	useCase, repository, queue, _ := newRegisterUser(t)
	var saved domain.User
	repository.EXPECT().
		Create(t.Context(), mock.MatchedBy(func(u domain.User) bool {
			return u.EmailAddress == "ann@example.com" && u.DisplayName == "Ann" && u.CreatedAt.Equal(registrationTime)
		})).
		Run(func(_ context.Context, user domain.User) { saved = user }).
		Return(nil).
		Once()
	queue.EXPECT().
		EnqueueWelcomeEmail(t.Context(), mock.MatchedBy(func(id uuid.UUID) bool { return id == saved.ID })).
		Return(nil).
		Once()

	got, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: " Ann@Example.COM ", DisplayName: " Ann "})

	require.NoError(t, err)
	assert.Equal(t, saved, got)
	assert.Equal(t, uuid.Version(7), got.ID.Version())
}

func TestRegisterUser_Execute_QueuesTheWelcomeEmailOnlyAfterTheUserIsSaved(t *testing.T) {
	t.Parallel()
	useCase, repository, queue, _ := newRegisterUser(t)
	var saved domain.User
	create := repository.EXPECT().
		Create(t.Context(), mock.MatchedBy(func(u domain.User) bool { return u.EmailAddress == "ann@example.com" })).
		Run(func(_ context.Context, user domain.User) { saved = user }).
		Return(nil).
		Once()
	queue.EXPECT().
		EnqueueWelcomeEmail(t.Context(), mock.MatchedBy(func(id uuid.UUID) bool { return id == saved.ID })).
		NotBefore(create).
		Return(nil).
		Once()

	_, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "Ann"})

	require.NoError(t, err)
}

func TestRegisterUser_Execute_ReturnsEmailTakenWhenTheRepositoryReportsADuplicate(t *testing.T) {
	t.Parallel()
	useCase, repository, _, _ := newRegisterUser(t) // The queue has no expected call, so a call fails the test.
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
			useCase, _, _, _ := newRegisterUser(t) // The mocks have no expected call, so a call fails the test.

			_, err := useCase.Execute(t.Context(), tt.input)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRegisterUser_Execute_ReturnsTheUnexpectedErrorOfTheRepository(t *testing.T) {
	t.Parallel()
	useCase, repository, _, _ := newRegisterUser(t) // The queue has no expected call, so a call fails the test.
	failure := errors.New("The connection is closed.")
	repository.EXPECT().Create(t.Context(), mock.MatchedBy(func(u domain.User) bool {
		return u.EmailAddress == "ann@example.com" && u.DisplayName == "Ann"
	})).Return(failure).Once()

	_, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "Ann"})

	require.ErrorIs(t, err, failure)
	require.NotErrorIs(t, err, domain.ErrEmailTaken)
}

func TestRegisterUser_Execute_ReturnsTheUserAndLogsTheErrorWhenTheQueueFails(t *testing.T) {
	t.Parallel()
	useCase, repository, queue, logs := newRegisterUser(t)
	var saved domain.User
	repository.EXPECT().Create(t.Context(), mock.MatchedBy(func(u domain.User) bool { return u.EmailAddress == "ann@example.com" })).
		Run(func(_ context.Context, user domain.User) { saved = user }).
		Return(nil).
		Once()
	queue.EXPECT().
		EnqueueWelcomeEmail(t.Context(), mock.MatchedBy(func(id uuid.UUID) bool { return id == saved.ID })).
		Return(errors.New("The Redis server is not reachable.")).
		Once()

	got, err := useCase.Execute(t.Context(), app.RegisterUserInput{EmailAddress: "ann@example.com", DisplayName: "Ann"})

	require.NoError(t, err)
	assert.Equal(t, saved, got)
	assert.Contains(t, logs.String(), `"level":"ERROR"`)
	assert.Contains(t, logs.String(), "The Redis server is not reachable.")
	assert.Contains(t, logs.String(), saved.ID.String())
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

func newSendWelcomeEmail(t *testing.T) (*app.SendWelcomeEmail, *mocks.MockUserRepository, *mocks.MockMailer) {
	t.Helper()
	repository := mocks.NewMockUserRepository(t)
	mailer := mocks.NewMockMailer(t)
	return app.NewSendWelcomeEmail(repository, mailer), repository, mailer
}

func TestSendWelcomeEmail_Execute_SendsTheEmailToTheUser(t *testing.T) {
	t.Parallel()
	useCase, repository, mailer := newSendWelcomeEmail(t)
	user := fixtures.User(fixtures.WithEmailAddress("ann@example.com"), fixtures.WithDisplayName("Ann"))
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	mailer.EXPECT().Send(t.Context(), welcomeEmailOfAnn).Return(nil).Once()

	err := useCase.Execute(t.Context(), user.ID)

	require.NoError(t, err)
}

func TestSendWelcomeEmail_Execute_ReturnsNotFoundAndSendsNothingWhenNoUserHasTheID(t *testing.T) {
	t.Parallel()
	useCase, repository, _ := newSendWelcomeEmail(t) // The mailer has no expected call, so a call fails the test.
	id := uuid.Must(uuid.NewV7())
	repository.EXPECT().FindByID(t.Context(), id).Return(domain.User{}, domain.ErrUserNotFound).Once()

	err := useCase.Execute(t.Context(), id)

	require.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestSendWelcomeEmail_Execute_ReturnsTheErrorOfTheMailer(t *testing.T) {
	t.Parallel()
	useCase, repository, mailer := newSendWelcomeEmail(t)
	user := fixtures.User(fixtures.WithEmailAddress("ann@example.com"), fixtures.WithDisplayName("Ann"))
	failure := errors.New("The mail server is not reachable.")
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	mailer.EXPECT().Send(t.Context(), welcomeEmailOfAnn).Return(failure).Once()

	err := useCase.Execute(t.Context(), user.ID)

	require.ErrorIs(t, err, failure)
}
