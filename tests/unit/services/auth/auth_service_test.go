package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	domainEvents "github.com/felipesantos/anki-backend/core/domain/events"
)

// mockUserRepository is a mock implementation of IUserRepository
type mockUserRepository struct {
	saveFunc       func(ctx context.Context, user *entities.User) error
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
	existsByEmailFunc func(ctx context.Context, email string) (bool, error)
}

func (m *mockUserRepository) Save(ctx context.Context, user *entities.User) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFunc != nil {
		return m.existsByEmailFunc(ctx, email)
	}
	return false, nil
}

// mockDeckRepository is a mock implementation of IDeckRepository
type mockDeckRepository struct {
	createDefaultDeckFunc func(ctx context.Context, userID int64) (int64, error)
}

func (m *mockDeckRepository) CreateDefaultDeck(ctx context.Context, userID int64) (int64, error) {
	if m.createDefaultDeckFunc != nil {
		return m.createDefaultDeckFunc(ctx, userID)
	}
	return 1, nil
}

// mockEventBus is a mock implementation of IEventBus
type mockEventBus struct {
	publishFunc func(ctx context.Context, event domainEvents.DomainEvent) error
}

func (m *mockEventBus) Publish(ctx context.Context, event domainEvents.DomainEvent) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, event)
	}
	return nil
}

func (m *mockEventBus) Subscribe(eventType string, handler secondary.EventHandler) error {
	return nil
}

func (m *mockEventBus) Unsubscribe(eventType string, handlerID string) error {
	return nil
}

func (m *mockEventBus) Start() error {
	return nil
}

func (m *mockEventBus) Stop() error {
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email doesn't exist
		},
		saveFunc: func(ctx context.Context, user *entities.User) error {
			// Simulate setting ID after save
			user.ID = 1
			return nil
		},
	}

	deckRepo := &mockDeckRepository{
		createDefaultDeckFunc: func(ctx context.Context, userID int64) (int64, error) {
			return 1, nil
		},
	}

	eventBus := &mockEventBus{}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()
	user, err := service.Register(ctx, "user@example.com", "password123")

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if user == nil {
		t.Fatalf("Register() user = nil, want non-nil")
	}

	if user.ID == 0 {
		t.Errorf("Register() user.ID = 0, want non-zero")
	}

	if user.Email.Value() != "user@example.com" {
		t.Errorf("Register() user.Email = %v, want 'user@example.com'", user.Email.Value())
	}

	if user.EmailVerified {
		t.Errorf("Register() user.EmailVerified = true, want false")
	}
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()
	_, err := service.Register(ctx, "existing@example.com", "password123")

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if err != ErrEmailAlreadyExists {
		t.Errorf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()
	_, err := service.Register(ctx, "invalid-email", "password123")

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if err != ErrInvalidEmail && !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("Register() error = %v, want ErrInvalidEmail or wrapped", err)
	}
}

func TestAuthService_Register_InvalidPassword(t *testing.T) {
	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
	}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()

	tests := []struct {
		name     string
		password string
	}{
		{"too short", "pass1"},
		{"no letters", "12345678"},
		{"no numbers", "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(ctx, "user@example.com", tt.password)

			if err == nil {
				t.Fatalf("Register() expected error, got nil")
			}

			if err != ErrInvalidPassword && !errors.Is(err, ErrInvalidPassword) {
				t.Errorf("Register() error = %v, want ErrInvalidPassword or wrapped", err)
			}
		})
	}
}

func TestAuthService_Register_CreatesDefaultDeck(t *testing.T) {
	deckCreated := false
	var createdUserID int64

	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, user *entities.User) error {
			user.ID = 1
			createdUserID = user.ID
			return nil
		},
	}

	deckRepo := &mockDeckRepository{
		createDefaultDeckFunc: func(ctx context.Context, userID int64) (int64, error) {
			if userID == createdUserID {
				deckCreated = true
			}
			return 1, nil
		},
	}

	eventBus := &mockEventBus{}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()
	_, err := service.Register(ctx, "user@example.com", "password123")

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if !deckCreated {
		t.Errorf("Register() should create default deck for user")
	}
}

func TestAuthService_Register_PublishesEvent(t *testing.T) {
	eventPublished := false

	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		saveFunc: func(ctx context.Context, user *entities.User) error {
			user.ID = 1
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{
		publishFunc: func(ctx context.Context, event domainEvents.DomainEvent) error {
			if event != nil && event.EventType() == domainEvents.UserRegisteredEventType {
				eventPublished = true
			}
			return nil
		},
	}

	service := NewAuthService(userRepo, deckRepo, eventBus)

	ctx := context.Background()
	_, err := service.Register(ctx, "user@example.com", "password123")

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if !eventPublished {
		t.Errorf("Register() should publish UserRegistered event")
	}
}
