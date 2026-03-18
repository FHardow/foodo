package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
)

func JSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

func Error(c *gin.Context, err error) {
	type errBody struct {
		Error string `json:"error"`
	}
	switch {
	case domerrors.Is(err, domerrors.ErrNotFound):
		c.JSON(http.StatusNotFound, errBody{err.Error()})
	case domerrors.Is(err, domerrors.ErrConflict):
		c.JSON(http.StatusConflict, errBody{err.Error()})
	case domerrors.Is(err, domerrors.ErrForbidden):
		c.JSON(http.StatusForbidden, errBody{err.Error()})
	case domerrors.Is(err, domerrors.ErrBadRequest):
		c.JSON(http.StatusBadRequest, errBody{err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errBody{"internal server error"})
	}
}
