package handler

import (
	"net/http"
	"time"

	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/http/middleware"
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *user.Service
}

func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:        u.ID().String(),
		Name:      u.Name(),
		Email:     u.Email(),
		Phone:     u.Phone(),
		Role:      string(u.Role()),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	id, err := uuid.Parse(sub)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
		return
	}
	var req struct {
		Name  string `json:"name"  binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	u, err := h.svc.Register(c.Request.Context(), id, req.Name, req.Email, req.Phone)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusCreated, toUserResponse(u))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID"))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toUserResponse(u))
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.svc.List(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}
	respond.JSON(c, http.StatusOK, resp)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID"))
		return
	}
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	rolesRaw, _ := c.Get(middleware.RolesKey)
	roles, _ := rolesRaw.([]string)
	isOwner := false
	for _, r := range roles {
		if r == "owner" {
			isOwner = true
			break
		}
	}
	if !isOwner && sub != id.String() {
		respond.Error(c, domerrors.Forbidden("cannot update another user's profile"))
		return
	}
	var req struct {
		Name  string `json:"name"  binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	u, err := h.svc.UpdateContact(c.Request.Context(), id, req.Name, req.Email, req.Phone)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.JSON(c, http.StatusOK, toUserResponse(u))
}
