package main

import (
	"github.com/cockroachdb/apd"
)

func floatToDec(f float64) *apd.Decimal {
	dec := apd.New(0, 0)
	dec.SetFloat64(f)
	return dec
}

func finishGeoMean(ctx *apd.Context, nValues int, product *apd.Decimal) (float float64, err error) {
	ed := apd.MakeErrDecimal(ctx)

	/*
	 * Geometric mean = nth root of product
	 * n = nValues
	 * nth root = to the power of 1/n
	 */

	// rootPow = 1 / nValues
	rootPow := apd.New(1, 0)
	ed.Quo(rootPow, rootPow, apd.New(int64(nValues), 0))

	// gmean = product ^ (rootPow)
	ed.Pow(product, product, rootPow)

	// Check for accumulated errors
	err = ed.Err()
	if err != nil {
		return
	}

	// Return as float64
	float, err = product.Float64()
	return
}
