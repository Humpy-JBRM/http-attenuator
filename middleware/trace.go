package middleware

import (
	"http-attenuator/data"

	"github.com/gin-gonic/gin"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get(data.HEADER_X_REQUEST_ID) != "" {
			c.Writer.Header().Add(data.HEADER_X_REQUEST_ID, c.Request.Header.Get(data.HEADER_X_REQUEST_ID))
		}
	}
}
