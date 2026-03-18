package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/http/handler"
	"github.com/fhardow/bread-order/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupUserRouter wires a fresh mock repo → service → handler into a gin engine.
func setupUserRouter(repo *mock.UserRepo) (*gin.Engine, *handler.UserHandler) {
	svc := user.NewService(repo)
	h := handler.NewUserHandler(svc)

	r := gin.New()
	r.POST("/users", h.Register)
	r.GET("/users", h.List)
	r.GET("/users/:id", h.GetByID)
	r.PUT("/users/:id", h.Update)

	return r, h
}

func postJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func getRequest(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func putJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestUserHandler_Register_Success(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := postJSON(router, "/users", map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"phone": "+1234",
	})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Alice", resp["name"])
	assert.Equal(t, "alice@example.com", resp["email"])
	assert.NotEmpty(t, resp["id"])
}

func TestUserHandler_Register_MissingName(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := postJSON(router, "/users", map[string]any{
		"email": "alice@example.com",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestUserHandler_Register_InvalidEmail(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := postJSON(router, "/users", map[string]any{
		"name":  "Alice",
		"email": "not-an-email",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestUserHandler_Register_ConflictOnDuplicateEmail(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	body := map[string]any{"name": "Alice", "email": "alice@example.com"}
	postJSON(router, "/users", body) // first registration succeeds
	w := postJSON(router, "/users", body)

	assert.Equal(t, http.StatusConflict, w.Code)
	assertErrorBody(t, w)
}

func TestUserHandler_Register_MalformedJSON(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestUserHandler_GetByID_Success(t *testing.T) {
	repo := mock.NewUserRepo()
	router, _ := setupUserRouter(repo)

	// Register a user first.
	w := postJSON(router, "/users", map[string]any{"name": "Bob", "email": "bob@example.com"})
	require.Equal(t, http.StatusCreated, w.Code)

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	w = getRequest(router, "/users/"+id)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, id, resp["id"])
	assert.Equal(t, "Bob", resp["name"])
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := getRequest(router, "/users/"+uuid.New().String())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorBody(t, w)
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := getRequest(router, "/users/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestUserHandler_List_Empty(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := getRequest(router, "/users")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)
}

func TestUserHandler_List_MultipleUsers(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	postJSON(router, "/users", map[string]any{"name": "Alice", "email": "alice@example.com"})
	postJSON(router, "/users", map[string]any{"name": "Bob", "email": "bob@example.com"})

	w := getRequest(router, "/users")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUserHandler_Update_Success(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := postJSON(router, "/users", map[string]any{"name": "Alice", "email": "alice@example.com"})
	require.Equal(t, http.StatusCreated, w.Code)

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	w = putJSON(router, "/users/"+id, map[string]any{
		"name":  "Alicia",
		"email": "alicia@example.com",
		"phone": "999",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Alicia", resp["name"])
	assert.Equal(t, "alicia@example.com", resp["email"])
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := putJSON(router, fmt.Sprintf("/users/%s", uuid.New()), map[string]any{
		"name":  "X",
		"email": "x@example.com",
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorBody(t, w)
}

func TestUserHandler_Update_InvalidUUID(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := putJSON(router, "/users/bad-id", map[string]any{"name": "X", "email": "x@example.com"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_MissingName(t *testing.T) {
	router, _ := setupUserRouter(mock.NewUserRepo())

	w := postJSON(router, "/users", map[string]any{"name": "Alice", "email": "alice@example.com"})
	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)

	w = putJSON(router, "/users/"+created["id"].(string), map[string]any{
		"email": "alice@example.com",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func assertErrorBody(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body, "error", "response body should contain an 'error' key")
}
