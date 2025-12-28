package primary

import "context"

// IEmailService defines the interface for email operations
type IEmailService interface {
	// SendVerificationEmail sends an email verification email to the user
	// It generates a verification token and sends an email with a verification link
	SendVerificationEmail(ctx context.Context, userID int64, email string) error

	// SendPasswordResetEmail sends a password reset email to the user
	// This is reserved for future implementation
	SendPasswordResetEmail(ctx context.Context, userID int64, email string, resetToken string) error
}

