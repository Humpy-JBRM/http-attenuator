package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TimeoutHandler struct {
	BaseHandler
	timeoutMillis int64
}

func NewTimeoutHandler(name string, timeoutMillis int64) Handler {
	return &TimeoutHandler{
		BaseHandler: BaseHandler{
			name: name,
		},
		timeoutMillis: timeoutMillis,
	}
}

// TimeoutHandler sleeps for a given time before responding,
// or it sleeps forever and never returns.
//
// It is used to simulate servers that are slow or which
// otherwise time out
func (h *TimeoutHandler) Handle(c *gin.Context) {
	if h.timeoutMillis <= 0 {
		// Sleep forever.  Never returns
		log.Printf("%s.Handle(): sleep forever", h.name)
		select {}
	}

	log.Printf("%s.Handle(): sleep for %dms", h.name, h.timeoutMillis)
	time.Sleep(time.Duration(h.timeoutMillis))

	// TODO(john): set the response to return in config
	c.Status(http.StatusOK)
}
