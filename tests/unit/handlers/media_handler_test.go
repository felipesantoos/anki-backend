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
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMediaHandler_Create(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockMediaService)
	handler := handlers.NewMediaHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.CreateMediaRequest{
			Filename:    "image.png",
			Hash:        "hash123",
			Size:        512,
			MimeType:    "image/png",
			StoragePath: "/path",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/media", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		m, _ := media.NewBuilder().WithID(1).WithUserID(userID).WithFilename(reqBody.Filename).Build()
		mockSvc.On("Create", mock.Anything, userID, reqBody.Filename, reqBody.Hash, reqBody.Size, reqBody.MimeType, reqBody.StoragePath).Return(m, nil).Once()

		if assert.NoError(t, handler.Create(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestMediaHandler_FindByID(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockMediaService)
	handler := handlers.NewMediaHandler(mockSvc)
	userID := int64(1)
	mediaID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/media/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		m, _ := media.NewBuilder().WithID(mediaID).WithUserID(userID).WithFilename("test.png").Build()
		mockSvc.On("FindByID", mock.Anything, userID, mediaID).Return(m, nil).Once()

		if assert.NoError(t, handler.FindByID(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestMediaHandler_Delete(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockMediaService)
	handler := handlers.NewMediaHandler(mockSvc)
	userID := int64(1)
	mediaID := int64(10)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/media/:id")
		c.SetParamNames("id")
		c.SetParamValues("10")
		c.Set(middlewares.UserIDContextKey, userID)

		mockSvc.On("Delete", mock.Anything, userID, mediaID).Return(nil).Once()

		if assert.NoError(t, handler.Delete(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

