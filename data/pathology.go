package data

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetName() string
	Handle(c *gin.Context)
}

type Pathology interface {
	Handler
	HasCDF
	GetProfileName() string
	SelectResponse() *HttpResponse
}
