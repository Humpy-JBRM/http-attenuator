package broker

import (
	"http-attenuator/data"
)

type BrokeredService interface {
	DoSync(req *data.GatewayRequest) error
}
