package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type execResult struct {
	duration time.Duration
	out      []byte
}

type resultCache map[uint64]execResult

// Run the benchmark and calculate a normalized score from the results
func (bm *Benchmark) Run(cache resultCache) (score float64, rawValue float64, duration time.Duration, err error) {
	var out []byte

	hash := bm.HashCmd()
	if result, cached := cache[hash]; cached {
		out = result.out
		duration = result.duration
	} else {
		// Use bundled program if possible, otherwise resort to system PATH
		progPath := "subtests/" + runtime.GOARCH + "/" + bm.Program
		if _, err = os.Stat(progPath); err != nil {
			if !os.IsNotExist(err) { // Not existing is normal, anything else is a warning
				fmt.Fprintf(os.Stderr, "Unable to stat '%s': %v; resorting to system PATH\n", progPath, err)
			}

			progPath = bm.Program
		}

		before := time.Now()
		out, err = exec.Command(progPath, bm.Arguments...).CombinedOutput()
		duration = time.Since(before)

		cache[hash] = execResult{
			out:      out,
			duration: duration,
		}
	}

	matches := bm.Pattern.FindSubmatch(out)
	if len(matches) < 2 {
		fmt.Print("\n")
		fmt.Println(string(out))
		err = fmt.Errorf("benchmark '%s': Output of %s doesn't match expected format", bm.Name, bm.Program)
		return
	}

	rawValue, err = strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		return
	}

	if bm.ValueFilter != nil {
		rawValue = bm.ValueFilter(rawValue)
	}

	// Only calculate score if a reference is available
	if bm.RefValue != 0 {
		// Normalize to reference
		if bm.HigherIsBetter {
			score = rawValue / bm.RefValue
		} else {
			score = bm.RefValue / rawValue
		}
		// Scale up to the target score
		score *= RefScore
		// Set a minimum bound of 0
		if score < 0 {
			score = 0
		}
	}

	return
}

// HashCmd returns a 64-bit FNV-1a hash of this Benchmark's command.
func (bm *Benchmark) HashCmd() uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(bm.Program))
	for _, arg := range bm.Arguments {
		hash.Write([]byte(arg))
	}

	return hash.Sum64()
}

func getMaxBmNameLen() (max int) {
	for _, bm := range benchmarks {
		nl := len(bm.Name)
		if nl > max {
			max = nl
		}
	}

	return
}

func runBenchmarks(trials int, speed Speed, monitorPower bool, powerInterval uint) {
	stopChan := make(chan chan float64)
	if monitorPower {
		go powerMonitor(powerInterval, stopChan)
	}

	// For calculating spaces
	maxBmNameLen := getMaxBmNameLen()

	var allScores float64
	beforeTrials := time.Now()
	for trial := 0; trial < trials; trial++ {
		fmt.Printf("Trial %d:\n", trial+1)

		cache := make(resultCache, len(benchmarks))
		var trialScoreState float64 // accumulating counter for geometric mean
		bmExecuted := 0

		for _, bm := range benchmarks {
			if bm.Speed < speed {
				continue
			}

			fmt.Printf("  %s: ", bm.Name)
			benchScore, rawValue, duration, err := bm.Run(cache)
			check(err)

			spaces := strings.Repeat(" ", maxBmNameLen-len(bm.Name))
			fmt.Printf("%s%7.0f  (%9.1f %4s; runtime: %3s)\n", spaces, benchScore, rawValue, bm.Unit, formatDuration(duration))

			trialScoreState = addGeoMean(trialScoreState, benchScore)
			bmExecuted++
		}

		trialScore := finishGeoMean(trialScoreState, bmExecuted)
		fmt.Printf("Score: %.0f\n\n", trialScore)
		allScores += trialScore
	}
	benchTime := time.Since(beforeTrials)

	var avgPowerMw float64
	if monitorPower {
		powerChan := make(chan float64, 1)
		stopChan <- powerChan
		avgPowerMw = <-powerChan
	}

	avgScore := allScores / float64(trials)

	/* Output results */
	fmt.Printf("\nAverage score: %.0f\n", avgScore)
	if monitorPower {
		fmt.Printf("Average power usage: %.0f mW\n", avgPowerMw)
		fmt.Printf("Energy usage: %.0f mWh\n", avgPowerMw*benchTime.Hours())
	}
	fmt.Printf("Time elapsed: %s\n", formatDuration(benchTime))
}
