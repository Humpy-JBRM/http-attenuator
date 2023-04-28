package data

import (
	"math"
	"testing"
)

func TestParseDurationConstantHappy(t *testing.T) {
	durationAsString := "5000ms"
	dc, err := parseDuration(durationAsString)
	if err != nil {
		t.Fatal(err)
	}

	if dc.(*DurationConfig).millis != 5000 {
		t.Errorf("Expected 5000ms, but got %d", dc.(*DurationConfig).millis)
	}
	if dc.(*DurationConfig).durationType != Constant {
		t.Error("Expected 'Constant' duration")
	}
	if dc.(*DurationConfig).rng != nil {
		t.Error("A constant distribution should not have a rng")
	}
}

func TestParseDurationConstantNegativeValue(t *testing.T) {
	durationAsString := "-5000ms"
	_, err := parseDuration(durationAsString)
	if err == nil {
		t.Fatal("Expected an error because of -ve duration")
	}
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	diff := math.Abs(a - b)
	return diff <= float64EqualityThreshold
}

func TestParseDurationNormalHappy(t *testing.T) {
	durationAsString := "normal(10.0, 1.0)"
	dc, err := parseDuration(durationAsString)
	if err != nil {
		t.Fatal(err)
	}

	if dc.(*DurationConfig).durationType != Normal {
		t.Error("Expected 'Normal' duration")
	}
	if !almostEqual(dc.(*DurationConfig).mean, 10.0) {
		t.Errorf("Expected mean=10.0, but got %f", dc.(*DurationConfig).mean)
	}
	if !almostEqual(dc.(*DurationConfig).stddev, 1.0) {
		t.Errorf("Expected stddev=1.0, but got %f", dc.(*DurationConfig).stddev)
	}
	if dc.(*DurationConfig).rng == nil {
		t.Error("A normal distribution should have a rng")
	}
}

func TestParseDurationUniformHappy(t *testing.T) {
	durationAsString := "uniform(5s)"
	dc, err := parseDuration(durationAsString)
	if err != nil {
		t.Fatal(err)
	}

	if dc.(*DurationConfig).durationType != Uniform {
		t.Error("Expected 'Uniform' duration")
	}
	if dc.(*DurationConfig).upper != 5000 {
		t.Errorf("Expected upper=5000, but got %d", dc.(*DurationConfig).upper)
	}
	if dc.(*DurationConfig).rng == nil {
		t.Error("A uniform distribution should have a rng")
	}
}
