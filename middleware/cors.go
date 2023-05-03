package middleware

import (
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Add("Access-Control-Allow-Headers", "X-Forwarded-For,X-Requested-With,X-Auth-Token,X-Faultmonkey-Api-Key,X-Request-Id,X-faultmonkey-tag,Content-Type,Content-Length,Authorization")
		c.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD, CONNECT")
		c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	}
}
