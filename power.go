package main

import (
	"strconv"
	"time"
)

// PsyPrefix is the path prefix for power supply properties
const PsyPrefix = "/sys/class/power_supply/battery/"

func powerReadSys(attr string) float64 {
	attrStr, err := readFileLine(PsyPrefix + attr)
	check(err)

	attrInt, err := strconv.Atoi(attrStr)
	check(err)

	return float64(attrInt)
}

func powerMonitor(interval uint, stop chan chan float64) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	var totalMw float64 = 0
	samples := 0

powerLoop:
	for {
		select {
		case <-ticker.C: // check current usage
			ma := powerReadSys("current_now") / 1000
			mv := powerReadSys("voltage_now") / 1000
			mw := ma * mv / 1000

			// negate because drain will be negative
			totalMw += -mw
			samples++
		case resultChan := <-stop: // stop and send result
			meanMw := totalMw / float64(samples)
			resultChan <- meanMw
			break powerLoop
		}
	}

	ticker.Stop()
}
