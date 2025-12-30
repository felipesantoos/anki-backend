package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// UserPreferencesHandler handles user preferences HTTP requests
type UserPreferencesHandler struct {
	service primary.IUserPreferencesService
}

// NewUserPreferencesHandler creates a new UserPreferencesHandler instance
func NewUserPreferencesHandler(service primary.IUserPreferencesService) *UserPreferencesHandler {
	return &UserPreferencesHandler{
		service: service,
	}
}

// FindByUserID handles GET /api/v1/user/preferences
// @Summary Get user preferences
// @Tags preferences
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserPreferencesResponse
// @Router /api/v1/user/preferences [get]
func (h *UserPreferencesHandler) FindByUserID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	prefs, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToUserPreferencesResponse(prefs))
}

// Update handles PUT /api/v1/user/preferences
// @Summary Update user preferences
// @Tags preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateUserPreferencesRequest true "Update request"
// @Success 200 {object} response.UserPreferencesResponse
// @Router /api/v1/user/preferences [put]
func (h *UserPreferencesHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.UpdateUserPreferencesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 1. Get existing preferences to get the ID and current values
	existingPrefs, err := h.service.FindByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch existing preferences")
	}

	// 2. Manual mapping to domain entity for update
	// Note: In a more complex scenario, this could be in a mapper or builder
	prefs, err := userpreferences.NewBuilder().
		WithID(existingPrefs.GetID()). // Crucial: use existing ID
		WithUserID(userID).
		WithLanguage(req.Language).
		WithTheme(valueobjects.ThemeType(req.Theme)).
		WithAutoSync(req.AutoSync).
		WithNextDayStartsAt(req.NextDayStartsAt).
		WithLearnAheadLimit(req.LearnAheadLimit).
		WithTimeboxTimeLimit(req.TimeboxTimeLimit).
		WithVideoDriver(req.VideoDriver).
		WithUISize(req.UISize).
		WithMinimalistMode(req.MinimalistMode).
		WithReduceMotion(req.ReduceMotion).
		WithPasteStripsFormatting(req.PasteStripsFormatting).
		WithPasteImagesAsPNG(req.PasteImagesAsPNG).
		WithDefaultDeckBehavior(req.DefaultDeckBehavior).
		WithShowPlayButtons(req.ShowPlayButtons).
		WithInterruptAudioOnAnswer(req.InterruptAudioOnAnswer).
		WithShowRemainingCount(req.ShowRemainingCount).
		WithShowNextReviewTime(req.ShowNextReviewTime).
		WithSpacebarAnswersCard(req.SpacebarAnswersCard).
		WithIgnoreAccentsInSearch(req.IgnoreAccentsInSearch).
		WithDefaultSearchText(req.DefaultSearchText).
		WithSyncAudioAndImages(req.SyncAudioAndImages).
		WithPeriodicallySyncMedia(req.PeriodicallySyncMedia).
		WithForceOneWaySync(req.ForceOneWaySync).
		WithSelfHostedSyncServerURL(req.SelfHostedSyncServerURL).
		Build()

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to build preferences: %w", err).Error())
	}

	if err := h.service.Update(ctx, userID, prefs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToUserPreferencesResponse(prefs))
}

// ResetToDefaults handles POST /api/v1/user/preferences/reset
// @Summary Reset user preferences to defaults
// @Tags preferences
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserPreferencesResponse
// @Router /api/v1/user/preferences/reset [post]
func (h *UserPreferencesHandler) ResetToDefaults(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	prefs, err := h.service.ResetToDefaults(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mappers.ToUserPreferencesResponse(prefs))
}

