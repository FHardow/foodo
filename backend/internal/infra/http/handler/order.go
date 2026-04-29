package handler

import (
	"net/http"
	"time"

	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/infra/http/respond"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	svc *order.Service
}

func NewOrderHandler(svc *order.Service) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type orderItemResponse struct {
	ProductID      string `json:"product_id"`
	ProductName    string `json:"product_name"`
	Unit           string `json:"unit"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	TotalCents     int64  `json:"total_cents"`
}

type orderResponse struct {
	ID         string              `json:"id"`
	UserID     string              `json:"user_id"`
	UserName   string              `json:"user_name,omitempty"`
	Status     string              `json:"status"`
	Items      []orderItemResponse `json:"items"`
	TotalCents int64               `json:"total_cents"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

func toOrderResponse(o *order.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items()))
	for _, item := range o.Items() {
		items = append(items, orderItemResponse{
			ProductID:      item.ProductID.String(),
			ProductName:    item.ProductName,
			Unit:           item.Unit,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			TotalCents:     item.TotalCents(),
		})
	}
	return orderResponse{
		ID:         o.ID().String(),
		UserID:     o.UserID().String(),
		UserName:   o.UserName(),
		Status:     string(o.Status()),
		Items:      items,
		TotalCents: o.TotalCents(),
		CreatedAt:  o.CreatedAt(),
		UpdatedAt:  o.UpdatedAt(),
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
		return
	}
	o, err := h.svc.Create(c.Request.Context(), userID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusCreated, toOrderResponse(o))
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) List(c *gin.Context) {
	var orders []*order.Order
	var err error

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			respond.Error(c, domerrors.BadRequest("invalid user_id"))
			return
		}
		orders, err = h.svc.ListByUser(c.Request.Context(), userID)
	} else {
		orders, err = h.svc.List(c.Request.Context())
	}
	if err != nil {
		respond.Error(c, err)
		return
	}
	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, toOrderResponse(o))
	}
	respond.JSON(c, http.StatusOK, resp)
}

func (h *OrderHandler) AddItem(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity"   binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	o, err := h.svc.AddItem(c.Request.Context(), orderID, productID, req.Quantity)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) RemoveItem(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	productID, err := uuid.Parse(c.Param("productID"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	o, err := h.svc.RemoveItem(c.Request.Context(), orderID, productID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) Confirm(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.Confirm(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) Accept(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.Accept(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) StartProgress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.StartProgress(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) Finish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.Finish(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) Unaccept(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.Unaccept(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) StopProgress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.StopProgress(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) Unfinish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid order ID"))
		return
	}
	o, err := h.svc.Unfinish(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toOrderResponse(o))
}
