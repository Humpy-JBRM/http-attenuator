package evt

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"net"
	"time"
)

type DNSCheck struct {
	data.EVTImpl

	// TODO(john): support dns records other than A
	recordType string

	name string
}

func NewDNSCheck(recordType string, name string) data.EVT {
	return &DNSCheck{
		recordType: recordType,
		name:       name,
	}
}

func (d *DNSCheck) GetName() string {
	return fmt.Sprintf("%s-%s", d.GetType(), d.name)
}

func (d *DNSCheck) Run() data.EVTResult {
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
		log.Printf("FAIL %s: %s", d.GetName(), err.Error())
		evtErrors.WithLabelValues(d.GetType(), d.GetName(), data.GetErrorClassifier().Classify(err)).Inc()
		return data.NewEVTResult(
			err,
			time.Now().UTC().UnixMilli()-now,
		)
	}

	if len(ips) == 0 {
		evtErrors.WithLabelValues(d.GetType(), d.GetName(), "notfound").Inc()
		return data.NewEVTResult(
			fmt.Errorf("%s: no matching records", d.name),
			time.Now().UTC().UnixMilli()-now,
		)
	}

	// This test passed
	return data.NewEVTResult(
		nil,
		time.Now().UTC().UnixMilli()-now,
	)
}
