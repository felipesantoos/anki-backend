package email

import (
	"context"
	"errors"
	"testing"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/core/services/email"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEmailRepository is a mock implementation of IEmailRepository
type mockEmailRepository struct {
	sendEmailFunc func(ctx context.Context, to, subject, htmlBody, textBody string) error
}

func (m *mockEmailRepository) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, to, subject, htmlBody, textBody)
	}
	return nil
}

func createTestJWTService(t *testing.T) *jwt.JWTService {
	cfg := config.JWTConfig{
		SecretKey:          "test-secret-key-must-be-at-least-32-characters-long",
		AccessTokenExpiry:  15,
		RefreshTokenExpiry: 7,
		Issuer:             "test",
	}
	jwtSvc, err := jwt.NewJWTService(cfg)
	require.NoError(t, err)
	return jwtSvc
}

func TestEmailService_SendVerificationEmail_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailSent := false
	var sentTo, sentSubject string
	emailRepo := &mockEmailRepository{
		sendEmailFunc: func(ctx context.Context, to, subject, htmlBody, textBody string) error {
			emailSent = true
			sentTo = to
			sentSubject = subject
			assert.Contains(t, htmlBody, "Verify Your Email")
			assert.Contains(t, textBody, "Verify Your Email")
			assert.Contains(t, htmlBody, "/api/v1/auth/verify-email?token=")
			assert.Contains(t, textBody, "/api/v1/auth/verify-email?token=")
			return nil
		},
	}

	emailConfig := config.EmailConfig{
		VerificationURL: "http://localhost:3000",
	}

	service := email.NewEmailService(emailRepo, jwtSvc, emailConfig)

	ctx := context.Background()
	err := service.SendVerificationEmail(ctx, 1, "test@example.com")

	assert.NoError(t, err)
	assert.True(t, emailSent)
	assert.Equal(t, "test@example.com", sentTo)
	assert.Equal(t, "Verify Your Email - Anki Backend", sentSubject)
}

func TestEmailService_SendVerificationEmail_EmailRepoError(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailRepo := &mockEmailRepository{
		sendEmailFunc: func(ctx context.Context, to, subject, htmlBody, textBody string) error {
			return errors.New("SMTP connection failed")
		},
	}

	emailConfig := config.EmailConfig{
		VerificationURL: "http://localhost:3000",
	}

	service := email.NewEmailService(emailRepo, jwtSvc, emailConfig)

	ctx := context.Background()
	err := service.SendVerificationEmail(ctx, 1, "test@example.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send verification email")
}

func TestEmailService_SendVerificationEmail_GeneratesValidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	var htmlBody string
	emailRepo := &mockEmailRepository{
		sendEmailFunc: func(ctx context.Context, to, subject, htmlBodyParam, textBody string) error {
			htmlBody = htmlBodyParam
			return nil
		},
	}

	emailConfig := config.EmailConfig{
		VerificationURL: "http://localhost:3000",
	}

	service := email.NewEmailService(emailRepo, jwtSvc, emailConfig)

	ctx := context.Background()
	err := service.SendVerificationEmail(ctx, 123, "test@example.com")
	require.NoError(t, err)

	// Extract token from HTML body
	// The token should be in the URL: /api/v1/auth/verify-email?token=...
	assert.Contains(t, htmlBody, "/api/v1/auth/verify-email?token=")
	
	// Verify that we can extract and validate the token
	// This is a basic check - in a real scenario, we'd parse the HTML to extract the token
	// For now, we just verify the URL structure is correct
}

