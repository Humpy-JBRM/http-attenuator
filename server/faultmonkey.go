package server

import (
	"fmt"
	"http-attenuator/data"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServerImpl struct {
	server *data.Server
}

type FaultMonkey interface {
	data.Handler
	ShouldHandle(c *gin.Context) bool
}

func NewFaultMonkey(server *data.Server) FaultMonkey {
	return &ServerImpl{
		server: server,
	}
}

func (s *ServerImpl) GetName() string {
	return s.server.Name
}

// ShouldHandle is used by proxy / gateway / broker mode
// to determine whether or not the request is the be handled
// by faultmonkey
func (s *ServerImpl) ShouldHandle(c *gin.Context) bool {
	return false
}

// a ServerImpl is-a Handler
func (s *ServerImpl) Handle(c *gin.Context) {
	// Find the matching host
	// TODO(john): some nice around host selection would be nice
	serverHost := s.server.Hosts[strings.ToLower(c.Request.Host)]
	if serverHost == nil {
		err := fmt.Errorf("server.Handle(%s): no host configured for '%s'", c.Request.URL, c.Request.Host)
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	// Defer to the configured pathology profile
	serverHost.GetPathologyProfile().Handle(c)
}
