package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc        *product.Service
	uploadsDir string
}

func NewProductHandler(svc *product.Service, uploadsDir string) *ProductHandler {
	return &ProductHandler{svc: svc, uploadsDir: uploadsDir}
}

type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Unit        string    `json:"unit"`
	Available   bool      `json:"available"`
	ImageURL    string    `json:"image_url,omitempty"`
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
		ImageURL:    p.ImageURL(),
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
		Available   bool   `json:"available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req.Name, req.Description, req.PriceCents, req.Unit, req.Available)
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

// UploadImage accepts a multipart upload, saves the file as <product-uuid>.<ext>
// in h.uploadsDir, and updates the product's image_url.
func (h *ProductHandler) UploadImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid product ID"))
		return
	}

	// Confirm product exists before touching the filesystem.
	if _, err := h.svc.GetByID(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respond.Error(c, domerrors.BadRequest("image file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	ext, err := imageExt(contentType)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("unsupported image type: %s", contentType))
		return
	}

	// Remove any existing image for this product (any extension).
	existing, _ := filepath.Glob(filepath.Join(h.uploadsDir, id.String()+".*"))
	for _, f := range existing {
		os.Remove(f) //nolint:errcheck
	}

	filename := id.String() + ext
	dst := filepath.Join(h.uploadsDir, filename)

	data, err := io.ReadAll(file)
	if err != nil {
		respond.Error(c, fmt.Errorf("reading upload: %w", err))
		return
	}
	if err := os.WriteFile(dst, data, 0644); err != nil { //nolint:gosec
		respond.Error(c, fmt.Errorf("saving upload: %w", err))
		return
	}

	p, err := h.svc.SetImageURL(c.Request.Context(), id, "/uploads/"+filename)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.JSON(c, http.StatusOK, toProductResponse(p))
}

func imageExt(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/gif":
		return ".gif", nil
	default:
		return "", fmt.Errorf("unsupported image content-type: %s", contentType)
	}
}
