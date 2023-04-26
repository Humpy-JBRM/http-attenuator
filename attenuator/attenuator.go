package attenuator

import (
	"context"
	"fmt"
	trafficlight "http-attenuator/facade/trafficlight"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var attenuatedRequestsWaitTime = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "attenuated_requests_wait",
		Help:      "The attenuator wait time in millis, keyed by name",
	},
	[]string{"name"},
)

type AttenuatorImpl struct {
	Name        string  `json:"name"`
	MaxHertz    float64 `json:"max_hertz"`
	MaxInflight int     `json:"max_inflight"`
	PulseName   string  `json:"pulse"`
	pulse       trafficlight.Pulse
}

func NewAttenuator(name string, maxHertz float64, maxInflight int) (Attenuator, error) {
	pulse := trafficlight.GetPulse(name)
	var err error
	if pulse == nil {
		// TODO(john): hook this up and encapsulate it behind the pulse
		// factory, so we can use redis
		pulse, err = trafficlight.NewPulse(name, maxInflight, maxHertz)
		if err != nil {
			return nil, err
		}
	}
	if pulse == nil {
		return nil, fmt.Errorf("Unable to get pulse '%s'", name)
	}

	if maxInflight <= 0 {
		return nil, fmt.Errorf("%s: cannot have an attenuator queue size of %d", name, maxInflight)
	}

	a := &AttenuatorImpl{
		Name:        name,
		MaxHertz:    maxHertz,
		MaxInflight: maxInflight,
		PulseName:   name,
		pulse:       pulse,
	}

	return a, nil
}

// func (a *attenuator) DoSync(req *data.GatewayRequest) (*data.GatewayResponse, error) {
// 	// wait for green light
// 	nowMillis := time.Now().UTC().UnixMilli()
// 	trafficlight.WaitForGreen(a.name, 1)
// 	attenuatedRequestsWaiting.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
// 	attenuatedRequests.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()

// 	// Do the request
// 	var err error
// 	cb, err := client.NewHttpClientBuilder().
// 		Retries(0).
// 		TimeoutMillis(10000).
// 		Build()
// 	if err != nil {
// 		// out of switch
// 		attenuatedRequestsFailures.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()
// 		return nil, err
// 	}

// 	resp, err := cb.Do(req)
// 	attenuatedRequestsLatency.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
// 	if err != nil {
// 		attenuatedRequestsFailures.WithLabelValues(req.GetUrl().Host, req.Method, req.GetUrl().Path).Inc()
// 	}
// 	return resp, err
// }

func (a *AttenuatorImpl) String() string {
	return fmt.Sprintf("%s (%.2fHz, %d max)", a.Name, a.MaxHertz, a.MaxInflight)
}

func (a *AttenuatorImpl) GetName() string {
	return a.Name
}

func (a *AttenuatorImpl) GetMaxHertz() float64 {
	return a.MaxHertz
}

func (a *AttenuatorImpl) GetMaxInflight() int {
	return a.MaxInflight
}

// TODO(john)
func (a *AttenuatorImpl) WaitForGreen(ctx context.Context, cancelFunc context.CancelFunc) error {
	if cancelFunc != nil {
		defer cancelFunc()
	}
	if a.pulse == nil {
		return nil
	}

	nowMillis := time.Now().UTC().UnixMilli()
	defer func() {
		attenuatedRequestsWaitTime.WithLabelValues(a.Name).Add(float64(time.Now().UTC().UnixMilli() - nowMillis))
	}()

	errChan := make(chan error, 1)
	defer close(errChan)
	go func() {
		errChan <- a.pulse.WaitForNext()
	}()

	var err error
	select {
	case <-ctx.Done():
		return NewErrWaitTimeout(fmt.Sprintf("%s.WaitForGreen(): timeout", a.Name))
	case err = <-errChan:
		break
	}
	return err
}
