package data

import (
	"math/rand"
)

// HasCDF is used for selecting items from a slice based on
// a probability cdf
type HasCDF interface {
	CDF() float64
	SetCDF(float64)
	Weight() int
}

func BackpatchCDF(cdf []HasCDF) {
	// Now get the total weight
	// This is the denominator for probability calculations
	var totalWeight float64
	for i := 0; i < len(cdf); i++ {
		totalWeight += float64(cdf[i].Weight())
	}

	// Now backpatch the cdf values
	var totalProbability float64
	for i := 0; i < len(cdf); i++ {
		totalProbability += float64(float64(cdf[i].Weight()) / float64(totalWeight))
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
