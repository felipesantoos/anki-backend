package email

import (
	"context"
	"log/slog"
)

// ConsoleRepository implements IEmailRepository by logging emails to console
// This is useful for development and testing when SMTP is not configured
type ConsoleRepository struct {
	logger *slog.Logger
}

// NewConsoleRepository creates a new console email repository
func NewConsoleRepository(logger *slog.Logger) *ConsoleRepository {
	return &ConsoleRepository{
		logger: logger,
	}
}

// SendEmail logs the email to console instead of actually sending it
func (r *ConsoleRepository) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	r.logger.Info("Email would be sent (console mode)",
		"to", to,
		"subject", subject,
		"html_body_length", len(htmlBody),
		"text_body_length", len(textBody),
	)
	r.logger.Debug("Email content",
		"to", to,
		"subject", subject,
		"text_body", textBody,
		"html_body", htmlBody,
	)
	return nil
}

