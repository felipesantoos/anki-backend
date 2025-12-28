package email

import (
	"context"
	"fmt"
	"net/url"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/email"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// EmailService implements IEmailService
type EmailService struct {
	emailRepo    secondary.IEmailRepository
	jwtService   *jwt.JWTService
	emailConfig  config.EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(
	emailRepo secondary.IEmailRepository,
	jwtService *jwt.JWTService,
	emailConfig config.EmailConfig,
) primary.IEmailService {
	return &EmailService{
		emailRepo:   emailRepo,
		jwtService:  jwtService,
		emailConfig: emailConfig,
	}
}

// SendVerificationEmail sends an email verification email to the user
func (s *EmailService) SendVerificationEmail(ctx context.Context, userID int64, userEmail string) error {
	// Generate verification token
	token, err := s.jwtService.GenerateEmailVerificationToken(userID)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Build verification URL
	verificationURL := s.buildVerificationURL(token)

	// Generate email content
	htmlBody := email.GenerateVerificationEmailHTML(verificationURL)
	textBody := email.GenerateVerificationEmailText(verificationURL)

	// Send email
	subject := "Verify Your Email - Anki Backend"
	err = s.emailRepo.SendEmail(ctx, userEmail, subject, htmlBody, textBody)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email to the user
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, userID int64, userEmail string, resetToken string) error {
	// Build reset URL
	resetURL := s.buildPasswordResetURL(resetToken)

	// Generate email content
	htmlBody := email.GeneratePasswordResetEmailHTML(resetURL)
	textBody := email.GeneratePasswordResetEmailText(resetURL)

	// Send email
	subject := "Reset Your Password - Anki Backend"
	err := s.emailRepo.SendEmail(ctx, userEmail, subject, htmlBody, textBody)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// buildVerificationURL builds the full verification URL with token
func (s *EmailService) buildVerificationURL(token string) string {
	baseURL := s.emailConfig.VerificationURL
	// Ensure base URL doesn't end with /
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	
	// Build URL with token query parameter
	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", baseURL, url.QueryEscape(token))
	return verificationURL
}

// buildPasswordResetURL builds the full password reset URL with token
func (s *EmailService) buildPasswordResetURL(token string) string {
	baseURL := s.emailConfig.VerificationURL
	// Ensure base URL doesn't end with /
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	
	// Build URL with token query parameter
	resetURL := fmt.Sprintf("%s/api/v1/auth/reset-password?token=%s", baseURL, url.QueryEscape(token))
	return resetURL
}

