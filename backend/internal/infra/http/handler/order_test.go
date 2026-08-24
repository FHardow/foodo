package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrderRouter(orderRepo *mock.OrderRepo, productRepo *mock.ProductRepo, userRepo *mock.UserRepo, authUserID string) *gin.Engine {
	svc := order.NewService(orderRepo, productRepo, userRepo)
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
	r.POST("/orders/:id/accept", h.Accept)
	r.POST("/orders/:id/start", h.StartProgress)
	r.POST("/orders/:id/finish", h.Finish)
	r.POST("/orders/:id/unaccept", h.Unaccept)
	r.POST("/orders/:id/stop", h.StopProgress)
	r.POST("/orders/:id/unfinish", h.Unfinish)

	return r
}

// seedUser inserts a user into the mock repo and returns their ID string.
// order.Service.Create requires the user to already exist.
func seedUser(t *testing.T, repo *mock.UserRepo) string {
	t.Helper()
	id := uuid.New()
	u, err := user.New(id, "Test User", "test@example.com", "")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), u))
	return id.String()
}

// seedProductInRepo adds a product directly to the product mock repo and returns its ID string.
func seedProductInRepo(t *testing.T, repo *mock.ProductRepo, available bool) string {
	t.Helper()
	p, err := product.New("Sourdough", "test loaf", 450, "loaf", available)
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

// confirmedOrderViaHTTP creates an order, adds one item, and confirms it (status -> created).
func confirmedOrderViaHTTP(t *testing.T, router *gin.Engine, pRepo *mock.ProductRepo) map[string]any {
	t.Helper()
	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo, true)
	postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	w := postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")
	require.Equal(t, http.StatusOK, w.Code, "confirm failed: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// acceptedOrderViaHTTP drives an order through confirm + accept (status -> accepted).
func acceptedOrderViaHTTP(t *testing.T, router *gin.Engine, pRepo *mock.ProductRepo) map[string]any {
	t.Helper()
	confirmed := confirmedOrderViaHTTP(t, router, pRepo)
	w := postEmpty(router, "/orders/"+confirmed["id"].(string)+"/accept")
	require.Equal(t, http.StatusOK, w.Code, "accept failed: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// ongoingOrderViaHTTP drives an order through confirm + accept + start (status -> ongoing).
func ongoingOrderViaHTTP(t *testing.T, router *gin.Engine, pRepo *mock.ProductRepo) map[string]any {
	t.Helper()
	accepted := acceptedOrderViaHTTP(t, router, pRepo)
	w := postEmpty(router, "/orders/"+accepted["id"].(string)+"/start")
	require.Equal(t, http.StatusOK, w.Code, "start failed: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// finishedOrderViaHTTP drives an order through confirm + accept + start + finish (status -> finished).
func finishedOrderViaHTTP(t *testing.T, router *gin.Engine, pRepo *mock.ProductRepo) map[string]any {
	t.Helper()
	ongoing := ongoingOrderViaHTTP(t, router, pRepo)
	w := postEmpty(router, "/orders/"+ongoing["id"].(string)+"/finish")
	require.Equal(t, http.StatusOK, w.Code, "finish failed: %s", w.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func postEmpty(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, nil)
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
	uRepo := mock.NewUserRepo()
	userID := seedUser(t, uRepo)
	router := setupOrderRouter(oRepo, pRepo, uRepo, userID)

	resp := createOrderViaHTTP(t, router)

	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, userID, resp["user_id"])
	assert.Equal(t, "pending", resp["status"])
	assert.Equal(t, float64(0), resp["total_cents"])
}

func TestOrderHandler_Create_MissingUserID(t *testing.T) {
	// Empty string is not a valid UUID — handler should return 400.
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), mock.NewUserRepo(), "")
	w := postEmpty(router, "/orders")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Create_InvalidUserID(t *testing.T) {
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), mock.NewUserRepo(), "not-a-uuid")
	w := postEmpty(router, "/orders")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Create_UserNotRegistered(t *testing.T) {
	// Valid UUID, but never seeded into the user repo.
	router := setupOrderRouter(mock.NewOrderRepo(), mock.NewProductRepo(), mock.NewUserRepo(), uuid.New().String())
	w := postEmpty(router, "/orders")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestOrderHandler_GetByID_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)

	w := getRequest(router, "/orders/"+created["id"].(string))
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, created["id"], resp["id"])
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := getRequest(router, "/orders/"+uuid.New().String())
	assert.Equal(t, http.StatusNotFound, w.Code)
	assertErrorBody(t, w)
}

func TestOrderHandler_GetByID_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := getRequest(router, "/orders/bad-id")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestOrderHandler_List_All(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()

	router1 := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))
	createOrderViaHTTP(t, router1)
	router2 := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))
	createOrderViaHTTP(t, router2)
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := getRequest(router, "/orders")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestOrderHandler_List_ByUserID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()

	userA := seedUser(t, uRepo)
	userB := seedUser(t, uRepo)
	routerA := setupOrderRouter(oRepo, pRepo, uRepo, userA)
	createOrderViaHTTP(t, routerA)
	createOrderViaHTTP(t, routerA)
	routerB := setupOrderRouter(oRepo, pRepo, uRepo, userB)
	createOrderViaHTTP(t, routerB)

	router := setupOrderRouter(oRepo, pRepo, uRepo, userA)
	w := getRequest(router, "/orders?user_id="+userA)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestOrderHandler_List_InvalidUserID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := getRequest(router, "/orders?user_id=not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// AddItem
// ---------------------------------------------------------------------------

func TestOrderHandler_AddItem_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo, true)

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

func TestOrderHandler_AddItem_UnavailableProduct(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo, false)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	// product.ErrUnavailable is a plain error, not a domerrors.DomainError,
	// so respond.Error falls through to its default 500 case.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOrderHandler_AddItem_OrderNotFound(t *testing.T) {
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(mock.NewOrderRepo(), pRepo, uRepo, seedUser(t, uRepo))
	productID := seedProductInRepo(t, pRepo, true)

	w := postJSON(router, "/orders/"+uuid.New().String()+"/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_AddItem_ProductNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)

	w := postJSON(router, "/orders/"+created["id"].(string)+"/items", map[string]any{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_AddItem_InvalidOrderUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postJSON(router, "/orders/bad-id/items", map[string]any{
		"product_id": uuid.New().String(),
		"quantity":   1,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_AddItem_InvalidProductUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

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
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo, true)

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
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)
	productID := seedProductInRepo(t, pRepo, true)

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
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)

	w := deleteRequest(router, fmt.Sprintf("/orders/%s/items/%s", created["id"], uuid.New().String()))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_RemoveItem_InvalidOrderUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := deleteRequest(router, "/orders/bad-id/items/"+uuid.New().String())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_RemoveItem_InvalidProductUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

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
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	resp := confirmedOrderViaHTTP(t, router, pRepo)
	assert.Equal(t, "created", resp["status"])
}

func TestOrderHandler_Confirm_EmptyOrder(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/confirm")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestOrderHandler_Confirm_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/confirm")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Confirm_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/confirm")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Accept
// ---------------------------------------------------------------------------

func TestOrderHandler_Accept_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	resp := acceptedOrderViaHTTP(t, router, pRepo)
	assert.Equal(t, "accepted", resp["status"])
}

func TestOrderHandler_Accept_NotCreated(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	created := createOrderViaHTTP(t, router)

	w := postEmpty(router, "/orders/"+created["id"].(string)+"/accept")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Accept_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/accept")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Accept_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/accept")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// StartProgress
// ---------------------------------------------------------------------------

func TestOrderHandler_StartProgress_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	resp := ongoingOrderViaHTTP(t, router, pRepo)
	assert.Equal(t, "ongoing", resp["status"])
}

func TestOrderHandler_StartProgress_NotAccepted(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	confirmed := confirmedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+confirmed["id"].(string)+"/start")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_StartProgress_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/start")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_StartProgress_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/start")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Finish
// ---------------------------------------------------------------------------

func TestOrderHandler_Finish_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	resp := finishedOrderViaHTTP(t, router, pRepo)
	assert.Equal(t, "finished", resp["status"])
}

func TestOrderHandler_Finish_NotOngoing(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	accepted := acceptedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+accepted["id"].(string)+"/finish")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Finish_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/finish")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Finish_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/finish")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Unaccept
// ---------------------------------------------------------------------------

func TestOrderHandler_Unaccept_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	accepted := acceptedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+accepted["id"].(string)+"/unaccept")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "created", resp["status"])
}

func TestOrderHandler_Unaccept_NotAccepted(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	confirmed := confirmedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+confirmed["id"].(string)+"/unaccept")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Unaccept_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/unaccept")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Unaccept_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/unaccept")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// StopProgress
// ---------------------------------------------------------------------------

func TestOrderHandler_StopProgress_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	ongoing := ongoingOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+ongoing["id"].(string)+"/stop")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "accepted", resp["status"])
}

func TestOrderHandler_StopProgress_NotOngoing(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	accepted := acceptedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+accepted["id"].(string)+"/stop")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_StopProgress_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/stop")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_StopProgress_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/stop")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Unfinish
// ---------------------------------------------------------------------------

func TestOrderHandler_Unfinish_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	finished := finishedOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+finished["id"].(string)+"/unfinish")
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ongoing", resp["status"])
}

func TestOrderHandler_Unfinish_NotFinished(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	ongoing := ongoingOrderViaHTTP(t, router, pRepo)

	w := postEmpty(router, "/orders/"+ongoing["id"].(string)+"/unfinish")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Unfinish_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/"+uuid.New().String()+"/unfinish")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_Unfinish_InvalidUUID(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	router := setupOrderRouter(oRepo, pRepo, uRepo, seedUser(t, uRepo))

	w := postEmpty(router, "/orders/bad-id/unfinish")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
