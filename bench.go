package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
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

func runMicrobenchmarks(trials int, monitorPower bool) {
	var finalAvg float64
	var powerAvg float64
	before := time.Now()

	for trial := 0; trial < trials; trial++ {
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

		fmt.Printf("Trial %d score: %.0f; power usage: %.0f mW\n\n", trial+1, accumulated, trialPowerAvg)
		finalAvg += accumulated
		powerAvg += trialPowerAvg

		if trial < trials - 1 {
			time.Sleep(2 * time.Second)
		}
	}

	finalScore := finalAvg / float64(trials)
	powerAvg /= float64(trials)
	fmt.Printf("\nFinal score: %.0f\n", finalScore)
	fmt.Printf("Average power usage: %.0f mW\n", powerAvg)
	fmt.Println("Time elapsed:", time.Since(before))
}
