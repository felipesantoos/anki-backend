package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuditHandler_GetDeletionLogs(t *testing.T) {
	e := echo.New()
	mockDlSvc := new(MockDeletionLogService)
	mockUhSvc := new(MockUndoHistoryService)
	handler := handlers.NewAuditHandler(mockDlSvc, mockUhSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/deletions", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		dl1, _ := deletionlog.NewBuilder().WithID(1).WithUserID(userID).WithObjectType("card").Build()
		mockDlSvc.On("FindByUserID", mock.Anything, userID).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		if assert.NoError(t, handler.GetDeletionLogs(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockDlSvc.AssertExpectations(t)
	})
}

func TestAuditHandler_GetUndoHistory(t *testing.T) {
	e := echo.New()
	mockDlSvc := new(MockDeletionLogService)
	mockUhSvc := new(MockUndoHistoryService)
	handler := handlers.NewAuditHandler(mockDlSvc, mockUhSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/undo", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		uh1, _ := undohistory.NewBuilder().WithID(1).WithUserID(userID).WithOperationType(undohistory.OperationTypeEditNote).Build()
		mockUhSvc.On("FindLatest", mock.Anything, userID, mock.Anything).Return([]*undohistory.UndoHistory{uh1}, nil).Once()

		if assert.NoError(t, handler.GetUndoHistory(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockUhSvc.AssertExpectations(t)
	})
}
