// Package jobs holds the handlers of the background tasks of the User context.
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// Handler runs the tasks of the User context.
type Handler struct {
	sendWelcomeEmail *app.SendWelcomeEmail
}

// NewHandler builds the handler.
func NewHandler(sendWelcomeEmail *app.SendWelcomeEmail) *Handler {
	return &Handler{sendWelcomeEmail: sendWelcomeEmail}
}

// Register adds the handlers of the context to the mux of the worker.
func (h *Handler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(app.WelcomeEmailTaskType, h.ProcessSendWelcomeEmail)
}

// ProcessSendWelcomeEmail sends the Welcome email of a task.
//
// A task that cannot succeed later, because its payload is not valid or its User does not exist,
// returns asynq.SkipRetry: the queue archives it at once. Any other error makes the queue try again.
func (h *Handler) ProcessSendWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var payload app.WelcomeEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return invalidPayload(err)
	}
	if payload.UserID == uuid.Nil {
		return invalidPayload(errors.New("The user ID is missing."))
	}
	err := h.sendWelcomeEmail.Execute(ctx, payload.UserID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("The User %s does not exist. %w", payload.UserID, asynq.SkipRetry)
	}
	return err
}

func invalidPayload(cause error) error {
	return fmt.Errorf("The payload of the task is not valid. Cause: %w", errors.Join(cause, asynq.SkipRetry))
}
