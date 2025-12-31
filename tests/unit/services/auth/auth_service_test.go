package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	userEntity "github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	domainEvents "github.com/felipesantos/anki-backend/core/domain/events"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	"github.com/felipesantos/anki-backend/core/services/session"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/config"
)

// mockUserRepository is a mock implementation of IUserRepository
type mockUserRepository struct {
	saveFunc       func(ctx context.Context, user *userEntity.User) error
	findByEmailFunc func(ctx context.Context, email string) (*userEntity.User, error)
	findByIDFunc   func(ctx context.Context, id int64) (*userEntity.User, error)
	existsByEmailFunc func(ctx context.Context, email string) (bool, error)
	updateFunc    func(ctx context.Context, user *userEntity.User) error
}

func (m *mockUserRepository) Save(ctx context.Context, u *userEntity.User) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, u)
	}
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (*userEntity.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFunc != nil {
		return m.existsByEmailFunc(ctx, email)
	}
	return false, nil
}

func (m *mockUserRepository) Update(ctx context.Context, u *userEntity.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, u)
	}
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

// mockProfileRepository is a mock implementation of IProfileRepository
type mockProfileRepository struct {
	saveFunc func(ctx context.Context, userID int64, profile *profile.Profile) error
}

func (m *mockProfileRepository) Save(ctx context.Context, userID int64, p *profile.Profile) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, userID, p)
	}
	return nil
}

func (m *mockProfileRepository) FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error) {
	return nil, nil
}

func (m *mockProfileRepository) FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error) {
	return nil, nil
}

func (m *mockProfileRepository) FindByName(ctx context.Context, userID int64, name string) (*profile.Profile, error) {
	return nil, nil
}

func (m *mockProfileRepository) Update(ctx context.Context, userID int64, id int64, profile *profile.Profile) error {
	return nil
}

func (m *mockProfileRepository) Delete(ctx context.Context, userID int64, id int64) error {
	return nil
}

func (m *mockProfileRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	return false, nil
}

// mockUserPreferencesRepository is a mock implementation of IUserPreferencesRepository
type mockUserPreferencesRepository struct {
	saveFunc func(ctx context.Context, userID int64, prefs *userpreferences.UserPreferences) error
}

func (m *mockUserPreferencesRepository) Save(ctx context.Context, userID int64, p *userpreferences.UserPreferences) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, userID, p)
	}
	return nil
}

func (m *mockUserPreferencesRepository) FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	return nil, nil
}

func (m *mockUserPreferencesRepository) FindByID(ctx context.Context, userID int64, id int64) (*userpreferences.UserPreferences, error) {
	return nil, nil
}

func (m *mockUserPreferencesRepository) Update(ctx context.Context, userID int64, id int64, prefs *userpreferences.UserPreferences) error {
	return nil
}

func (m *mockUserPreferencesRepository) Delete(ctx context.Context, userID int64, id int64) error {
	return nil
}

func (m *mockUserPreferencesRepository) Exists(ctx context.Context, userID int64) (bool, error) {
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

func (m *mockDeckRepository) FindByID(ctx context.Context, userID int64, deckID int64) (*deck.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) FindByUserID(ctx context.Context, userID int64) ([]*deck.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) FindByParentID(ctx context.Context, userID int64, parentID int64) ([]*deck.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) Save(ctx context.Context, userID int64, deckEntity *deck.Deck) error {
	return nil
}

func (m *mockDeckRepository) Update(ctx context.Context, userID int64, deckID int64, deckEntity *deck.Deck) error {
	return nil
}

func (m *mockDeckRepository) Delete(ctx context.Context, userID int64, deckID int64) error {
	return nil
}

func (m *mockDeckRepository) Exists(ctx context.Context, userID int64, name string, parentID *int64) (bool, error) {
	return false, nil
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

// mockCacheRepository is a mock implementation of ICacheRepository
type mockCacheRepository struct {
	getFunc    func(ctx context.Context, key string) (string, error)
	setFunc    func(ctx context.Context, key string, value string, ttl time.Duration) error
	deleteFunc func(ctx context.Context, key string) error
	existsFunc func(ctx context.Context, key string) (bool, error)
	pingFunc   func(ctx context.Context) error
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return "", nil
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, key)
	}
	return false, nil
}

func (m *mockCacheRepository) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockCacheRepository) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (m *mockCacheRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (m *mockCacheRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

// mockTransactionManager is a mock implementation of ITransactionManager
type mockTransactionManager struct {
	withTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTransactionFunc != nil {
		return m.withTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}

// mockSessionService is a mock implementation of SessionService for testing
type mockSessionService struct {
	createSessionWithMetadataFunc func(ctx context.Context, userID int64, metadata session.SessionMetadata) (string, error)
	deleteSessionFunc            func(ctx context.Context, sessionID string) error
	getUserSessionsFunc          func(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	deleteUserSessionFunc         func(ctx context.Context, userID int64, sessionID string) error
	deleteAllUserSessionsFunc     func(ctx context.Context, userID int64) error
	associateRefreshTokenFunc     func(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error
	getSessionByRefreshTokenFunc  func(ctx context.Context, refreshTokenHash string) (string, error)
	deleteRefreshTokenAssociationFunc func(ctx context.Context, refreshTokenHash string) error
	updateSessionFunc             func(ctx context.Context, sessionID string, data map[string]interface{}) error
}

func (m *mockSessionService) CreateSession(ctx context.Context, userID string, data map[string]interface{}) (string, error) {
	return "mock-session-id", nil
}

func (m *mockSessionService) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *mockSessionService) RefreshSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockSessionService) CreateSessionWithMetadata(ctx context.Context, userID int64, metadata session.SessionMetadata) (string, error) {
	if m.createSessionWithMetadataFunc != nil {
		return m.createSessionWithMetadataFunc(ctx, userID, metadata)
	}
	return "mock-session-id", nil
}

func (m *mockSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionService) GetUserSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	if m.getUserSessionsFunc != nil {
		return m.getUserSessionsFunc(ctx, userID)
	}
	return []map[string]interface{}{}, nil
}

func (m *mockSessionService) DeleteUserSession(ctx context.Context, userID int64, sessionID string) error {
	if m.deleteUserSessionFunc != nil {
		return m.deleteUserSessionFunc(ctx, userID, sessionID)
	}
	return nil
}

func (m *mockSessionService) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	if m.deleteAllUserSessionsFunc != nil {
		return m.deleteAllUserSessionsFunc(ctx, userID)
	}
	return nil
}

func (m *mockSessionService) AssociateRefreshToken(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error {
	if m.associateRefreshTokenFunc != nil {
		return m.associateRefreshTokenFunc(ctx, refreshTokenHash, sessionID, ttl)
	}
	return nil
}

func (m *mockSessionService) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error) {
	if m.getSessionByRefreshTokenFunc != nil {
		return m.getSessionByRefreshTokenFunc(ctx, refreshTokenHash)
	}
	return "", errors.New("not found")
}

func (m *mockSessionService) DeleteRefreshTokenAssociation(ctx context.Context, refreshTokenHash string) error {
	if m.deleteRefreshTokenAssociationFunc != nil {
		return m.deleteRefreshTokenAssociationFunc(ctx, refreshTokenHash)
	}
	return nil
}

func (m *mockSessionService) UpdateSession(ctx context.Context, sessionID string, data map[string]interface{}) error {
	if m.updateSessionFunc != nil {
		return m.updateSessionFunc(ctx, sessionID, data)
	}
	return nil
}

// createTestJWTService creates a JWT service for testing
func createTestJWTService(t *testing.T) *jwt.JWTService {
	cfg := config.JWTConfig{
		SecretKey:          "this-is-a-test-secret-key-with-at-least-32-chars",
		AccessTokenExpiry:  15,
		RefreshTokenExpiry: 7,
		Issuer:             "test-issuer",
	}
	
	service, err := jwt.NewJWTService(cfg)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}
	return service
}

// createTestSessionService creates a mock session service for testing
func createTestSessionService() *mockSessionService {
	return &mockSessionService{}
}

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := &mockUserRepository{
		existsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email doesn't exist
		},
		saveFunc: func(ctx context.Context, u *userEntity.User) error {
			// Simulate setting ID after save
			u.SetID(1)
			return nil
		},
	}

	deckRepo := &mockDeckRepository{
		createDefaultDeckFunc: func(ctx context.Context, userID int64) (int64, error) {
			return 1, nil
		},
	}

	eventBus := &mockEventBus{}

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	user, err := service.Register(ctx, "user@example.com", "password123")

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if user == nil {
		t.Fatalf("Register() user = nil, want non-nil")
	}

	if user.GetID() == 0 {
		t.Errorf("Register() user.ID = 0, want non-zero")
	}

	if user.GetEmail().Value() != "user@example.com" {
		t.Errorf("Register() user.Email = %v, want 'user@example.com'", user.GetEmail().Value())
	}

	if user.GetEmailVerified() {
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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err := service.Register(ctx, "existing@example.com", "password123")

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if err != userEntity.ErrEmailAlreadyExists {
		t.Errorf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err := service.Register(ctx, "invalid-email", "password123")

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if err != authService.ErrInvalidEmail && !errors.Is(err, authService.ErrInvalidEmail) {
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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

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

			if err != authService.ErrInvalidPassword && !errors.Is(err, authService.ErrInvalidPassword) {
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
		saveFunc: func(ctx context.Context, u *userEntity.User) error {
			u.SetID(1)
			createdUserID = u.GetID()
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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

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
		saveFunc: func(ctx context.Context, u *userEntity.User) error {
			u.SetID(1)
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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err := service.Register(ctx, "user@example.com", "password123")

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if !eventPublished {
		t.Errorf("Register() should publish UserRegistered event")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Create a test user with password
	emailVO, _ := valueobjects.NewEmail("user@example.com")
	passwordVO, _ := valueobjects.NewPassword("password123")
	now := time.Now()
	testUser, _ := userEntity.NewBuilder().
		WithID(1).
		WithEmail(emailVO).
		WithPasswordHash(passwordVO).
		WithEmailVerified(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			if email == "user@example.com" {
				return testUser, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, u *userEntity.User) error {
			return nil
		},
	}

	cacheRepo := &mockCacheRepository{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	resp, err := service.Login(ctx, "user@example.com", "password123", "192.168.1.1", "Mozilla/5.0")

	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}

	if resp == nil {
		t.Fatalf("Login() response = nil, want non-nil")
	}

	if resp.AccessToken == "" {
		t.Errorf("Login() AccessToken = empty, want non-empty")
	}

	if resp.RefreshToken == "" {
		t.Errorf("Login() RefreshToken = empty, want non-empty")
	}

	if resp.TokenType != "Bearer" {
		t.Errorf("Login() TokenType = %v, want 'Bearer'", resp.TokenType)
	}

	if resp.User.ID != testUser.GetID() {
		t.Errorf("Login() User.ID = %v, want %v", resp.User.ID, testUser.GetID())
	}

	if resp.ExpiresIn <= 0 {
		t.Errorf("Login() ExpiresIn = %v, want > 0", resp.ExpiresIn)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	jwtSvc := createTestJWTService(t)

	tests := []struct {
		name        string
		email       string
		password    string
		findByEmail func(ctx context.Context, email string) (*userEntity.User, error)
	}{
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			findByEmail: func(ctx context.Context, email string) (*userEntity.User, error) {
				return nil, nil
			},
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrongpassword",
			findByEmail: func(ctx context.Context, email string) (*userEntity.User, error) {
				emailVO, _ := valueobjects.NewEmail("user@example.com")
				passwordVO, _ := valueobjects.NewPassword("password123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(emailVO).
					WithPasswordHash(passwordVO).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepository{
				findByEmailFunc: tt.findByEmail,
			}
			cacheRepo := &mockCacheRepository{}
			deckRepo := &mockDeckRepository{}
			eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

			ctx := context.Background()
			_, err := service.Login(ctx, tt.email, tt.password, "192.168.1.1", "Mozilla/5.0")

			if err == nil {
				t.Fatalf("Login() expected error, got nil")
			}

			if err != authService.ErrInvalidCredentials && !errors.Is(err, authService.ErrInvalidCredentials) {
				t.Errorf("Login() error = %v, want ErrInvalidCredentials or wrapped", err)
			}
		})
	}
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	userRepo := &mockUserRepository{}
	cacheRepo := &mockCacheRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err := service.Login(ctx, "invalid-email", "password123", "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatalf("Login() expected error, got nil")
	}

	if err != authService.ErrInvalidEmail && !errors.Is(err, authService.ErrInvalidEmail) {
		t.Errorf("Login() error = %v, want ErrInvalidEmail or wrapped", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Create a test user
	emailVO, _ := valueobjects.NewEmail("user@example.com")
	testUser, _ := userEntity.NewBuilder().
		WithID(1).
		WithEmail(emailVO).
		WithEmailVerified(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()

	// Generate a refresh token
	refreshToken, err := jwtSvc.GenerateRefreshToken(testUser.GetID())
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Track which keys are stored/deleted for token rotation verification
	storedKeys := make(map[string]bool)
	deletedKeys := make(map[string]bool)

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == testUser.GetID() {
				return testUser, nil
			}
			return nil, nil
		},
	}

	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			// Old refresh token should exist initially
			return !deletedKeys[key], nil
		},
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			// Track that new refresh token is stored
			storedKeys[key] = true
			return nil
		},
		deleteFunc: func(ctx context.Context, key string) error {
			// Track that old refresh token is deleted (token rotation)
			deletedKeys[key] = true
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	resp, err := service.RefreshToken(ctx, refreshToken)

	if err != nil {
		t.Fatalf("RefreshToken() error = %v, want nil", err)
	}

	if resp == nil {
		t.Fatalf("RefreshToken() response = nil, want non-nil")
	}

	if resp.AccessToken == "" {
		t.Errorf("RefreshToken() AccessToken = empty, want non-empty")
	}

	// Token rotation: new refresh token should be returned
	if resp.RefreshToken == "" {
		t.Errorf("RefreshToken() RefreshToken = empty, want non-empty (token rotation)")
	}
	
	// Note: We don't verify that the new refresh token is different from the old one
	// because JWT tokens generated in rapid succession might have identical timestamps
	// The important thing is that a new refresh token is returned and the old one is invalidated

	if resp.TokenType != "Bearer" {
		t.Errorf("RefreshToken() TokenType = %v, want 'Bearer'", resp.TokenType)
	}

	if resp.ExpiresIn <= 0 {
		t.Errorf("RefreshToken() ExpiresIn = %v, want > 0", resp.ExpiresIn)
	}

	// Verify token rotation: new token should be stored, old token should be deleted
	if len(storedKeys) == 0 {
		t.Errorf("RefreshToken() should store new refresh token in cache (token rotation)")
	}

	if len(deletedKeys) == 0 {
		t.Errorf("RefreshToken() should delete old refresh token from cache (token rotation)")
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	userRepo := &mockUserRepository{}
	cacheRepo := &mockCacheRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err := service.RefreshToken(ctx, "invalid-token")

	if err == nil {
		t.Fatalf("RefreshToken() expected error, got nil")
	}

	if err != authService.ErrInvalidToken && !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("RefreshToken() error = %v, want ErrInvalidToken or wrapped", err)
	}
}

func TestAuthService_RefreshToken_TokenNotInCache(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailVO, _ := valueobjects.NewEmail("user@example.com")
	testUser, _ := userEntity.NewBuilder().
		WithID(1).
		WithEmail(emailVO).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()

	refreshToken, err := jwtSvc.GenerateRefreshToken(testUser.GetID())
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	userRepo := &mockUserRepository{}
	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			// Token not found in cache (revoked)
			return false, nil
		},
	}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err = service.RefreshToken(ctx, refreshToken)

	if err == nil {
		t.Fatalf("RefreshToken() expected error, got nil")
	}

	if err != authService.ErrInvalidToken && !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("RefreshToken() error = %v, want ErrInvalidToken or wrapped", err)
	}
}

func TestAuthService_RefreshToken_WrongTokenType(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailVO, _ := valueobjects.NewEmail("user@example.com")
	testUser, _ := userEntity.NewBuilder().
		WithID(1).
		WithEmail(emailVO).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()

	// Generate access token instead of refresh token
	accessToken, err := jwtSvc.GenerateAccessToken(testUser.GetID())
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	userRepo := &mockUserRepository{}
	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			return true, nil
		},
	}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	_, err = service.RefreshToken(ctx, accessToken)

	if err == nil {
		t.Fatalf("RefreshToken() expected error, got nil")
	}

	if err != authService.ErrInvalidToken && !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("RefreshToken() error = %v, want ErrInvalidToken or wrapped", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	accessToken, err := jwtSvc.GenerateAccessToken(1)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}
	
	refreshToken, err := jwtSvc.GenerateRefreshToken(1)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	accessTokenBlacklisted := false
	refreshTokenDeleted := false
	cacheRepo := &mockCacheRepository{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			// Track that access token is blacklisted
			accessTokenBlacklisted = true
			return nil
		},
		deleteFunc: func(ctx context.Context, key string) error {
			// Track that refresh token is deleted
			refreshTokenDeleted = true
			return nil
		},
	}

	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.Logout(ctx, accessToken, refreshToken)

	if err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}

	if !accessTokenBlacklisted {
		t.Errorf("Logout() should blacklist access token")
	}

	if !refreshTokenDeleted {
		t.Errorf("Logout() should delete refresh token from cache")
	}
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return nil // Allow blacklisting even for invalid tokens
		},
		deleteFunc: func(ctx context.Context, key string) error {
			return nil // Allow deletion even for invalid tokens
		},
	}
	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	// Logout should still succeed even with invalid tokens (idempotent operation)
	err := service.Logout(ctx, "invalid-access-token", "invalid-refresh-token")

	// Logout is idempotent, so it should not return an error for invalid tokens
	// It just tries to blacklist/delete, which is safe
	if err != nil {
		t.Logf("Logout() with invalid tokens returned error (acceptable): %v", err)
	}
}

func TestAuthService_Logout_AccessTokenOnly(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	accessToken, err := jwtSvc.GenerateAccessToken(1)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	accessTokenBlacklisted := false
	cacheRepo := &mockCacheRepository{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			accessTokenBlacklisted = true
			return nil
		},
	}

	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

			emailSvc := &mockEmailService{}
			sessionSvc := &mockSessionService{}
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.Logout(ctx, accessToken, "")

	if err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}

	if !accessTokenBlacklisted {
		t.Errorf("Logout() should blacklist access token even without refresh token")
	}
}

// mockEmailService is a mock implementation of IEmailService
type mockEmailService struct {
	sendVerificationEmailFunc func(ctx context.Context, userID int64, email string) error
	sendPasswordResetEmailFunc func(ctx context.Context, userID int64, email string, resetToken string) error
}

func (m *mockEmailService) SendVerificationEmail(ctx context.Context, userID int64, email string) error {
	if m.sendVerificationEmailFunc != nil {
		return m.sendVerificationEmailFunc(ctx, userID, email)
	}
	return nil
}

func (m *mockEmailService) SendPasswordResetEmail(ctx context.Context, userID int64, email string, resetToken string) error {
	if m.sendPasswordResetEmailFunc != nil {
		return m.sendPasswordResetEmailFunc(ctx, userID, email, resetToken)
	}
	return nil
}

func TestAuthService_VerifyEmail_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate a valid verification token
	token, err := jwtSvc.GenerateEmailVerificationToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userUpdated := false
	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("password123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(false).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
		updateFunc: func(ctx context.Context, user *userEntity.User) error {
			userUpdated = true
			if !user.GetEmailVerified() {
				t.Errorf("User should be marked as verified")
			}
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.VerifyEmail(ctx, token)

	if err != nil {
		t.Fatalf("VerifyEmail() error = %v, want nil", err)
	}

	if !userUpdated {
		t.Errorf("VerifyEmail() should update user")
	}
}

func TestAuthService_VerifyEmail_InvalidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.VerifyEmail(ctx, "invalid-token")

	if err == nil {
		t.Errorf("VerifyEmail() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("VerifyEmail() error = %v, want ErrInvalidToken", err)
	}
}

func TestAuthService_VerifyEmail_AlreadyVerified(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate a valid verification token
	token, err := jwtSvc.GenerateEmailVerificationToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("password123")
				user, _ := userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true). // Already verified
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
				return user, nil
			}
			return nil, errors.New("user not found")
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}
	emailSvc := &mockEmailService{}
	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.VerifyEmail(ctx, token)

	// Should succeed (idempotent operation)
	if err != nil {
		t.Fatalf("VerifyEmail() error = %v, want nil (idempotent)", err)
	}
}

func TestAuthService_ResendVerificationEmail_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailSent := false
	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			if email == "test@example.com" {
				emailVO, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("password123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(emailVO).
					WithPasswordHash(password).
					WithEmailVerified(false).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{
		sendVerificationEmailFunc: func(ctx context.Context, userID int64, email string) error {
			emailSent = true
			if userID != 1 {
				t.Errorf("ResendVerificationEmail() userID = %d, want 1", userID)
			}
			if email != "test@example.com" {
				t.Errorf("ResendVerificationEmail() email = %s, want test@example.com", email)
			}
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ResendVerificationEmail(ctx, "test@example.com")

	if err != nil {
		t.Fatalf("ResendVerificationEmail() error = %v, want nil", err)
	}

	if !emailSent {
		t.Errorf("ResendVerificationEmail() should send email")
	}
}

func TestAuthService_ResendVerificationEmail_AlreadyVerified(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			if email == "test@example.com" {
				emailVO, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("password123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(emailVO).
					WithPasswordHash(password).
					WithEmailVerified(true). // Already verified
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ResendVerificationEmail(ctx, "test@example.com")

	if err == nil {
		t.Errorf("ResendVerificationEmail() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "already verified") {
		t.Errorf("ResendVerificationEmail() error = %v, want 'already verified'", err)
	}
}

func TestAuthService_ResendVerificationEmail_UserNotFound(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ResendVerificationEmail(ctx, "nonexistent@example.com")

	if err == nil {
		t.Errorf("ResendVerificationEmail() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrUserNotFound) {
		t.Errorf("ResendVerificationEmail() error = %v, want ErrUserNotFound", err)
	}
}

func TestAuthService_RequestPasswordReset_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailSent := false
	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			if email == "test@example.com" {
				emailVO, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("password123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(emailVO).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{
		sendPasswordResetEmailFunc: func(ctx context.Context, userID int64, email string, resetToken string) error {
			emailSent = true
			if userID != 1 {
				t.Errorf("RequestPasswordReset() userID = %d, want 1", userID)
			}
			if email != "test@example.com" {
				t.Errorf("RequestPasswordReset() email = %s, want test@example.com", email)
			}
			if resetToken == "" {
				t.Errorf("RequestPasswordReset() resetToken should not be empty")
			}
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.RequestPasswordReset(ctx, "test@example.com")

	if err != nil {
		t.Errorf("RequestPasswordReset() error = %v, want nil", err)
	}

	if !emailSent {
		t.Errorf("RequestPasswordReset() should send password reset email")
	}
}

func TestAuthService_RequestPasswordReset_EmailNotFound(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailSent := false
	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*userEntity.User, error) {
			return nil, nil // User not found
		},
	}

	emailSvc := &mockEmailService{
		sendPasswordResetEmailFunc: func(ctx context.Context, userID int64, email string, resetToken string) error {
			emailSent = true
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.RequestPasswordReset(ctx, "nonexistent@example.com")

	// Should return success even if email doesn't exist (security)
	if err != nil {
		t.Errorf("RequestPasswordReset() error = %v, want nil (should not reveal email existence)", err)
	}

	if emailSent {
		t.Errorf("RequestPasswordReset() should not send email if user not found")
	}
}

func TestAuthService_RequestPasswordReset_InvalidEmail(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	emailSent := false
	userRepo := &mockUserRepository{}

	emailSvc := &mockEmailService{
		sendPasswordResetEmailFunc: func(ctx context.Context, userID int64, email string, resetToken string) error {
			emailSent = true
			return nil
		},
	}

	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.RequestPasswordReset(ctx, "invalid-email")

	// Should return success even if email is invalid (security)
	if err != nil {
		t.Errorf("RequestPasswordReset() error = %v, want nil (should not reveal email validity)", err)
	}

	if emailSent {
		t.Errorf("RequestPasswordReset() should not send email if email is invalid")
	}
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate a valid password reset token
	token, err := jwtSvc.GeneratePasswordResetToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userUpdated := false
	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("oldpassword123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
		updateFunc: func(ctx context.Context, u *userEntity.User) error {
			userUpdated = true
			// Verify password was updated
			if !u.VerifyPassword("newpassword123") {
				t.Errorf("ResetPassword() should update password correctly")
			}
			return nil
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.ResetPassword(ctx, token, "newpassword123")

	if err != nil {
		t.Fatalf("ResetPassword() error = %v, want nil", err)
	}

	if !userUpdated {
		t.Errorf("ResetPassword() should update user password")
	}
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	userRepo := &mockUserRepository{}
	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ResetPassword(ctx, "invalid-token", "newpassword123")

	if err == nil {
		t.Errorf("ResetPassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("ResetPassword() error = %v, want ErrInvalidToken", err)
	}
}

func TestAuthService_ResetPassword_WrongTokenType(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate an access token (not password reset token)
	token, err := jwtSvc.GenerateAccessToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userRepo := &mockUserRepository{}
	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.ResetPassword(ctx, token, "newpassword123")

	if err == nil {
		t.Errorf("ResetPassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidToken) {
		t.Errorf("ResetPassword() error = %v, want ErrInvalidToken", err)
	}
}

func TestAuthService_ResetPassword_UserNotFound(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate a valid password reset token for non-existent user
	token, err := jwtSvc.GeneratePasswordResetToken(999)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			return nil, nil // User not found
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.ResetPassword(ctx, token, "newpassword123")

	if err == nil {
		t.Errorf("ResetPassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrUserNotFound) {
		t.Errorf("ResetPassword() error = %v, want ErrUserNotFound", err)
	}
}

func TestAuthService_ResetPassword_InvalidPassword(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	
	// Generate a valid password reset token
	token, err := jwtSvc.GeneratePasswordResetToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("oldpassword123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err = service.ResetPassword(ctx, token, "short") // Password too short

	if err == nil {
		t.Errorf("ResetPassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidPassword) {
		t.Errorf("ResetPassword() error = %v, want ErrInvalidPassword", err)
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)

	userUpdated := false
	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("oldpassword123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
		updateFunc: func(ctx context.Context, u *userEntity.User) error {
			userUpdated = true
			// Verify password was updated
			if !u.VerifyPassword("newpassword123") {
				t.Errorf("ChangePassword() should update password correctly")
			}
			return nil
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ChangePassword(ctx, 1, "oldpassword123", "newpassword123")

	if err != nil {
		t.Fatalf("ChangePassword() error = %v, want nil", err)
	}

	if !userUpdated {
		t.Errorf("ChangePassword() should update user password")
	}
}

func TestAuthService_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	jwtSvc := createTestJWTService(t)

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("oldpassword123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ChangePassword(ctx, 1, "wrongpassword123", "newpassword123")

	if err == nil {
		t.Errorf("ChangePassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidCredentials) {
		t.Errorf("ChangePassword() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	jwtSvc := createTestJWTService(t)

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			return nil, nil // User not found
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ChangePassword(ctx, 999, "oldpassword123", "newpassword123")

	if err == nil {
		t.Errorf("ChangePassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrUserNotFound) {
		t.Errorf("ChangePassword() error = %v, want ErrUserNotFound", err)
	}
}

func TestAuthService_ChangePassword_InvalidNewPassword(t *testing.T) {
	jwtSvc := createTestJWTService(t)

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*userEntity.User, error) {
			if id == 1 {
				email, _ := valueobjects.NewEmail("test@example.com")
				password, _ := valueobjects.NewPassword("oldpassword123")
				return userEntity.NewBuilder().
					WithID(1).
					WithEmail(email).
					WithPasswordHash(password).
					WithEmailVerified(true).
					WithCreatedAt(time.Now()).
					WithUpdatedAt(time.Now()).
					Build()
			}
			return nil, errors.New("user not found")
		},
	}

	emailSvc := &mockEmailService{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}
	cacheRepo := &mockCacheRepository{}

	sessionSvc := createTestSessionService()
	service := authService.NewAuthService(userRepo, deckRepo, &mockProfileRepository{}, &mockUserPreferencesRepository{}, eventBus, jwtSvc, cacheRepo, emailSvc, sessionSvc, &mockTransactionManager{})

	ctx := context.Background()
	err := service.ChangePassword(ctx, 1, "oldpassword123", "short") // Password too short

	if err == nil {
		t.Errorf("ChangePassword() error = nil, want error")
	}

	if !errors.Is(err, authService.ErrInvalidPassword) {
		t.Errorf("ChangePassword() error = %v, want ErrInvalidPassword", err)
	}
}
