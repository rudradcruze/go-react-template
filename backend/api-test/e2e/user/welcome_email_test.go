//go:build e2e

package user_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testapi"
	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testenv"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/queue"
	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/infra"
)

// pendingTasks gives the tasks that wait in the default queue.
func pendingTasks(t *testing.T, api *testapi.API) []*asynq.TaskInfo {
	t.Helper()
	tasks, err := api.QueueInspector.ListPendingTasks("default")
	if errors.Is(err, asynq.ErrQueueNotFound) {
		return nil
	}
	require.NoError(t, err)
	return tasks
}

// registerAnn registers Ann and returns the only task in the queue.
func registerAnn(t *testing.T, api *testapi.API) (userBody, *asynq.TaskInfo) {
	t.Helper()
	res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`)
	require.Equal(t, http.StatusCreated, res.StatusCode, string(res.Body))
	tasks := pendingTasks(t, api)
	require.Len(t, tasks, 1)
	return decode[userBody](t, res), tasks[0]
}

func TestAPI_RegisterUser_QueuesOneWelcomeEmailTaskForTheUser(t *testing.T) {
	api := testapi.Start(t)

	registered, task := registerAnn(t, api)

	assert.Equal(t, "user:send_welcome_email", task.Type)
	var payload app.WelcomeEmailPayload
	require.NoError(t, json.Unmarshal(task.Payload, &payload))
	assert.Equal(t, registered.ID, payload.UserID.String())
}

func TestAPI_RegisterUser_GivesTheWelcomeEmailTaskThreeRetriesAndATimeout(t *testing.T) {
	api := testapi.Start(t)

	_, task := registerAnn(t, api)

	assert.Equal(t, 3, task.MaxRetry)
	assert.Equal(t, time.Minute, task.Timeout)
}

func TestAPI_RegisterUser_QueuesTheWelcomeEmailInRedisDatabase1(t *testing.T) {
	api := testapi.Start(t)

	registerAnn(t, api) // The inspector reads database 1. The cache database, database 0, has no queue keys.

	keys, err := api.Redis.Keys(t.Context(), "asynq:*").Result()
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestAPI_RegisterUser_QueuesNoWelcomeEmailWhenTheEmailAddressIsTaken(t *testing.T) {
	api := testapi.Start(t)
	body := `{"emailAddress":"ann@example.com","displayName":"Ann"}`
	require.Equal(t, http.StatusCreated, api.PostJSON(t, "/api/v1/users", body).StatusCode)

	res := api.PostJSON(t, "/api/v1/users", body)

	require.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Len(t, pendingTasks(t, api), 1)
}

func TestAPI_RegisterUser_Returns201AndLogsTheErrorWhenTheQueueFails(t *testing.T) {
	api := testapi.Start(t)
	api.BreakQueue(t)

	res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`)

	require.Equal(t, http.StatusCreated, res.StatusCode, string(res.Body))
	assert.Contains(t, api.Logs(), "The application did not queue the Welcome email.")
	assert.Equal(t, 1, countUsers(t, api))
}

func TestWorker_ProcessTask_SendsTheWelcomeEmailOfARegistration(t *testing.T) {
	api := testapi.Start(t)
	api.StartWorker(t)

	res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`)

	require.Equal(t, http.StatusCreated, res.StatusCode, string(res.Body))
	require.Eventually(t, func() bool { return strings.Contains(api.Logs(), "Welcome to go-react-template") },
		10*time.Second, 100*time.Millisecond, "The worker did not send the email. Logs: %s", api.Logs())
	assert.Contains(t, api.Logs(), `"to":"ann@example.com"`)
	assert.Contains(t, api.Logs(), "Hello Ann.")
	require.Eventually(t, func() bool { return len(pendingTasks(t, api)) == 0 }, 5*time.Second, 100*time.Millisecond)
}

func TestWorker_ProcessTask_ArchivesAtOnceATaskForAUserThatDoesNotExist(t *testing.T) {
	api := testapi.Start(t)
	api.StartWorker(t)
	payload, err := json.Marshal(app.WelcomeEmailPayload{UserID: uuid.Must(uuid.NewV7())})
	require.NoError(t, err)

	api.Enqueue(t, asynq.NewTask(app.WelcomeEmailTaskType, payload), asynq.MaxRetry(app.WelcomeEmailMaxRetry))

	require.Eventually(t, func() bool {
		archived, err := api.QueueInspector.ListArchivedTasks("default")
		return err == nil && len(archived) == 1
	}, 10*time.Second, 100*time.Millisecond, "The queue did not archive the task. Logs: %s", api.Logs())
	assert.NotContains(t, api.Logs(), "Welcome to go-react-template")
	// The worker logs each failed task, also a task that the queue does not try again.
	assert.Contains(t, api.Logs(), "The worker did not complete the task.")
	assert.Contains(t, api.Logs(), "does not exist")
}

func TestWelcomeEmailQueue_Enqueue_StillQueuesTheTaskWhenTheRequestIsCancelled(t *testing.T) {
	res := testenv.Begin(t)
	client, err := queue.Open(res.Redis)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	inspector := asynq.NewInspector(queue.WorkerOpt(res.Redis))
	t.Cleanup(func() { _ = inspector.Close() })
	// The client of a request can disconnect after the database saved the User. The email must not be lost.
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()

	err = infra.NewWelcomeEmailQueue(client).EnqueueWelcomeEmail(cancelled, uuid.Must(uuid.NewV7()))

	require.NoError(t, err)
	tasks, err := inspector.ListPendingTasks("default")
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestWelcomeEmailQueue_Enqueue_FailsInUnder3SecondsWhenRedisDoesNotAnswer(t *testing.T) {
	// A server that accepts connections and never answers.
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { _ = conn.Close() })
		}
	}()
	cfg := testenv.Begin(t).Redis
	cfg.Host, cfg.Port = "127.0.0.1", listener.Addr().(*net.TCPAddr).Port
	client := asynq.NewClient(queue.ClientOpt(cfg)) // queue.Open would fail at its check of the connection.
	t.Cleanup(func() { _ = client.Close() })

	start := time.Now()
	err = infra.NewWelcomeEmailQueue(client).EnqueueWelcomeEmail(t.Context(), uuid.Must(uuid.NewV7()))

	require.Error(t, err)
	assert.Less(t, time.Since(start), 3*time.Second)
}
