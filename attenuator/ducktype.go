package attenuator

import (
	"context"
	"fmt"
)

type ErrWaitTimeout struct {
	msg string
}

func (e *ErrWaitTimeout) Error() string {
	return e.msg
}

func NewErrWaitTimeout(msg string) *ErrWaitTimeout {
	return &ErrWaitTimeout{
		msg: msg,
	}
}

type Attenuator interface {
	fmt.Stringer
	//DoSync(req *data.GatewayRequest) (*data.GatewayResponse, error)
	GetName() string
	GetMaxHertz() float64
	GetMaxInflight() int
	WaitForGreen(ctx context.Context, cancelFunc context.CancelFunc) error
}
