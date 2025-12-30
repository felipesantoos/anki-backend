package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockBackupService)
	handler := handlers.NewBackupHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateBackupRequest{
			Filename:    "backup.colpkg",
			Size:        1024,
			StoragePath: "/path",
			BackupType:  "automatic",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/backups", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		b, _ := backup.NewBuilder().WithID(1).WithUserID(userID).WithFilename(reqBody.Filename).WithBackupType(reqBody.BackupType).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Filename, reqBody.Size, reqBody.StoragePath, reqBody.BackupType).Return(b, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestBackupHandler_FindAll(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockBackupService)
	handler := handlers.NewBackupHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/backups", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		b1, _ := backup.NewBuilder().WithID(1).WithUserID(userID).WithFilename("f1").WithBackupType("manual").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return([]*backup.Backup{b1}, nil).Once()

		if assert.NoError(t, handler.FindAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestBackupHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockBackupService)
	handler := handlers.NewBackupHandler(mockSvc)
	userID := int64(1)
	backupID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/backups/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, backupID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

