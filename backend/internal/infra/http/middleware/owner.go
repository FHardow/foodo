package middleware

import (
	"github.com/fhardow/bread-order/internal/infra/http/respond"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/gin-gonic/gin"
)

// RequireOwner enforces that the authenticated caller has the "owner" realm role in their Keycloak JWT.
func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(RolesKey)
		if !exists {
			respond.Error(c, domerrors.Forbidden("authentication required"))
			c.Abort()
			return
		}

		roles, ok := raw.([]string)
		if !ok {
			respond.Error(c, domerrors.Forbidden("authentication required"))
			c.Abort()
			return
		}

		for _, r := range roles {
			if r == "owner" {
				c.Next()
				return
			}
		}

		respond.Error(c, domerrors.Forbidden("only owners can perform this action"))
		c.Abort()
	}
}
