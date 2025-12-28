package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/services/session"
)

// SessionHandler handles session management HTTP requests
type SessionHandler struct {
	sessionService *session.SessionService
}

// NewSessionHandler creates a new SessionHandler instance
func NewSessionHandler(sessionService *session.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// GetSessions handles GET /api/v1/auth/sessions requests
// @Summary List user sessions
// @Description Returns a list of all active sessions for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer JWT token (access token)"
// @Success 200 {object} response.SessionListResponse "List of sessions"
// @Failure 401 {object} response.ErrorResponse "Not authenticated"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions [get]
func (h *SessionHandler) GetSessions(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract user ID from context (set by auth middleware)
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get current session ID from context (if available)
	currentSessionID := middlewares.GetSessionID(c)

	// Get all sessions for the user
	sessions, err := h.sessionService.GetUserSessions(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve sessions")
	}

	// Convert to response format
	sessionResponses := make([]response.SessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		// Get session ID from the map
		sessionID, _ := sess["id"].(string)
		sessionResp := h.mapSessionToResponse(sess, currentSessionID)
		sessionResp.ID = sessionID
		sessionResponses = append(sessionResponses, sessionResp)
	}

	return c.JSON(http.StatusOK, response.SessionListResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	})
}

// GetSession handles GET /api/v1/auth/sessions/:id requests
// @Summary Get session details
// @Description Returns details of a specific session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer JWT token (access token)"
// @Param id path string true "Session ID"
// @Success 200 {object} response.SessionResponse "Session details"
// @Failure 401 {object} response.ErrorResponse "Not authenticated"
// @Failure 403 {object} response.ErrorResponse "Session does not belong to user"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions/{id} [get]
func (h *SessionHandler) GetSession(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract user ID from context
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get session ID from path
	sessionID := c.Param("id")
	if sessionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Session ID is required")
	}

	// Get session data
	sessionData, err := h.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	// Verify session belongs to user
	sessionUserID, ok := sessionData["userIDInt"].(int64)
	if !ok {
		// Fallback to string conversion
		userIDStr, ok := sessionData["userID"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Invalid session data")
		}
		var parseErr error
		if sessionUserID, parseErr = strconv.ParseInt(userIDStr, 10, 64); parseErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Invalid session data")
		}
	}

	if sessionUserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Session does not belong to user")
	}

	// Get current session ID
	currentSessionID := middlewares.GetSessionID(c)

	// Convert to response
	sessionResp := h.mapSessionToResponse(sessionData, currentSessionID)
	sessionResp.ID = sessionID

	return c.JSON(http.StatusOK, sessionResp)
}

// DeleteSession handles DELETE /api/v1/auth/sessions/:id requests
// @Summary Delete a session
// @Description Invalidates a specific session (logout from specific device)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer JWT token (access token)"
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]string "Session deleted successfully"
// @Failure 401 {object} response.ErrorResponse "Not authenticated"
// @Failure 403 {object} response.ErrorResponse "Session does not belong to user"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract user ID from context
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get session ID from path
	sessionID := c.Param("id")
	if sessionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Session ID is required")
	}

	// Delete session (this will verify ownership)
	err := h.sessionService.DeleteUserSession(ctx, userID, sessionID)
	if err != nil {
		if err.Error() == "session does not belong to user" || err.Error() == "session not found" {
			return echo.NewHTTPError(http.StatusForbidden, "Session does not belong to user or not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete session")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Session deleted successfully",
	})
}

// DeleteAllSessions handles DELETE /api/v1/auth/sessions requests
// @Summary Delete all sessions
// @Description Invalidates all sessions for the authenticated user (logout from all devices)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer JWT token (access token)"
// @Success 200 {object} map[string]string "All sessions deleted successfully"
// @Failure 401 {object} response.ErrorResponse "Not authenticated"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/auth/sessions [delete]
func (h *SessionHandler) DeleteAllSessions(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract user ID from context
	userID := middlewares.GetUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Delete all sessions
	err := h.sessionService.DeleteAllUserSessions(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete sessions")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "All sessions deleted successfully",
	})
}

// mapSessionToResponse converts session data map to SessionResponse
func (h *SessionHandler) mapSessionToResponse(sessionData map[string]interface{}, currentSessionID string) response.SessionResponse {
	sessionID, _ := sessionData["id"].(string)
	
	// Extract user ID
	var userID int64
	if userIDInt, ok := sessionData["userIDInt"].(int64); ok {
		userID = userIDInt
	} else if userIDStr, ok := sessionData["userID"].(string); ok {
		if parsed, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = parsed
		}
	}

	// Extract timestamps
	var createdAt, lastActivity time.Time
	if createdAtUnix, ok := sessionData["createdAt"].(int64); ok {
		createdAt = time.Unix(createdAtUnix, 0)
	}
	if lastActivityUnix, ok := sessionData["lastActivity"].(int64); ok {
		lastActivity = time.Unix(lastActivityUnix, 0)
	} else {
		lastActivity = createdAt // Fallback to createdAt if lastActivity not set
	}

	// Extract strings
	ipAddress, _ := sessionData["ipAddress"].(string)
	userAgent, _ := sessionData["userAgent"].(string)
	deviceInfo, _ := sessionData["deviceInfo"].(string)

	// Check if this is the current session
	isCurrent := sessionID != "" && sessionID == currentSessionID

	return response.SessionResponse{
		ID:           sessionID,
		UserID:      userID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		DeviceInfo:  deviceInfo,
		CreatedAt:   createdAt,
		LastActivity: lastActivity,
		IsCurrent:   isCurrent,
	}
}

