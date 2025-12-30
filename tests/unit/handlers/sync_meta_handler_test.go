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
	"github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncMetaHandler_FindMe(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSyncMetaService)
	handler := handlers.NewSyncMetaHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/meta", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		sm, _ := syncmeta.NewBuilder().WithID(1).WithUserID(userID).WithClientID("c1").Build()
		mockSvc.On("FindByUserID", mock.Anything, userID).Return(sm, nil).Once()

		if assert.NoError(t, handler.FindMe(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}

func TestSyncMetaHandler_Update(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockSyncMetaService)
	handler := handlers.NewSyncMetaHandler(mockSvc)
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		reqBody := request.UpdateSyncMetaRequest{
			ClientID:    "client123",
			LastSyncUSN: 100,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/sync/meta", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middlewares.UserIDContextKey, userID)

		sm, _ := syncmeta.NewBuilder().WithID(1).WithUserID(userID).WithClientID(reqBody.ClientID).Build()
		mockSvc.On("Update", mock.Anything, userID, reqBody.ClientID, reqBody.LastSyncUSN).Return(sm, nil).Once()

		if assert.NoError(t, handler.Update(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
		mockSvc.AssertExpectations(t)
	})
}
