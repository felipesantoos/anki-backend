package secondary

import "context"

// IEmailRepository defines the interface for email sending operations
type IEmailRepository interface {
	// SendEmail sends an email with HTML and plain text content
	// to: recipient email address
	// subject: email subject
	// htmlBody: HTML content of the email
	// textBody: plain text content of the email
	SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error
}

