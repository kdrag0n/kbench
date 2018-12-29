package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("KBench by @kdrag0n")

	var trials int
	flag.IntVar(&trials, "trials", 3, "Number of trials to run for each microbenchmark")
	flag.Parse()

	user, err := user.Current()
	check(err)
	if user.Uid != "0" {
		fmt.Fprintf(os.Stderr, "Must be run as root!")
		os.Exit(1)
	}

	fmt.Println("Preparing environment...")
	setupPerfEnv()
	defer cleanupPerfEnv()

	fmt.Print("Running warmup round...")
	runWarmup()
	fmt.Print("\n")

	fmt.Print("Running benchmark...\n\n")
	runMicrobenchmarks(trials)
}
