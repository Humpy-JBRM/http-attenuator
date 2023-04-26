package gateway

import (
	"http-attenuator/data"
)

type Gateway interface {
	DoSync(req *data.GatewayRequest) error
}

type Pathology interface {
}
