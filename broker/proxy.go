package broker

import (
	"context"
	"fmt"
	"http-attenuator/client"
	"http-attenuator/data"
	"log"
	"net/http"
	"time"
)

// ForwardProxy is a simple forward proxy which just plumbs a request
// directly to a backend
type ForwardProxy struct {
	upstream data.Upstream
}

func NewForwardProxy(upstream data.Upstream) *ForwardProxy {
	return &ForwardProxy{
		upstream: upstream,
	}
}

func (p *ForwardProxy) DoSync(ctx context.Context, req *data.GatewayRequest) (*data.GatewayResponse, error) {
	// Defer to the client that deals with all of the attenuation
	// and retry etc

	// Make the request
	httpClient, err := client.NewHttpClientBuilder().Retries(1).TimeoutMillis(1000).Build()
	if err != nil {
		err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), err.Error())
		log.Println(err)
		response := &data.GatewayResponse{
			GatewayBase: data.GatewayBase{
				Id:         req.Id,
				Headers:    req.Headers,
				Body:       req.Body,
				WhenMillis: time.Now().UTC().UnixMilli(),
			},
			StatusCode: http.StatusInternalServerError,
		}

		response.Headers.Add(data.HEADER_X_FAULTMONKEY_ERROR, err.Error())
		return response, err
	}

	log.Printf("ForwardProxy().DoSync(): %s", req.GetRequest().URL.String())
	resp, e := httpClient.Do(ctx, req)
	if err != nil {
		err = fmt.Errorf("ForwardProxy.DoSync(%s): %s", req.GetUrl().String(), e.Error())
		log.Println(err)
		resp.Headers.Add(data.HEADER_X_FAULTMONKEY_ERROR, err.Error())
	}

	return resp, err
}
