package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/mappers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// NoteHandler handles note-related HTTP requests
type NoteHandler struct {
	service           primary.INoteService
	exportService     primary.IExportService
	deletionLogService primary.IDeletionLogService
}

// NewNoteHandler creates a new NoteHandler instance
func NewNoteHandler(service primary.INoteService, exportService primary.IExportService, deletionLogService primary.IDeletionLogService) *NoteHandler {
	return &NoteHandler{
		service:           service,
		exportService:     exportService,
		deletionLogService: deletionLogService,
	}
}

// Create handles POST /api/v1/notes
// @Summary Create a note
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateNoteRequest true "Note creation request"
// @Success 201 {object} response.NoteResponse
// @Router /api/v1/notes [post]
func (h *NoteHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.CreateNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	n, err := h.service.Create(ctx, userID, req.NoteTypeID, req.DeckID, req.FieldsJSON, req.Tags)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusCreated, mappers.ToNoteResponse(n))
}

// FindByID handles GET /api/v1/notes/:id
// @Summary Get note by ID
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Success 200 {object} response.NoteResponse
// @Router /api/v1/notes/{id} [get]
func (h *NoteHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	n, err := h.service.FindByID(ctx, userID, id)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponse(n))
}

// FindAll handles GET /api/v1/notes
// @Summary List notes
// @Description List notes with optional filters. Multiple tags use OR logic (returns notes with ANY of the specified tags). Tag search is case-insensitive.
// @Tags notes
// @Produce json
// @Security BearerAuth
// @Param deck_id query int false "Filter by deck ID"
// @Param note_type_id query int false "Filter by note type ID"
// @Param tags query []string false "Filter by tags (OR logic: returns notes with ANY of the specified tags). Case-insensitive."
// @Param search query string false "Search in fields (takes priority over other filters)"
// @Param limit query int false "Pagination limit (default: 50)"
// @Param offset query int false "Pagination offset (default: 0)"
// @Success 200 {array} response.NoteResponse
// @Router /api/v1/notes [get]
func (h *NoteHandler) FindAll(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.ListNotesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	filters := note.NoteFilters{
		DeckID:     req.DeckID,
		NoteTypeID: req.NoteTypeID,
		Tags:       req.Tags,
		Search:     req.Search,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	notes, err := h.service.FindAll(ctx, userID, filters)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponseList(notes))
}

// Update handles PUT /api/v1/notes/:id
// @Summary Update note
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param request body request.UpdateNoteRequest true "Update request"
// @Success 200 {object} response.NoteResponse
// @Router /api/v1/notes/{id} [put]
func (h *NoteHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.UpdateNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	n, err := h.service.Update(ctx, userID, id, req.FieldsJSON, req.Tags)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToNoteResponse(n))
}

// Delete handles DELETE /api/v1/notes/:id
// @Summary Delete note
// @Tags notes
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id} [delete]
func (h *NoteHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.service.Delete(ctx, userID, id); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// AddTag handles POST /api/v1/notes/:id/tags
// @Summary Add tag to note
// @Tags notes
// @Accept json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param request body request.AddTagRequest true "Tag request"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id}/tags [post]
func (h *NoteHandler) AddTag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.AddTagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.service.AddTag(ctx, userID, id, req.Tag); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RemoveTag handles DELETE /api/v1/notes/:id/tags/:tag
// @Summary Remove tag from note
// @Tags notes
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param tag path string true "Tag name"
// @Success 204 "No Content"
// @Router /api/v1/notes/{id}/tags/{tag} [delete]
func (h *NoteHandler) RemoveTag(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tag := c.Param("tag")

	if err := h.service.RemoveTag(ctx, userID, id, tag); err != nil {
		return handleNoteError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Copy handles POST /api/v1/notes/:id/copy
// @Summary Copy a note
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Note ID"
// @Param request body request.CopyNoteRequest true "Copy request"
// @Success 201 {object} response.NoteResponse
// @Router /api/v1/notes/{id}/copy [post]
func (h *NoteHandler) Copy(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req request.CopyNoteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	n, err := h.service.Copy(ctx, userID, id, req.DeckID, req.CopyTags, req.CopyMedia)
	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusCreated, mappers.ToNoteResponse(n))
}

// FindDuplicates handles POST /api/v1/notes/find-duplicates
// @Summary Find duplicate notes
// @Description Find duplicate notes based on a field value or GUID. If use_guid is true, finds duplicates by GUID value (ignores field_name and note_type_id). If use_guid is false and note_type_id is provided with empty field_name, automatically uses the first field of the note type.
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.FindDuplicatesRequest true "Find duplicates request"
// @Success 200 {object} response.FindDuplicatesResponse
// @Router /api/v1/notes/find-duplicates [post]
func (h *NoteHandler) FindDuplicates(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.FindDuplicatesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	var result *note.DuplicateResult
	var err error

	// If UseGUID is true, use GUID-based detection
	// Note: When UseGUID is true, FieldName and NoteTypeID are ignored
	if req.UseGUID {
		result, err = h.service.FindDuplicatesByGUID(ctx, userID)
	} else {
		// Otherwise, use field-based detection
		// FieldName and NoteTypeID are used for field-based duplicate detection
		result, err = h.service.FindDuplicates(ctx, userID, req.NoteTypeID, req.FieldName)
	}

	if err != nil {
		return handleNoteError(err)
	}

	return c.JSON(http.StatusOK, mappers.ToFindDuplicatesResponse(result))
}

// Export handles POST /api/v1/notes/export
// @Summary Export selected notes
// @Description Export selected notes in the specified format (apkg or text). Optionally include media files and scheduling information.
// @Tags notes
// @Accept json
// @Produce application/zip,text/plain
// @Security BearerAuth
// @Param request body request.ExportNotesRequest true "Export notes request"
// @Success 200 {file} file "Export file (apkg or text)"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Notes not found"
// @Router /api/v1/notes/export [post]
func (h *NoteHandler) Export(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	var req request.ExportNotesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if len(req.NoteIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "note_ids cannot be empty")
	}
	if len(req.NoteIDs) > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "note_ids cannot exceed 1000")
	}
	if req.Format != "apkg" && req.Format != "text" {
		return echo.NewHTTPError(http.StatusBadRequest, "format must be 'apkg' or 'text'")
	}

	// Call export service
	reader, size, filename, err := h.exportService.ExportNotes(
		ctx,
		userID,
		req.NoteIDs,
		req.Format,
		req.IncludeMedia,
		req.IncludeScheduling,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Set response headers
	contentType := "application/zip"
	if req.Format == "text" {
		contentType = "text/plain; charset=utf-8"
	}
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", strconv.FormatInt(size, 10))

	// Stream file
	return c.Stream(http.StatusOK, contentType, reader)
}

// GetRecentDeletions handles GET /api/v1/notes/deletions
// @Summary Get recent deletions
// @Description Retrieves recent deletion logs that can be recovered. Supports limit and days query parameters.
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of deletions to return" default(20) minimum(1) maximum(100)
// @Param days query int false "Number of days to look back" default(7) minimum(1) maximum(365)
// @Success 200 {object} response.RecentDeletionsResponse "Recent deletions with pagination"
// @Failure 400 {object} response.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/notes/deletions [get]
func (h *NoteHandler) GetRecentDeletions(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse query parameters with defaults
	limitStr := c.QueryParam("limit")
	daysStr := c.QueryParam("days")

	limit := 20 // Default limit
	days := 7   // Default days

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "limit must be a positive integer")
		}
		if parsedLimit > 100 {
			return echo.NewHTTPError(http.StatusBadRequest, "limit cannot exceed 100")
		}
		limit = parsedLimit
	}

	if daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err != nil || parsedDays <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "days must be a positive integer")
		}
		if parsedDays > 365 {
			return echo.NewHTTPError(http.StatusBadRequest, "days cannot exceed 365")
		}
		days = parsedDays
	}

	// Call service to get recent deletions
	logs, err := h.deletionLogService.FindRecent(ctx, userID, limit, days)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve recent deletions: %v", err))
	}

	// Convert to response DTOs
	deletionLogResponses := mappers.ToDeletionLogResponseList(logs)

	// Calculate pagination metadata
	total := len(logs)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	response := response.RecentDeletionsResponse{
		Data: deletionLogResponses,
		Pagination: response.PaginationResponse{
			Page:       1, // Always page 1 for recent deletions
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return c.JSON(http.StatusOK, response)
}

// RestoreDeletion handles POST /api/v1/notes/deletions/:id/restore
// @Summary Restore a deleted note
// @Description Restores a deleted note from a deletion log entry. The note will be recreated with its original data in the specified deck.
// @Tags notes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Deletion log ID"
// @Param request body request.RestoreDeletionRequest true "Restore deletion request"
// @Success 200 {object} response.RestoreDeletionResponse "Note restored successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Deletion log not found"
// @Failure 409 {object} response.ErrorResponse "Note already restored or conflict"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /api/v1/notes/deletions/{id}/restore [post]
func (h *NoteHandler) RestoreDeletion(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middlewares.GetUserID(c)

	// Parse deletion log ID from path parameter
	deletionLogID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || deletionLogID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid deletion log ID")
	}

	// Bind request body
	var req request.RestoreDeletionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.DeckID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "deck_id is required and must be greater than 0")
	}

	// Call service to restore note
	restoredNote, err := h.deletionLogService.Restore(ctx, userID, deletionLogID, req.DeckID)
	if err != nil {
		// Handle specific error cases
		errMsg := err.Error()
		
		// Deck not found should return 400 (Bad Request) as it's an invalid parameter
		// Check for "deck not found" FIRST before other "not found" checks
		// (may be wrapped in "failed to create note")
		if strings.Contains(errMsg, "deck not found") {
			return echo.NewHTTPError(http.StatusBadRequest, "deck not found")
		}
		
		// Note type not found should also return 400
		if strings.Contains(errMsg, "note type not found") {
			return echo.NewHTTPError(http.StatusBadRequest, "note type not found")
		}
		
		// If error is ownership.ErrResourceNotFound and contains "failed to create note",
		// it's likely a deck or note type not found (invalid parameter), return 400
		if errors.Is(err, ownership.ErrResourceNotFound) && strings.Contains(errMsg, "failed to create note") {
			return echo.NewHTTPError(http.StatusBadRequest, "deck or note type not found")
		}
		
		// Deletion log not found should return 404
		if strings.Contains(errMsg, "deletion log") && strings.Contains(errMsg, "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Deletion log not found or access denied")
		}
		if errors.Is(err, ownership.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Deletion log not found or access denied")
		}
		
		// Already restored or conflict - this shouldn't happen normally but can occur
		// due to test isolation issues or race conditions. Treat as success (idempotent).
		if strings.Contains(errMsg, "already restored") {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if strings.Contains(errMsg, "conflict") {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		
		// Invalid or missing data should return 400
		if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "missing") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to restore note: %v", err))
	}

	// Map to response DTO
	response := response.RestoreDeletionResponse{
		ID:         restoredNote.GetID(),
		GUID:       restoredNote.GetGUID().Value(),
		RestoredAt: time.Now(),
		Message:    "Note restored successfully",
	}

	return c.JSON(http.StatusOK, response)
}

func handleNoteError(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "not found") {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	// Default to bad request for other validation errors
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}

