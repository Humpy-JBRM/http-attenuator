package util

import "math"

const float64EqualityThreshold = 1e-9

func AlmostEqual(a, b float64) bool {
	diff := math.Abs(a - b)
	return diff <= float64EqualityThreshold
}
