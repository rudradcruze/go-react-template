package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
)

var _ app.WelcomeEmailQueue = (*AsynqWelcomeEmailQueue)(nil)

// AsynqWelcomeEmailQueue queues the Welcome email as an Asynq task.
type AsynqWelcomeEmailQueue struct {
	client *asynq.Client
}

// NewWelcomeEmailQueue builds the queue on an Asynq client.
func NewWelcomeEmailQueue(client *asynq.Client) *AsynqWelcomeEmailQueue {
	return &AsynqWelcomeEmailQueue{client: client}
}

// enqueueTimeout is the longest wait for the queue. Redis can refuse or not answer, and the Redis client tries again.
const enqueueTimeout = 2 * time.Second

// EnqueueWelcomeEmail adds the task. The queue tries again up to app.WelcomeEmailMaxRetry times after a failure.
//
// The call keeps the values of ctx but not its cancellation. The database already saved the User, so a client
// that disconnects must not cancel the task. The call waits at most enqueueTimeout.
func (q *AsynqWelcomeEmailQueue) EnqueueWelcomeEmail(ctx context.Context, userID uuid.UUID) error {
	payload, err := json.Marshal(app.WelcomeEmailPayload{UserID: userID})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), enqueueTimeout)
	defer cancel()
	task := asynq.NewTask(app.WelcomeEmailTaskType, payload)
	_, err = q.client.EnqueueContext(ctx, task, asynq.MaxRetry(app.WelcomeEmailMaxRetry), asynq.Timeout(app.WelcomeEmailTimeout))
	return err
}
