package middleware

import (
	"net/http"

	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireOwner returns a Gin middleware that enforces the caller is an owner.
// It expects an X-User-ID header containing the caller's UUID and looks up the
// user in the database to verify their role.
func RequireOwner(svc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawID := c.GetHeader("X-User-ID")
		if rawID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header is required"})
			c.Abort()
			return
		}

		id, err := uuid.Parse(rawID)
		if err != nil {
			respond.Error(c, domerrors.BadRequest("invalid user ID in X-User-ID header"))
			c.Abort()
			return
		}

		u, err := svc.GetByID(c.Request.Context(), id)
		if err != nil {
			respond.Error(c, err)
			c.Abort()
			return
		}

		if !u.IsOwner() {
			respond.Error(c, domerrors.Forbidden("only owners can perform this action"))
			c.Abort()
			return
		}

		c.Next()
	}
}
