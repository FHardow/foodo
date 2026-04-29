package http

import (
	"fmt"
	"net/http"

	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	users *handler.UserHandler,
	products *handler.ProductHandler,
	orders *handler.OrderHandler,
	userSvc *user.Service,
	keycloakURL string,
	keycloakRealm string,
	uploadsDir string,
	corsOrigin string,
) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	if corsOrigin != "" {
		r.Use(cors.New(cors.Config{
			AllowOrigins: []string{corsOrigin},
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		}))
	}

	// Serve uploaded images without auth — URLs are UUID-based and not guessable.
	r.Static("/uploads", uploadsDir)

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, keycloakRealm)
	ownerOnly := middleware.RequireOwner()

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(jwksURL), middleware.SyncUser(userSvc))
	{
		u := v1.Group("/users")
		u.GET("", users.List)
		u.POST("", users.Register)
		u.GET("/me", users.Me)
		u.GET("/:id", users.GetByID)
		u.PUT("/:id", users.Update)

		p := v1.Group("/products")
		p.GET("", products.List)
		p.GET("/:id", products.GetByID)
		p.POST("", ownerOnly, products.Create)
		p.PUT("/:id", ownerOnly, products.Update)
		p.PATCH("/:id/availability", ownerOnly, products.SetAvailable)
		p.DELETE("/:id", ownerOnly, products.Delete)
		p.POST("/:id/image", ownerOnly, products.UploadImage)

		o := v1.Group("/orders")
		o.GET("", orders.List)
		o.POST("", orders.Create)
		o.GET("/:id", orders.GetByID)
		o.POST("/:id/items", orders.AddItem)
		o.DELETE("/:id/items/:productID", orders.RemoveItem)
		o.POST("/:id/confirm", orders.Confirm)
		o.POST("/:id/accept", ownerOnly, orders.Accept)
		o.POST("/:id/start", ownerOnly, orders.StartProgress)
		o.POST("/:id/finish", ownerOnly, orders.Finish)
		o.POST("/:id/unaccept", ownerOnly, orders.Unaccept)
		o.POST("/:id/stop", ownerOnly, orders.StopProgress)
		o.POST("/:id/unfinish", ownerOnly, orders.Unfinish)
	}

	return r
}
