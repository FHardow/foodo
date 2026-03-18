package handler

import (
	"net/http"
	"time"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc *product.Service
}

func NewProductHandler(svc *product.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Unit        string    `json:"unit"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toProductResponse(p *product.Product) productResponse {
	return productResponse{
		ID:          p.ID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		PriceCents:  p.PriceCents(),
		Unit:        p.Unit(),
		Available:   p.Available(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name"        binding:"required"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents" binding:"min=0"`
		Unit        string `json:"unit"        binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req.Name, req.Description, req.PriceCents, req.Unit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusCreated, toProductResponse(p))
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) List(c *gin.Context) {
	availableOnly := c.Query("available") == "true"
	products, err := h.svc.List(c.Request.Context(), availableOnly)
	if err != nil {
		respond.Error(c, err)
		return
	}
	resp := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, toProductResponse(p))
	}
	respond.JSON(c, http.StatusOK, resp)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	var req struct {
		Name        string `json:"name"        binding:"required"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents" binding:"min=0"`
		Unit        string `json:"unit"        binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, req.Name, req.Description, req.PriceCents, req.Unit)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) SetAvailable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	var req struct {
		Available bool `json:"available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.SetAvailable(c.Request.Context(), id, req.Available)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
