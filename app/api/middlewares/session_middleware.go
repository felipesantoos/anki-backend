package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/core/services/session"
)

const (
	// SessionIDCookieName is the name of the cookie used to store session ID
	SessionIDCookieName = "session_id"
	// SessionIDHeaderName is the name of the header used to store session ID (alternative to cookie)
	SessionIDHeaderName = "X-Session-ID"
	// SessionContextKey is the key used to store session data in Echo context
	SessionContextKey = "session_data"
)

// SessionMiddleware creates a middleware for managing sessions
// It extracts sessionID from cookie or header, loads session data, and stores it in context
// Optionally refreshes session TTL on each request
func SessionMiddleware(sessionService *session.SessionService, refreshOnRequest bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract sessionID from cookie or header
			sessionID := extractSessionID(c)

			if sessionID == "" {
				// No session ID found, continue without session data
				return next(c)
			}

			// Get session data
			sessionData, err := sessionService.GetSession(c.Request().Context(), sessionID)
			if err != nil {
				// Session not found or error - continue without session data
				// Optionally, could clear invalid session cookie here
				return next(c)
			}

			// Store session data in context
			c.Set(SessionContextKey, sessionData)
			c.Set("session_id", sessionID)

			// Refresh session TTL if enabled
			if refreshOnRequest {
				if err := sessionService.RefreshSession(c.Request().Context(), sessionID); err != nil {
					// Log error but don't fail the request
					// Session refresh failure is not critical
				}
			}

			return next(c)
		}
	}
}

// extractSessionID extracts session ID from cookie or header
// Priority: cookie > header
func extractSessionID(c echo.Context) string {
	// Try cookie first
	cookie, err := c.Cookie(SessionIDCookieName)
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fallback to header
	return c.Request().Header.Get(SessionIDHeaderName)
}

// GetSessionData retrieves session data from Echo context
// Returns nil if session data is not found
func GetSessionData(c echo.Context) map[string]interface{} {
	data, ok := c.Get(SessionContextKey).(map[string]interface{})
	if !ok {
		return nil
	}
	return data
}

// GetSessionID retrieves session ID from Echo context
// Returns empty string if session ID is not found
func GetSessionID(c echo.Context) string {
	sessionID, ok := c.Get("session_id").(string)
	if !ok {
		return ""
	}
	return sessionID
}

// SetSessionCookie sets the session ID cookie in the response
func SetSessionCookie(c echo.Context, sessionID string, maxAge int) {
	cookie := &http.Cookie{
		Name:     SessionIDCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.Scheme() == "https", // Secure in HTTPS only
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// ClearSessionCookie clears the session ID cookie
func ClearSessionCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     SessionIDCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

