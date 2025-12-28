package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	jwtpkg "github.com/felipesantos/anki-backend/pkg/jwt"
)

const (
	// UserIDContextKey is the key used to store user ID in Echo context
	UserIDContextKey = "user_id"
	// AccessTokenContextKey is the key used to store access token in Echo context
	AccessTokenContextKey = "access_token"
)

// AuthMiddleware creates a middleware for JWT authentication
// It extracts and validates JWT tokens from Authorization header,
// checks if token is blacklisted, and stores userID in context
func AuthMiddleware(jwtService *jwtpkg.JWTService, cacheRepo secondary.ICacheRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header is required")
			}

			// Extract token from "Bearer <token>" format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format. Expected: Bearer <token>")
			}

			tokenString := parts[1]
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token is required")
			}

			// Check if token is blacklisted
			ctx := c.Request().Context()
			// Hash token using SHA256 (same approach as in auth_service.go)
			tokenHash := sha256.Sum256([]byte(tokenString))
			tokenHashHex := hex.EncodeToString(tokenHash[:])
			blacklistKey := fmt.Sprintf("access_token_blacklist:%s", tokenHashHex)
			exists, err := cacheRepo.Exists(ctx, blacklistKey)
			if err != nil {
				// Log error but don't fail - cache errors shouldn't prevent authentication
				// In production, you might want to log this
			} else if exists {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token has been invalidated")
			}

			// Validate token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				// Check if it's a token type mismatch error
				if strings.Contains(err.Error(), "not an access token") {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token type")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Store user ID and token in context
			c.Set(UserIDContextKey, claims.UserID)
			c.Set(AccessTokenContextKey, tokenString)

			return next(c)
		}
	}
}

// GetUserID extracts the user ID from Echo context
// Returns 0 if user ID is not found (user not authenticated)
func GetUserID(c echo.Context) int64 {
	userID, ok := c.Get(UserIDContextKey).(int64)
	if !ok {
		return 0
	}
	return userID
}

// GetAccessToken extracts the access token from Echo context
// Returns empty string if token is not found
func GetAccessToken(c echo.Context) string {
	token, ok := c.Get(AccessTokenContextKey).(string)
	if !ok {
		return ""
	}
	return token
}

