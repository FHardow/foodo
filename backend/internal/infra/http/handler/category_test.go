package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhardow/foodo/internal/domain/category"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCategoryRouter(repo *mock.CategoryRepo) *gin.Engine {
	h := handler.NewCategoryHandler(category.NewService(repo))
	r := gin.New()
	r.GET("/categories", h.List)
	r.DELETE("/categories/:id", h.Delete)
	return r
}

func TestCategoryHandler_List(t *testing.T) {
	repo := mock.NewCategoryRepo()
	fruit := repo.Seed("Fruit")
	router := setupCategoryRouter(repo)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/categories", nil))
	require.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	assert.Equal(t, fruit.ID().String(), resp[0]["id"])
	assert.Equal(t, "Fruit", resp[0]["name"])
}

func TestCategoryHandler_List_EmptyIsArray(t *testing.T) {
	router := setupCategoryRouter(mock.NewCategoryRepo())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/categories", nil))
	assert.JSONEq(t, "[]", w.Body.String())
}

func TestCategoryHandler_Delete(t *testing.T) {
	repo := mock.NewCategoryRepo()
	fruit := repo.Seed("Fruit")
	router := setupCategoryRouter(repo)

	w := deleteRequest(router, "/categories/"+fruit.ID().String())
	assert.Equal(t, http.StatusNoContent, w.Code)

	w = deleteRequest(router, "/categories/"+fruit.ID().String())
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryHandler_Delete_InvalidID(t *testing.T) {
	router := setupCategoryRouter(mock.NewCategoryRepo())
	w := deleteRequest(router, "/categories/not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_Categories(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())

	w := postJSON(router, "/products", map[string]any{
		"name": "Rye", "unit": "loaf", "price_cents": 0,
		"categories": []string{"Bakery", " bakery ", "Grains"},
	})
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, []any{"Bakery", "Grains"}, created["categories"])
	id := created["id"].(string)

	// Omitting categories on update keeps them.
	w = putJSON(router, "/products/"+id, map[string]any{"name": "Rye", "unit": "loaf"})
	require.Equal(t, http.StatusOK, w.Code)
	var updated map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, []any{"Bakery", "Grains"}, updated["categories"])

	// Empty array clears them.
	w = putJSON(router, "/products/"+id, map[string]any{"name": "Rye", "unit": "loaf", "categories": []string{}})
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, []any{}, updated["categories"])
}

func TestProductHandler_NoCategoriesIsEmptyArray(t *testing.T) {
	router, _ := setupProductRouter(t, mock.NewProductRepo())
	resp := registerProduct(t, router, "Rye")
	assert.Equal(t, []any{}, resp["categories"])
}
