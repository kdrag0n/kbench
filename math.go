package main

import (
	"math"
)

// This exists in case this math gets more complicated in the future
func addGeoMean(counter, value float64) float64 {
	return counter + math.Log(value)
}

func finishGeoMean(counter float64, nValues int) float64 {
	return math.Exp(counter / float64(nValues))
}
