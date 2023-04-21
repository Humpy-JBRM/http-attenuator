package proxy

import (
	"http-attenuator/data"
)

type Proxy interface {
	DoSync(req *data.ProxyRequest) error
}
