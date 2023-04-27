package server

import "github.com/gin-gonic/gin"

type Handler interface {
	GetName() string
	Handle(c *gin.Context)
}

type BaseHandler struct {
	name string
}

func (h *BaseHandler) GetName() string {
	return h.name
}
