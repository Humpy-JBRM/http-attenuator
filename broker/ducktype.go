package broker

import (
	"http-attenuator/data"
)

type Gateway interface {
	DoSync(req *data.GatewayRequest) error
}