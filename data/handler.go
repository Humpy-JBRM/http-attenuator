package data

import "github.com/gin-gonic/gin"

type Handler interface {
	GetName() string
	Handle(c *gin.Context)
}

type BaseHandler struct {
	Profile string
	Name    string
}

func (h *BaseHandler) GetProfile() string {
	return h.Profile
}

func (h *BaseHandler) GetName() string {
	return h.Name
}
