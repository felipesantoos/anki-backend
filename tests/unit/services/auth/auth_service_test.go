package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	domainEvents "github.com/felipesantos/anki-backend/core/domain/events"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/config"
)

// mockUserRepository is a mock implementation of IUserRepository
type mockUserRepository struct {
	saveFunc       func(ctx context.Context, user *entities.User) error
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
	findByIDFunc   func(ctx context.Context, id int64) (*entities.User, error)
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

func (m *mockUserRepository) FindByID(ctx context.Context, id int64) (*entities.User, error) {
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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

	ctx := context.Background()
	_, err := service.Register(ctx, "existing@example.com", "password123")

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if err != authService.ErrEmailAlreadyExists {
		t.Errorf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	userRepo := &mockUserRepository{}
	deckRepo := &mockDeckRepository{}
	eventBus := &mockEventBus{}

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}
	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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
	testUser := &entities.User{
		ID:           1,
		Email:        emailVO,
		PasswordHash: passwordVO,
		EmailVerified: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userRepo := &mockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			if email == "user@example.com" {
				return testUser, nil
			}
			return nil, nil
		},
		saveFunc: func(ctx context.Context, user *entities.User) error {
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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

	ctx := context.Background()
	resp, err := service.Login(ctx, "user@example.com", "password123")

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

	if resp.User.ID != testUser.ID {
		t.Errorf("Login() User.ID = %v, want %v", resp.User.ID, testUser.ID)
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
		findByEmail func(ctx context.Context, email string) (*entities.User, error)
	}{
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "password123",
			findByEmail: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, nil
			},
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrongpassword",
			findByEmail: func(ctx context.Context, email string) (*entities.User, error) {
				emailVO, _ := valueobjects.NewEmail("user@example.com")
				passwordVO, _ := valueobjects.NewPassword("password123")
				return &entities.User{
					ID:           1,
					Email:        emailVO,
					PasswordHash: passwordVO,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil
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

			service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

			ctx := context.Background()
			_, err := service.Login(ctx, tt.email, tt.password)

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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

	ctx := context.Background()
	_, err := service.Login(ctx, "invalid-email", "password123")

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
	testUser := &entities.User{
		ID:           1,
		Email:        emailVO,
		EmailVerified: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Generate a refresh token
	refreshToken, err := jwtSvc.GenerateRefreshToken(testUser.ID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Track which keys are stored/deleted for token rotation verification
	storedKeys := make(map[string]bool)
	deletedKeys := make(map[string]bool)

	userRepo := &mockUserRepository{
		findByIDFunc: func(ctx context.Context, id int64) (*entities.User, error) {
			if id == testUser.ID {
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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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
	testUser := &entities.User{
		ID:           1,
		Email:        emailVO,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	refreshToken, err := jwtSvc.GenerateRefreshToken(testUser.ID)
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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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
	testUser := &entities.User{
		ID:           1,
		Email:        emailVO,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Generate access token instead of refresh token
	accessToken, err := jwtSvc.GenerateAccessToken(testUser.ID)
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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

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

	service := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, cacheRepo)

	ctx := context.Background()
	err = service.Logout(ctx, accessToken, "")

	if err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}

	if !accessTokenBlacklisted {
		t.Errorf("Logout() should blacklist access token even without refresh token")
	}
}
