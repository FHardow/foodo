package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPushRouter(repo *mock.PushRepo, authUserID string, roles []string) *gin.Engine {
	svc := push.NewService(repo)
	h := handler.NewPushHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, authUserID)
		c.Set(middleware.RolesKey, roles)
		c.Next()
	})
	r.POST("/push/subscribe", h.Subscribe)
	r.DELETE("/push/subscribe", h.Unsubscribe)

	return r
}

func deleteJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodDelete, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func validSubscribeBody() map[string]any {
	return map[string]any{
		"endpoint": "https://push.example.com/abc",
		"keys": map[string]any{
			"p256dh": "p256dh-key",
			"auth":   "auth-key",
		},
	}
}

func TestPushHandler_Subscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New().String()
	router := setupPushRouter(repo, userID, nil)

	w := postJSON(router, "/push/subscribe", validSubscribeBody())
	assert.Equal(t, http.StatusNoContent, w.Code)

	parsedUserID, err := uuid.Parse(userID)
	require.NoError(t, err)
	found, err := repo.ListByUser(context.Background(), parsedUserID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
}

func TestPushHandler_Subscribe_InvalidUserID(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), "not-a-uuid", nil)

	w := postJSON(router, "/push/subscribe", validSubscribeBody())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushHandler_Subscribe_MissingFields(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), uuid.New().String(), nil)

	w := postJSON(router, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestPushHandler_Unsubscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New().String()
	router := setupPushRouter(repo, userID, nil)

	postJSON(router, "/push/subscribe", validSubscribeBody())

	w := deleteJSON(router, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusNoContent, w.Code)

	parsedUserID, err := uuid.Parse(userID)
	require.NoError(t, err)
	found, err := repo.ListByUser(context.Background(), parsedUserID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushHandler_Unsubscribe_MissingEndpoint(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), uuid.New().String(), nil)

	w := deleteJSON(router, "/push/subscribe", map[string]any{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushHandler_Unsubscribe_WrongUser_Forbidden(t *testing.T) {
	repo := mock.NewPushRepo()
	ownerRouter := setupPushRouter(repo, uuid.New().String(), nil)
	postJSON(ownerRouter, "/push/subscribe", validSubscribeBody())

	attackerRouter := setupPushRouter(repo, uuid.New().String(), nil)
	w := deleteJSON(attackerRouter, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPushHandler_Unsubscribe_Owner_CanDeleteAnyUsersSubscription(t *testing.T) {
	repo := mock.NewPushRepo()
	subscriberID := uuid.New().String()
	subscriberRouter := setupPushRouter(repo, subscriberID, nil)
	postJSON(subscriberRouter, "/push/subscribe", validSubscribeBody())

	adminRouter := setupPushRouter(repo, uuid.New().String(), []string{"owner"})
	w := deleteJSON(adminRouter, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusNoContent, w.Code)

	parsedSubscriberID, err := uuid.Parse(subscriberID)
	require.NoError(t, err)
	found, err := repo.ListByUser(context.Background(), parsedSubscriberID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushHandler_Unsubscribe_NotFound(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), uuid.New().String(), nil)

	w := deleteJSON(router, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/never-subscribed"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}
