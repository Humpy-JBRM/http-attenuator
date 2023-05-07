package api

import (
	"http-attenuator/broker"

	"github.com/gin-gonic/gin"
)

func BrokerHandler(c *gin.Context) {
	broker.GetServiceBroker().Handle(c)
}
