package evt

import (
	"http-attenuator/data"
	"regexp"
)

// tcp HOST:PORT
var reTCPHealthcheck *regexp.Regexp = regexp.MustCompile("tcp (?P<HostPort>.*)")

// dns HOSTNAME
var reDNSHealthcheck *regexp.Regexp = regexp.MustCompile("dns (?P<HostPort>.*)")

func ParseEVT(healthcheck string) (data.EVT, error) {
	if healthcheck == "" {
		// Not an error
		return nil, nil
	}

	//TODO(john): a mini-DSL instead of regex
	if matches := reDNSHealthcheck.FindAllStringSubmatch(healthcheck, -1); len(matches) > 0 {
		return NewDNSCheck("A", matches[0][1]), nil
	}
	if matches := reTCPHealthcheck.FindAllStringSubmatch(healthcheck, -1); len(matches) > 0 {
		return NewTCPCheck(matches[0][1]), nil
	}
	return nil, nil
}
