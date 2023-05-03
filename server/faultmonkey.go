package server

import (
	"http-attenuator/data"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServerImpl struct {
	server *data.Server
}

type FaultMonkey interface {
	data.Handler
	ShouldHandle(c *gin.Context) (bool, *data.ServerHost)
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
//
// TODO(john): a simple rules language around host selection would be nice
func (s *ServerImpl) ShouldHandle(c *gin.Context) (bool, *data.ServerHost) {
	// Check for a Host: header match
	serverHost, hasHostMapping := s.server.Hosts[strings.ToLower(c.Request.Host)]
	if !hasHostMapping {
		// Check for the default mapping
		serverHost, hasHostMapping = s.server.Hosts["default"]
	}
	return hasHostMapping, serverHost
}

// a ServerImpl is-a Handler
//
// This is executed as part of gin middleware, so it intercepts the
// various gateway/broker requsts
//
// TODO(john): allow the gateway/broker requests to have intermittent failures
func (s *ServerImpl) Handle(c *gin.Context) {
	shouldHandle, serverHost := s.ShouldHandle(c)
	if !shouldHandle {
		// Not something that FaultMonkey is to deal with
		return
	}

	// Defer to the configured pathology profile
	serverHost.GetPathologyProfile().Handle(c)
}
