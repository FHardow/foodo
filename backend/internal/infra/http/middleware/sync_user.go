package middleware

import (
	"log/slog"

	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SyncUser upserts the authenticated user into the local users table on every
// request, using name and email from the JWT claims. This ensures the users
// table stays in sync with the identity provider without a separate registration step.
func SyncUser(svc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		subRaw, _ := c.Get(UserIDKey)
		sub, _ := subRaw.(string)
		name, _ := c.Get(UserNameKey)
		email, _ := c.Get(UserEmailKey)

		id, err := uuid.Parse(sub)
		if err != nil || name == "" || email == "" {
			c.Next()
			return
		}

		if err := svc.Upsert(c.Request.Context(), id, name.(string), email.(string)); err != nil {
			slog.Warn("failed to sync user from token", "user_id", sub, "err", err)
		}

		c.Next()
	}
}
