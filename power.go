package main

import (
	"strings"
	"strconv"
	"io/ioutil"
	"time"
)

func powerReadSys(attr string) float64 {
	raw, err := ioutil.ReadFile("/sys/class/power_supply/battery/" + attr)
	check(err)

	attrInt, err := strconv.Atoi(strings.TrimSuffix(string(raw), "\n"))
	check(err)

	return float64(attrInt)
}

func powerMonitor(interval uint, stop chan chan float64) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	var total float64 // mW
	samples := 0

powerLoop:
	for {
		select {
		case <- ticker.C: // check current usage
			microamp := powerReadSys("current_now")
			microvolt := powerReadSys("voltage_now")
			milliamp := microamp / 1000
			millivolt := microvolt / 1000
			
			microwatt := milliamp * millivolt
			milliwatt := microwatt / 1000

			total += milliwatt
			samples++
		case resultChan := <-stop: // stop and send result
			avgMw := total / float64(samples)
			resultChan <- avgMw
			break powerLoop
		}
	}

	ticker.Stop()
}
