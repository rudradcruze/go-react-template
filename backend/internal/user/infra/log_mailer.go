package infra

import (
	"context"
	"log/slog"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
)

var _ app.Mailer = (*LogMailer)(nil)

// LogMailer does not send an email. It writes the email to the log.
// Replace it with an adapter for a mail service when the application must send real emails.
type LogMailer struct {
	log *slog.Logger
}

// NewLogMailer builds the mailer.
func NewLogMailer(log *slog.Logger) *LogMailer {
	return &LogMailer{log: log}
}

// Send writes the email to the log.
func (m *LogMailer) Send(ctx context.Context, email app.Email) error {
	m.log.InfoContext(ctx, "The application writes an email to the log.",
		slog.String("to", string(email.To)), slog.String("subject", email.Subject), slog.String("body", email.Body))
	return nil
}
