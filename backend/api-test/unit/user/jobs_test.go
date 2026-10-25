package user_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/fixtures"
	"github.com/rudradcruze/go-react-template/backend/api-test/mocks"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
	"github.com/rudradcruze/go-react-template/backend/internal/user/jobs"
)

func newJobHandler(t *testing.T) (*jobs.Handler, *mocks.MockUserRepository, *mocks.MockMailer) {
	t.Helper()
	repository := mocks.NewMockUserRepository(t)
	mailer := mocks.NewMockMailer(t)
	return jobs.NewHandler(app.NewSendWelcomeEmail(repository, mailer)), repository, mailer
}

func welcomeEmailTask(t *testing.T, userID uuid.UUID) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(app.WelcomeEmailPayload{UserID: userID})
	require.NoError(t, err)
	return asynq.NewTask(app.WelcomeEmailTaskType, payload)
}

func TestJobHandler_ProcessSendWelcomeEmail_SendsTheEmail(t *testing.T) {
	t.Parallel()
	handler, repository, mailer := newJobHandler(t)
	user := fixtures.User(fixtures.WithEmailAddress("ann@example.com"), fixtures.WithDisplayName("Ann"))
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	mailer.EXPECT().Send(t.Context(), welcomeEmailOfAnn).Return(nil).Once()

	err := handler.ProcessSendWelcomeEmail(t.Context(), welcomeEmailTask(t, user.ID))

	require.NoError(t, err)
}

func TestJobHandler_ProcessSendWelcomeEmail_DoesNotTryAgainWhenThePayloadIsNotValid(t *testing.T) {
	t.Parallel()
	payloads := map[string]string{"not JSON": `not json`, "no user ID": `{}`, "ID is not a UUID": `{"userId":"abc"}`}
	for name, payload := range payloads {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			handler, _, _ := newJobHandler(t) // The mocks have no expected call, so a call fails the test.

			err := handler.ProcessSendWelcomeEmail(t.Context(), asynq.NewTask(app.WelcomeEmailTaskType, []byte(payload)))

			require.ErrorIs(t, err, asynq.SkipRetry)
		})
	}
}

func TestJobHandler_ProcessSendWelcomeEmail_DoesNotTryAgainWhenNoUserHasTheID(t *testing.T) {
	t.Parallel()
	handler, repository, _ := newJobHandler(t)
	id := uuid.Must(uuid.NewV7())
	repository.EXPECT().FindByID(t.Context(), id).Return(domain.User{}, domain.ErrUserNotFound).Once()

	err := handler.ProcessSendWelcomeEmail(t.Context(), welcomeEmailTask(t, id))

	require.ErrorIs(t, err, asynq.SkipRetry)
}

func TestJobHandler_ProcessSendWelcomeEmail_TriesAgainWhenTheMailerFails(t *testing.T) {
	t.Parallel()
	handler, repository, mailer := newJobHandler(t)
	user := fixtures.User(fixtures.WithEmailAddress("ann@example.com"), fixtures.WithDisplayName("Ann"))
	failure := errors.New("The mail server is not reachable.")
	repository.EXPECT().FindByID(t.Context(), user.ID).Return(user, nil).Once()
	mailer.EXPECT().Send(t.Context(), welcomeEmailOfAnn).Return(failure).Once()

	err := handler.ProcessSendWelcomeEmail(t.Context(), welcomeEmailTask(t, user.ID))

	require.ErrorIs(t, err, failure)
	assert.NotErrorIs(t, err, asynq.SkipRetry)
}
