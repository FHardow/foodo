package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/infra/http/handler"
	"github.com/fhardow/bread-order/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProductRouter(repo *mock.ProductRepo) *gin.Engine {
	svc := product.NewService(repo)
	h := handler.NewProductHandler(svc)

	r := gin.New()
	r.POST("/products", h.Create)
	r.GET("/products", h.List)
	r.GET("/products/:id", h.GetByID)
	r.PUT("/products/:id", h.Update)
	r.PATCH("/products/:id/availability", h.SetAvailable)
	r.DELETE("/products/:id", h.Delete)

	return r
}

func deleteRequest(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func patchJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// registerProduct is a test helper that creates a product through the handler
// and returns the parsed response map.
func registerProduct(t *testing.T, router *gin.Engine, name string) map[string]any {
	t.Helper()
	w := postJSON(router, "/products", map[string]any{
		"name":        name,
		"description": "test product",
		"price_cents": 450,
		"unit":        "loaf",
	})
	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestProductHandler_Create_Success(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := postJSON(router, "/products", map[string]any{
		"name":        "Sourdough",
		"description": "classic loaf",
		"price_cents": 450,
		"unit":        "loaf",
	})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Sourdough", resp["name"])
	assert.Equal(t, float64(450), resp["price_cents"])
	assert.Equal(t, true, resp["available"])
	assert.NotEmpty(t, resp["id"])
}

func TestProductHandler_Create_MissingName(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := postJSON(router, "/products", map[string]any{
		"price_cents": 100,
		"unit":        "loaf",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestProductHandler_Create_MissingUnit(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := postJSON(router, "/products", map[string]any{
		"name":        "Bread",
		"price_cents": 100,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_Create_NegativePriceRejectedByBinding(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	// gin's binding:"min=0" rejects negative values at the binding layer.
	w := postJSON(router, "/products", map[string]any{
		"name":        "Bread",
		"price_cents": -1,
		"unit":        "loaf",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_Create_MalformedJSON(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("{broken"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestProductHandler_GetByID_Success(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Rye")

	w := getRequest(router, "/products/"+created["id"].(string))
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Rye", resp["name"])
}

func TestProductHandler_GetByID_NotFound(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := getRequest(router, "/products/"+uuid.New().String())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorBody(t, w)
}

func TestProductHandler_GetByID_InvalidUUID(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := getRequest(router, "/products/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestProductHandler_List_All(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	registerProduct(t, router, "A")
	registerProduct(t, router, "B")

	w := getRequest(router, "/products")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestProductHandler_List_AvailableOnly(t *testing.T) {
	repo := mock.NewProductRepo()
	router := setupProductRouter(repo)

	p1 := registerProduct(t, router, "Available")
	p2 := registerProduct(t, router, "Unavailable")

	// Mark p2 unavailable.
	patchJSON(router, fmt.Sprintf("/products/%s/availability", p2["id"]), map[string]any{"available": false})
	_ = p1

	w := getRequest(router, "/products?available=true")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
}

func TestProductHandler_List_Empty(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := getRequest(router, "/products")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestProductHandler_Update_Success(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")

	w := putJSON(router, "/products/"+created["id"].(string), map[string]any{
		"name":        "Whole Wheat",
		"description": "updated",
		"price_cents": 600,
		"unit":        "kg",
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Whole Wheat", resp["name"])
	assert.Equal(t, float64(600), resp["price_cents"])
}

func TestProductHandler_Update_NotFound(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := putJSON(router, "/products/"+uuid.New().String(), map[string]any{
		"name":        "X",
		"price_cents": 100,
		"unit":        "loaf",
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProductHandler_Update_InvalidUUID(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := putJSON(router, "/products/bad-id", map[string]any{"name": "X", "unit": "loaf"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_Update_MissingName(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")

	w := putJSON(router, "/products/"+created["id"].(string), map[string]any{
		"price_cents": 100,
		"unit":        "loaf",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// SetAvailable
// ---------------------------------------------------------------------------

func TestProductHandler_SetAvailable_ToFalse(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")

	w := patchJSON(router, "/products/"+created["id"].(string)+"/availability", map[string]any{
		"available": false,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["available"])
}

func TestProductHandler_SetAvailable_ToTrue(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")
	id := created["id"].(string)

	patchJSON(router, "/products/"+id+"/availability", map[string]any{"available": false})

	w := patchJSON(router, "/products/"+id+"/availability", map[string]any{"available": true})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["available"])
}

func TestProductHandler_SetAvailable_NotFound(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := patchJSON(router, "/products/"+uuid.New().String()+"/availability", map[string]any{"available": false})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProductHandler_SetAvailable_InvalidUUID(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := patchJSON(router, "/products/bad/availability", map[string]any{"available": false})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestProductHandler_Delete_Success(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())
	created := registerProduct(t, router, "Sourdough")
	id := created["id"].(string)

	w := deleteRequest(router, "/products/"+id)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Confirm it's gone.
	w = getRequest(router, "/products/"+id)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProductHandler_Delete_InvalidUUID(t *testing.T) {
	router := setupProductRouter(mock.NewProductRepo())

	w := deleteRequest(router, "/products/bad-id")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
