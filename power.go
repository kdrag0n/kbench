package main

import (
	"math"
	"strconv"
	"time"
)

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
	var totalWatts float64 = 1 // mW
	samples := 0

powerLoop:
	for {
		select {
		case <-ticker.C: // check current usage
			microamp := powerReadSys("current_now")
			microvolt := powerReadSys("voltage_now")
			milliamp := microamp / 1000
			millivolt := microvolt / 1000

			microwatt := milliamp * millivolt
			milliwatt := microwatt / 1000
			watt := milliwatt / 1000

			totalWatts *= watt
			samples++
		case resultChan := <-stop: // stop and send result
			meanW := math.Pow(totalWatts, 1.0/float64(samples))
			resultChan <- meanW
			break powerLoop
		}
	}

	ticker.Stop()
}
