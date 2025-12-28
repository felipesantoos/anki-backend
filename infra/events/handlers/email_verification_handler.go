package handlers

import (
	"context"
	"log/slog"

	"github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// EmailVerificationHandler handles UserRegistered events by sending verification emails
type EmailVerificationHandler struct {
	*infraEvents.BaseEventHandler
	logger       *slog.Logger
	emailService primary.IEmailService
}

// NewEmailVerificationHandler creates a new email verification event handler
func NewEmailVerificationHandler(emailService primary.IEmailService) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		BaseEventHandler: infraEvents.NewBaseEventHandler(events.UserRegisteredEventType),
		logger:           logger.GetLogger(),
		emailService:     emailService,
	}
}

// Handle processes the UserRegistered event and sends a verification email
func (h *EmailVerificationHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	// Type assert to UserRegistered event
	userRegistered, ok := event.(*events.UserRegistered)
	if !ok {
		h.logger.Warn("EmailVerificationHandler received non-UserRegistered event",
			"event_type", event.EventType(),
		)
		return nil
	}

	h.logger.Info("Processing UserRegistered event for email verification",
		"user_id", userRegistered.UserID,
		"email", userRegistered.Email,
	)

	// Send verification email
	err := h.emailService.SendVerificationEmail(ctx, userRegistered.UserID, userRegistered.Email)
	if err != nil {
		h.logger.Error("Failed to send verification email",
			"user_id", userRegistered.UserID,
			"email", userRegistered.Email,
			"error", err,
		)
		return err
	}

	h.logger.Info("Verification email sent successfully",
		"user_id", userRegistered.UserID,
		"email", userRegistered.Email,
	)

	return nil
}

