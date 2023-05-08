package evt

import (
	"fmt"
	"net"
	"time"
)

type DNSCheck struct {
	EVTImpl

	// TODO(john): support dns records other than A
	recordType string

	name string
}

func NewDNSCheck(recordType string, name string) EVT {
	return &DNSCheck{
		EVTImpl: EVTImpl{},
	}
}

func (d *DNSCheck) GetName() string {
	return fmt.Sprintf("%s: %s", d.evtType, d.name)
}

func (d *DNSCheck) Run() EVTResult {
	// Do the lookup
	// TODO(john): records other than A / AAAA
	now := time.Now().UTC().UnixMilli()
	evtExecuted.WithLabelValues(d.GetType(), d.GetName()).Inc()
	defer func() {
		evtDuration.WithLabelValues(d.GetType(), d.GetName()).Add(
			float64(time.Now().UTC().UnixMilli() - now),
		)
	}()
	ips, err := net.LookupIP(d.name)
	if err != nil {
		evtErrors.WithLabelValues(d.GetType(), d.GetName(), GetErrorClassifier().Classify(err)).Inc()
		return NewEVTResult(
			err,
			time.Now().UTC().UnixMilli()-now,
		)
	}

	if len(ips) == 0 {
		evtErrors.WithLabelValues(d.GetType(), d.GetName(), "notfound").Inc()
		return NewEVTResult(
			fmt.Errorf("%s: no matching records", d.name),
			time.Now().UTC().UnixMilli()-now,
		)
	}

	// This test passed
	return NewEVTResult(
		nil,
		time.Now().UTC().UnixMilli()-now,
	)
}
