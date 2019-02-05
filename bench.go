package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
	"github.com/shopspring/decimal"
)

func runWarmup() {
	for _, mb := range microbenchmarks {
		switch mb.Program {
		case ProgramSysbench:
			exec.Command("./sysbench", mb.Arguments...).Output()
		case ProgramPerf:
			execPerf(mb.Arguments...)
		}

		fmt.Print(".")
	}
}

func runMicrobenchmarks(trials uint, monitorPower bool) {
	final := decimal.NewFromFloat(1) // Initial value for multiplied scores
	var powerAvg float64
	var curTrial uint
	before := time.Now()

	for curTrial = 0; curTrial < trials; curTrial++ {
		var accumulated float64
		stopChan := make(chan chan float64)
		powerResultChan := make(chan float64, 1)
		go powerMonitor(stopChan)

		for _, mb := range microbenchmarks {
			var out []byte
			var err error
			fmt.Printf("%s: ", mb.Name)

			switch mb.Program {
			case ProgramSysbench:
				out, err = exec.Command("./sysbench", mb.Arguments...).Output()
			case ProgramPerf:
				out, err = execPerf(mb.Arguments...)
			default:
				panic(fmt.Sprintf("Unsupported program %d specified for microbenchmark %s", mb.Program, mb.Name))
			}
			check(err)

			matches := mb.Pattern.FindSubmatch(out)
			if len(matches) < 2 {
				fmt.Println(string(out))
				panic(fmt.Sprintf("Output microbenchmark %s did not match expected format", mb.Name))
			}

			value, err := strconv.ParseFloat(string(matches[1]), 64)
			check(err)

			score := value * mb.Factor
			if !mb.MoreIsBetter {
				score = 1000 - score
			}
			if score < 0 {
				score = 0
			}

			var better string
			if mb.MoreIsBetter {
				better = "more"
			} else {
				better = "less"
			}
			fmt.Printf("%.2f %s (%s is better), score: %.0f\n", value, mb.Unit, better, score)
			accumulated += score
		}

		stopChan <- powerResultChan
		trialPowerAvg := <- powerResultChan

		fmt.Printf("Trial %d score: %.0f; power usage: %.0f mW\n\n", curTrial+1, accumulated, trialPowerAvg)
		final = final.Mul(decimal.NewFromFloat(accumulated))
		powerAvg += trialPowerAvg

		if curTrial < trials - 1 {
			time.Sleep(2 * time.Second)
		}
	}

	/* Take the geometric mean of `final` */
	// Power: 1/n == nth root - compute this value
	nthRootPow := decimal.NewFromFloat(1).Div(decimal.NewFromFloat(float64(trials)))
	// Take the [trials]th root of the multiplied scores for the geometric mean
	finalScore := final.Pow(nthRootPow)
	// Convert the precise decimal into a float64 to display (we round it anyway)
	finalScoreFloat, _ := finalScore.Float64()

	/* Take the arithmetic mean of `powerAvg` */
	powerAvg /= float64(trials)

	/* Output results */
	fmt.Printf("\nFinal score: %.0f\n", finalScoreFloat)
	fmt.Printf("Average power usage: %.0f mW\n", powerAvg)
	fmt.Println("Time elapsed:", time.Since(before))
}
