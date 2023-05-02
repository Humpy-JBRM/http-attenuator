package data

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	NoType int = iota
	Constant
	Uniform
	Normal
)

type DurationConfig struct {
	durationType int

	raw string

	// Constant distribution
	millis int64

	// Uniform distribution
	lower int64
	upper int64

	// Normal / poisson distribution
	mean float64

	// Normal distribution only
	stddev float64

	// a rng used when the distribution is anything except
	// constant
	rng *rand.Rand
}

func (d *DurationConfig) GetDuration() *time.Duration {
	switch d.durationType {
	case Constant:
		duration := time.Duration(d.millis * int64(time.Millisecond))
		return &duration

	case Uniform:
		duration := time.Duration(d.millis * int64(time.Millisecond))
		return &duration

	case Normal:
		duration := time.Duration(d.millis * int64(time.Millisecond))
		return &duration
	}

	return nil
}

// constant values
// 10m
// 10s
// 5000ms
var reConstant = regexp.MustCompile("^(?P<Value>[0-9]+)(?P<Units>[a-z]+)$")

// Uniform distribution
// uniform(10s)			return [0..10000)
// uniform(1000ms)		return [0..1000)
var reUniform = regexp.MustCompile(`^uniform\((?P<Max>[0-9]+)(?P<Units>[a-z]+)\)$`)

// Normal distribution
// normal(10.0, 1.0)		mean and stdev are IN SECONDS
var reNormal = regexp.MustCompile(`^normal\((?P<Mean>[0-9]+.[0-9]+), (?P<Stddev>[0-9]+.[0-9]+)\)$`)

func ParseDuration(durationAsStringX string) (HasDuration, error) {
	trimmed := strings.TrimSpace(durationAsStringX)
	// Try constant value first
	if reConstant.MatchString(trimmed) {
		matches := reConstant.FindStringSubmatch(trimmed)
		if len(matches) != 3 {
			return nil, fmt.Errorf("'%s' does not have positive value and units", trimmed)
		}

		// Validate the units
		valueAsNumber, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("'%s': invalid value: %s", trimmed, matches[1])
		}

		switch strings.ToLower(matches[2]) {
		case "m":
			return &DurationConfig{
				durationType: Constant,
				millis:       valueAsNumber * 1000 * 60,
			}, nil

		case "s":
			return &DurationConfig{
				durationType: Constant,
				millis:       valueAsNumber * 1000,
			}, nil

		case "ms":
			return &DurationConfig{
				durationType: Constant,
				millis:       valueAsNumber,
			}, nil

		default:
			return nil, fmt.Errorf("'%s': unknown units: '%s'", trimmed, matches[2])
		}
	}

	// try uniform(max)
	if reUniform.MatchString(trimmed) {
		matches := reUniform.FindStringSubmatch(trimmed)
		if len(matches) != 3 {
			return nil, fmt.Errorf("'%s' does not have positive value and units", trimmed)
		}

		// Validate the units
		valueAsNumber, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("'%s': invalid value: %s", trimmed, matches[1])
		}

		switch strings.ToLower(matches[2]) {
		case "m":
			return &DurationConfig{
				durationType: Uniform,
				upper:        valueAsNumber * 1000 * 60,
				rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
			}, nil

		case "s":
			return &DurationConfig{
				durationType: Uniform,
				upper:        valueAsNumber * 1000,
				rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
			}, nil

		case "ms":
			return &DurationConfig{
				durationType: Uniform,
				upper:        valueAsNumber,
				rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
			}, nil

		default:
			return nil, fmt.Errorf("'%s': unknown units: '%s'", trimmed, matches[2])
		}
	}

	// try normal(mean, stddev)
	if reNormal.MatchString(trimmed) {
		matches := reNormal.FindStringSubmatch(trimmed)
		if len(matches) != 3 {
			return nil, fmt.Errorf("'%s' invalid normal distribution (must be `normal(mean, stddev)`)", trimmed)
		}

		meanAsNumber, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return nil, fmt.Errorf("'%s': invalid mean: %s", trimmed, matches[1])
		}
		stddevAsNumber, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return nil, fmt.Errorf("'%s': invalid stddev: %s", trimmed, matches[1])
		}
		return &DurationConfig{
			durationType: Normal,
			mean:         meanAsNumber,
			stddev:       stddevAsNumber,
			rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		}, nil
	}

	return nil, fmt.Errorf("'%s' invalid duration", trimmed)
}
