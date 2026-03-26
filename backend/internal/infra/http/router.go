package http

import (
	"fmt"
	"net/http"

	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/http/handler"
	"github.com/fhardow/bread-order/internal/infra/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	users *handler.UserHandler,
	products *handler.ProductHandler,
	orders *handler.OrderHandler,
	keycloakURL string,
	keycloakRealm string,
	userSvc *user.Service,
) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakURL, keycloakRealm)
	ownerOnly := middleware.RequireOwner(userSvc)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(jwksURL))
	{
		u := v1.Group("/users")
		u.GET("", users.List)
		u.POST("", users.Register)
		u.GET("/:id", users.GetByID)
		u.PUT("/:id", users.Update)

		p := v1.Group("/products")
		p.GET("", products.List)
		p.GET("/:id", products.GetByID)
		p.POST("", ownerOnly, products.Create)
		p.PUT("/:id", ownerOnly, products.Update)
		p.PATCH("/:id/availability", ownerOnly, products.SetAvailable)
		p.DELETE("/:id", ownerOnly, products.Delete)

		o := v1.Group("/orders")
		o.GET("", orders.List)
		o.POST("", orders.Create)
		o.GET("/:id", orders.GetByID)
		o.POST("/:id/items", orders.AddItem)
		o.DELETE("/:id/items/:productID", orders.RemoveItem)
		o.POST("/:id/confirm", orders.Confirm)
		o.POST("/:id/fulfill", orders.Fulfill)
		o.POST("/:id/cancel", orders.Cancel)
	}

	return r
}
