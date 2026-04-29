package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrderRouter(orderRepo *mock.OrderRepo, productRepo *mock.ProductRepo, authUserID string) *gin.Engine {
	svc := order.NewService(orderRepo, productRepo)
	h := handler.NewOrderHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, authUserID)
		c.Next()
	})
	r.POST("/orders", h.Create)
	r.GET("/orders", h.List)
	r.GET("/orders/:id", h.GetByID)
	r.POST("/orders/:id/items", h.AddItem)
	r.DELETE("/orders/:id/items/:productID", h.RemoveItem)
	r.POST("/orders/:id/confirm", h.Confirm)
	r.POST("/orders/:id/fulfill", h.Fulfill)
	r.POST("/orders/:id/cancel", h.Cancel)

	return r
}

// seedProductInRepo adds a product directly to the product mock repo and returns its ID string.
func seedProductInRepo(t *testing.T, repo *mock.ProductRepo) string {
	t.Helper()
	p, err := product.New("Sourdough", "test loaf", 450, "loaf", true)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), p))
	return p.ID().String()
}

// createOrderViaHTTP creates an order through the handler and returns the response body map.
func createOrderViaHTTP(t *testing.T, router *gin.Engine) map[string]any {
	t.Helper()
	w := postEmpty(router, "/orders")
	require.Equal(t, http.StatusCreated, w.Code, "create order failed: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func postEmpty(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestOrderHandler_Create_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	userID := uuid.New().String()
	router := setupOrderRouter(oRepo, pRepo, userID)

	resp := createOrderViaHTTP(t, router)

	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, userID, resp["user_id"])
	assert.Equal(t, "pending", resp["status"])
	assert.Equal(t, float64(0), resp["total_cents"])
}

func TestOrderHandler_Create_MissingUserID(t *testing.T) {
	// Empty string is not a valid UUID — handler should return 400.
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), "")
	w := postEmpty(router, "/orders")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Create_InvalidUserID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), "not-a-uuid")
	w := postEmpty(router, "/orders")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestOrderHandler_GetByID_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := getRequest(router, "/orders/"+created["id"].(string))
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, created["id"], resp["id"])
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := getRequest(router, "/orders/"+uuid.New().String())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorBody(t, w)
}

func TestOrderHandler_GetByID_InvalidUUID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := getRequest(router, "/orders/bad-id")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestOrderHandler_List_All(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()

	router1 := setupOrderRouter(oRepo, pRepo, uuid.New().String())
	createOrderViaHTTP(t, router1)
	router2 := setupOrderRouter(oRepo, pRepo, uuid.New().String())
	createOrderViaHTTP(t, router2)
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	w := getRequest(router, "/orders")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestOrderHandler_List_ByUserID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()

	userA := uuid.New().String()
	userB := uuid.New().String()
	routerA := setupOrderRouter(oRepo, pRepo, userA)
	createOrderViaHTTP(t, routerA)
	createOrderViaHTTP(t, routerA)
	routerB := setupOrderRouter(oRepo, pRepo, userB)
	createOrderViaHTTP(t, routerB)

	router := setupOrderRouter(oRepo, pRepo, userA)
	w := getRequest(router, "/orders?user_id="+userA)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestOrderHandler_List_InvalidUserID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := getRequest(router, "/orders?user_id=not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// AddItem
// ---------------------------------------------------------------------------

func TestOrderHandler_AddItem_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   2,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	items := resp["items"].([]any)
	require.Len(t, items, 1)
	item := items[0].(map[string]any)
	assert.Equal(t, float64(2), item["quantity"])
	assert.Equal(t, float64(900), resp["total_cents"]) // 2 * 450
}

func TestOrderHandler_AddItem_OrderNotFound(t *testing.T) {
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(mock.NewOrderRepo(), pRepo, uuid.New().String())
	productID := seedProductInRepo(t, pRepo)

	w := postJSON(router, "/orders/"+uuid.New().String()+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_AddItem_ProductNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_AddItem_InvalidOrderUUID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postJSON(router, "/orders/bad-id/items", map[string]any{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_AddItem_InvalidProductUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": "not-a-uuid",
		"quantity":   1,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_AddItem_ZeroQuantityRejected(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   0,
	})
	// gin binding:"min=1" rejects this at the binding layer.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// RemoveItem
// ---------------------------------------------------------------------------

func TestOrderHandler_RemoveItem_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)

	// Add item first.
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})

	w := deleteRequest(router, fmt.Sprintf("/orders/%s/items/%s", created["id"], productID))
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	items := resp["items"].([]any)
	assert.Empty(t, items)
	assert.Equal(t, float64(0), resp["total_cents"])
}

func TestOrderHandler_RemoveItem_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := deleteRequest(router, fmt.Sprintf("/orders/%s/items/%s", created["id"], uuid.New().String()))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_RemoveItem_InvalidOrderUUID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := deleteRequest(router, "/orders/bad-id/items/"+uuid.New().String())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_RemoveItem_InvalidProductUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	w := deleteRequest(router, "/orders/"+created["id"].(string)+"/items/bad-id")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Confirm
// ---------------------------------------------------------------------------

func TestOrderHandler_Confirm_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "confirmed", resp["status"])
}

func TestOrderHandler_Confirm_EmptyOrder(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	router := setupOrderRouter(oRepo, mock.NewProductRepo(), uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestOrderHandler_Confirm_NotFound(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/confirm")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Confirm_InvalidUUID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postEmpty(router, "/orders/bad-id/confirm")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Fulfill
// ---------------------------------------------------------------------------

func TestOrderHandler_Fulfill_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/fulfill")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "fulfilled", resp["status"])
}

func TestOrderHandler_Fulfill_NotConfirmed(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	router := setupOrderRouter(oRepo, mock.NewProductRepo(), uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/fulfill")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Fulfill_NotFound(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/fulfill")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

func TestOrderHandler_Cancel_FromPending(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	router := setupOrderRouter(oRepo, mock.NewProductRepo(), uuid.New().String())

	created := createOrderViaHTTP(t, router)

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/cancel")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cancelled", resp["status"])
}

func TestOrderHandler_Cancel_FromConfirmed(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/cancel")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cancelled", resp["status"])
}

func TestOrderHandler_Cancel_FromFulfilled(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	router := setupOrderRouter(oRepo, pRepo, uuid.New().String())

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo)
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")
	postEmpty(router, "/orders/"+created["id"].(string)+"/fulfill")

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/cancel")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Cancel_NotFound(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/cancel")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Cancel_InvalidUUID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), uuid.New().String())

	w := postEmpty(router, "/orders/bad-id/cancel")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
