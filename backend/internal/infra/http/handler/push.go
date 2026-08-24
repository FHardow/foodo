package handler

import (
	"net/http"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/infra/http/respond"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PushHandler struct {
	svc *push.Service
}

func NewPushHandler(svc *push.Service) *PushHandler {
	return &PushHandler{svc: svc}
}

type subscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
}

type unsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
		return
	}
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	if _, err := h.svc.Subscribe(c.Request.Context(), userID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PushHandler) Unsubscribe(c *gin.Context) {
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
		return
	}
	var req unsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	isOwner := middleware.HasRole(c, "owner")
	if err := h.svc.Unsubscribe(c.Request.Context(), userID, isOwner, req.Endpoint); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
