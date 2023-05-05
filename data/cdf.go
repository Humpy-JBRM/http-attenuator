package data

import (
	"math/rand"
	"time"
)

// HasCDF is used for selecting items from a slice based on
// a probability cdf
type HasCDF interface {
	CDF() float64
	SetCDF(float64)
	GetWeight() int
}

type HasDuration interface {
	GetDuration() *time.Duration
}

func BackpatchCDF(cdf []HasCDF) {
	// Now get the total weight
	// This is the denominator for probability calculations
	var totalWeight float64
	for i := 0; i < len(cdf); i++ {
		totalWeight += float64(cdf[i].GetWeight())
	}

	// Now backpatch the cdf values
	var totalProbability float64
	for i := 0; i < len(cdf); i++ {
		totalProbability += float64(float64(cdf[i].GetWeight()) / float64(totalWeight))
		cdf[i].SetCDF(totalProbability)
	}
}

func ChooseFromCDF(probability float64, cdf []HasCDF) HasCDF {
	if len(cdf) == 0 {
		return nil
	}
	if len(cdf) == 1 {
		return cdf[0]
	}
	for i := 0; i < len(cdf); i++ {
		if cdf[i].CDF() >= probability || cdf[i].CDF() >= 1.0 {
			return cdf[i]
		}
	}

	// We should never reach here, but it keeps the compiler happy
	return cdf[rand.Intn(len(cdf))]
}

func Choose(rule string, cdf []HasCDF, rng *rand.Rand) HasCDF {
	var choice HasCDF
	switch rule {
	case "weighted":
		choice = ChooseFromCDF(rng.Float64(), cdf)

	case "uniform":
		choice = cdf[rng.Intn(len(cdf))]
		fallthrough

	default:
	}

	return choice
}
