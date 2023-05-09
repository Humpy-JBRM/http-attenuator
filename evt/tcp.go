package evt

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"net"
	"time"
)

type TCPCheck struct {
	data.EVTImpl

	// TODO(john): support tcp records other than A
	hostPort string
}

func NewTCPCheck(hostPort string) data.EVT {
	return &TCPCheck{
		hostPort: hostPort,
	}
}

func (e *TCPCheck) GetName() string {
	return fmt.Sprintf("%s-%s", e.GetType(), e.hostPort)
}

func (e *TCPCheck) Run() data.EVTResult {
	// Do the lookup
	// TODO(john): records other than A / AAAA
	now := time.Now().UTC().UnixMilli()
	evtExecuted.WithLabelValues(e.GetType(), e.GetName()).Inc()
	defer func() {
		evtDuration.WithLabelValues(e.GetType(), e.GetName()).Add(
			float64(time.Now().UTC().UnixMilli() - now),
		)
	}()

	// Do the TCP connect
	conn, err := net.Dial("tcp", e.hostPort)
	if err != nil {
		log.Printf("FAIL %s: %s", e.GetName(), err.Error())
		evtErrors.WithLabelValues(e.GetType(), e.GetName(), data.GetErrorClassifier().Classify(err)).Inc()
		return data.NewEVTResult(
			err,
			time.Now().UTC().UnixMilli()-now,
		)
	}
	conn.Close()

	// This test passed
	return data.NewEVTResult(
		nil,
		time.Now().UTC().UnixMilli()-now,
	)
}
