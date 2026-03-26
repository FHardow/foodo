package middleware

import (
	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireOwner enforces that the authenticated caller has the owner role.
// It reads the user ID from the Gin context (set by JWTAuth) and looks up
// the user in the database to verify their role.
func RequireOwner(svc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawID, exists := c.Get(UserIDKey)
		if !exists {
			respond.Error(c, domerrors.Forbidden("authentication required"))
			c.Abort()
			return
		}

		id, err := uuid.Parse(rawID.(string))
		if err != nil {
			respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
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
