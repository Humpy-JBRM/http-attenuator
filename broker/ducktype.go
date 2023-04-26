package broker

import (
	"http-attenuator/data"
)

type BrokeredServiceImpl struct {
	Name     string    `json:"name"`
	Rule     string    `json:"rule"`
	Backends []Backend `json:"backends"`
}

type BrokeredService interface {
	DoSync(req *data.GatewayRequest) error
}
