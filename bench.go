package main

import (
	"time"
	"strconv"
	"fmt"
	"os/exec"
)

func runMicrobenchmarks(trials int) {
	var finalAvg float64

	for trial := 0; trial < trials; trial++ {
		var accumulated float64

		for _, mb := range microbenchmarks {
			var out []byte
			var err error
			fmt.Printf("%s: ", mb.Name)

			switch (mb.Program) {
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

			var better string
			if mb.MoreIsBetter { better = "more" } else { better = "less" }
			fmt.Printf("%.2f %s (%s is better), score: %.0f\n", value, mb.Unit, better, score)
			accumulated += score
		}

		fmt.Printf("Trial %d score: %.0f\n\n", trial + 1, accumulated)
		finalAvg += accumulated
		time.Sleep(2 * time.Second)
	}

	finalScore := finalAvg / float64(trials)
	fmt.Printf("\nFinal score: %.0f\n", finalScore)
}
