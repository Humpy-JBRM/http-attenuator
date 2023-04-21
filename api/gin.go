package api

import (
	"fmt"
	"http-attenuator/middleware"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter() (*gin.Engine, error) {
	router := gin.Default()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// Make sure we propagate the headers so they can be logged
	router.Use(
		func(ctx *gin.Context) {
			ctx.Set("X-Real-Ip", ctx.Request.Header.Get("X-Real-Ip"))
			ctx.Set("X-Humpy-Api-Key", ctx.Request.Header.Get("X-Humpy-Api-Key"))
		},
	)

	router.SetTrustedProxies([]string{
		"192.168.1.0/24",
		"127.0.0.1",
	})
	router.Use(middleware.CorsMiddleware())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s | %s | %s:%s | %s %s | %d | %d | %s\n",
			param.TimeStamp.Format(time.RFC1123),
			param.ClientIP,
			param.Request.Header.Get("X-Real-Ip"),
			param.Request.Header.Get("X-Humpy-Api-Key"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency.Microseconds(),
			param.ErrorMessage,
		)
	}))

	// Add routes / endpoints here
	// router.NoRoute(ServeCachedFile)
	router.GET("/metrics", prometheusHandler())
	return router, nil
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
