package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/felipesantos/anki-backend/config"
)

var (
	// ErrInvalidToken is returned when token is invalid or expired
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token expired")
)

// Claims represents JWT claims structure
type Claims struct {
	UserID int64  `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTService provides JWT token generation and validation
type JWTService struct {
	secretKey           []byte
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
	issuer              string
}

// NewJWTService creates a new JWT service instance
func NewJWTService(cfg config.JWTConfig) (*JWTService, error) {
	if cfg.SecretKey == "" {
		return nil, errors.New("JWT secret key is required")
	}

	if len(cfg.SecretKey) < 32 {
		return nil, errors.New("JWT secret key must be at least 32 characters long")
	}

	// Convert expiry times from minutes/days to duration
	accessTokenExpiry := time.Duration(cfg.AccessTokenExpiry) * time.Minute
	refreshTokenExpiry := time.Duration(cfg.RefreshTokenExpiry) * 24 * time.Hour

	return &JWTService{
		secretKey:           []byte(cfg.SecretKey),
		accessTokenExpiry:   accessTokenExpiry,
		refreshTokenExpiry:  refreshTokenExpiry,
		issuer:              cfg.Issuer,
	}, nil
}

// GenerateAccessToken generates a new access token for the given user ID
func (s *JWTService) GenerateAccessToken(userID int64) (string, error) {
	return s.generateToken(userID, "access", s.accessTokenExpiry)
}

// GenerateRefreshToken generates a new refresh token for the given user ID
func (s *JWTService) GenerateRefreshToken(userID int64) (string, error) {
	return s.generateToken(userID, "refresh", s.refreshTokenExpiry)
}

// generateToken generates a JWT token with the given parameters
func (s *JWTService) generateToken(userID int64, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and decodes a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractUserID extracts the user ID from a token string
func (s *JWTService) ExtractUserID(tokenString string) (int64, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetAccessTokenExpiry returns the access token expiry duration
func (s *JWTService) GetAccessTokenExpiry() time.Duration {
	return s.accessTokenExpiry
}

// GetRefreshTokenExpiry returns the refresh token expiry duration
func (s *JWTService) GetRefreshTokenExpiry() time.Duration {
	return s.refreshTokenExpiry
}

