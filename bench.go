package main

import (
	"runtime"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
	"github.com/cockroachdb/apd"
)

// Run the microbenchmark and calculate a normalized score from the results
func (mb *Microbenchmark) Run() (score float64, rawValue float64, err error) {
	var out []byte

	// Use bundled program if possible, otherwise resort to system PATH
	progPath := runtime.GOARCH + "/" + mb.Program
	if _, err = os.Stat(progPath); err != nil {
		if !os.IsNotExist(err) { // Not existing is normal, anything else is a warning
			fmt.Fprintf(os.Stderr, "Unable to stat '%s': %v; resorting to system PATH\n", progPath, err)
		}

		progPath = mb.Program
	}
	out, err = exec.Command(progPath, mb.Arguments...).CombinedOutput()

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

	return
}

func runMicrobenchmarks(trials uint, monitorPower bool, powerInterval uint) {
	c := apd.BaseContext.WithPrecision(5)
	ed := apd.MakeErrDecimal(c)
	final := apd.New(1, 0) // Initial value for multiplied scores
	var powerAvg float64
	var curTrial uint
	before := time.Now()

	for curTrial = 0; curTrial < trials; curTrial++ {
		var accumulated float64
		stopChan := make(chan chan float64)
		powerResultChan := make(chan float64, 1)
		if monitorPower {
			go powerMonitor(powerInterval, stopChan)
		}

		for _, mb := range microbenchmarks {
			fmt.Printf("%s: ", mb.Name)

			score, rawValue, err := mb.Run()
			check(err)

			var better string
			if mb.MoreIsBetter {
				better = "more"
			} else {
				better = "less"
			}
			fmt.Printf("%.2f %s (%s is better), score: %.0f\n", rawValue, mb.Unit, better, score)
			accumulated += score
		}

		trialPowerSuf := ""
		if monitorPower {
			stopChan <- powerResultChan
			trialPowerAvg := <- powerResultChan
			powerAvg += trialPowerAvg
			trialPowerSuf = fmt.Sprintf("; power usage: %.0f mW", trialPowerAvg)
		}

		fmt.Printf("Trial %d score: %.0f%s\n\n", curTrial+1, accumulated, trialPowerSuf)
		score, _, err := c.NewFromString(strconv.FormatFloat(accumulated, 'f', -1, 64))
		check(err)
		ed.Mul(final, final, score)
		check(ed.Err())

		if curTrial < trials - 1 {
			time.Sleep(1 * time.Second)
		}
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

	/* Take the arithmetic mean of `powerAvg` */
	powerAvg /= float64(trials)

	/* Output results */
	fmt.Printf("\nFinal score: %.0f\n", finalScoreFloat)
	if monitorPower {
		fmt.Printf("Average power usage: %.0f mW\n", powerAvg)
	}
	fmt.Println("Time elapsed:", time.Since(before))
}
