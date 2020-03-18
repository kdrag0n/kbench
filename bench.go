package main

import (
	"hash/fnv"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/cockroachdb/apd"
)

// Run the microbenchmark and calculate a normalized score from the results
func (mb *Microbenchmark) Run(resultCache map[uint64][]byte) (score float64, rawValue float64, cached bool, err error) {
	var out []byte
	var progPath string
	var hash uint64
	foundHash := false

	if mb.CacheOutput {
		hash = mb.HashCmd()
		if out, foundHash = resultCache[hash]; foundHash {
			cached = true
			goto skipRun
		}
	}

	// Use bundled program if possible, otherwise resort to system PATH
	progPath = "subtests/" + runtime.GOARCH + "/" + mb.Program
	if _, err = os.Stat(progPath); err != nil {
		if !os.IsNotExist(err) { // Not existing is normal, anything else is a warning
			fmt.Fprintf(os.Stderr, "Unable to stat '%s': %v; resorting to system PATH\n", progPath, err)
		}

		progPath = mb.Program
	}
	out, err = exec.Command(progPath, mb.Arguments...).CombinedOutput()

skipRun:
	matches := mb.Pattern.FindSubmatch(out)
	if len(matches) < 2 {
		fmt.Print("\n")
		fmt.Println(string(out))
		err = fmt.Errorf("microbenchmark '%s': Output of %s does not match expected format", mb.Name, mb.Program)
		return
	}

	rawValue, err = strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		return
	}

	score = rawValue * mb.Factor
	if !mb.MoreIsBetter {
		score = 1000 - score
	}
	if score < 0 {
		score = 0
	}

	if mb.CacheOutput && !foundHash {
		resultCache[hash] = out
	}

	return
}

// HashCmd returns a 64-bit FNV-1a hash of this Microbenchmark's command.
func (mb *Microbenchmark) HashCmd() uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(mb.Program))
	for _, arg := range mb.Arguments {
		hash.Write([]byte(arg))
	}

	return hash.Sum64()
}

func runMicrobenchmarks(trials uint, speed Speed, monitorPower bool, powerInterval uint) {
	c := apd.BaseContext.WithPrecision(5)
	ed := apd.MakeErrDecimal(c)
	final := apd.New(1, 0) // Initial value for multiplied scores
	var curTrial uint

	stopChan := make(chan chan float64)
	powerResultChan := make(chan float64, 1)
	if monitorPower {
		go powerMonitor(powerInterval, stopChan)
	}

	beforeTrials := time.Now()
	for curTrial = 0; curTrial < trials; curTrial++ {
		fmt.Printf("Trial %d:\n", curTrial+1)

		var accumulated float64
		resultCache := make(map[uint64][]byte)
		for _, mb := range microbenchmarks {
			// Only run benchmark if speed is at desired speed or faster
			if mb.Speed < speed {
				continue
			}

			fmt.Printf("  %s: ", mb.Name)

			beforeBench := time.Now()
			score, rawValue, cached, err := mb.Run(resultCache)
			check(err)

			timeFmt := "%s"
			if cached {
				timeFmt = "[cached: %s]"
			}
			timeStr := fmt.Sprintf(timeFmt, formatDuration(time.Since(beforeBench)))
			fmt.Printf("%.2f %s, score: %.0f, time: %s\n", rawValue, mb.Unit, score, timeStr)
			accumulated += score
		}

		fmt.Printf("Score: %.0f\n\n", accumulated)
		score, _, err := c.NewFromString(strconv.FormatFloat(accumulated, 'f', -1, 64))
		check(err)
		ed.Mul(final, final, score)
		check(ed.Err())
	}

	/* Get the geometric mean of the power usage during benchmarks */
	var powerMean float64
	if monitorPower {
		stopChan <- powerResultChan
		powerMean = <-powerResultChan
		powerMean *= 1000 // W -> mW
	}

	/* Take the geometric mean of `final` */
	// nthRootPow := 1 / trials
	nthRootPow := apd.New(1, 0)
	ed.Quo(nthRootPow, nthRootPow, apd.New(int64(trials), 0))
	check(ed.Err())
	// Take the [trials]th root of the multiplied scores for the geometric mean
	finalScore := apd.New(1, 0)
	ed.Pow(finalScore, final, nthRootPow)
	check(ed.Err())
	// Convert the precise decimal into a float64 to display (we round it anyway)
	finalScoreFloat, err := finalScore.Float64()
	check(err)

	/* Output results */
	fmt.Printf("\nFinal score: %.0f\n", finalScoreFloat)
	if monitorPower {
		fmt.Printf("Power usage: %.0f mW\n", powerMean)
	}
	fmt.Printf("Time elapsed: %s\n", formatDuration(time.Since(beforeTrials)))
}
