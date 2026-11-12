package handler

import (
	"net/http"

	"github.com/fhardow/foodo/internal/domain/category"
	"github.com/fhardow/foodo/internal/infra/http/respond"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	svc *category.Service
}

func NewCategoryHandler(svc *category.Service) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

type categoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	cats, err := h.svc.List(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	resp := make([]categoryResponse, 0, len(cats))
	for _, cat := range cats {
		resp = append(resp, categoryResponse{ID: cat.ID().String(), Name: cat.Name()})
	}
	respond.JSON(c, http.StatusOK, resp)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid category ID"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
