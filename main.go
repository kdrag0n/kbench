package main

import (
	"os"
	"fmt"
	"os/user"
	"flag"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var trials int
	flag.IntVar(&trials, "trials", 3, "Number of trials to run for each microbenchmark")
	flag.Parse()

	user, err := user.Current()
	check(err)
	if user.Uid != "0" {
		fmt.Fprintf(os.Stderr, "Must be run as root!")
		os.Exit(1)
	}

	fmt.Printf("Preparing environment...\n")
	setupPerfEnv()
	defer cleanupPerfEnv()

	fmt.Printf("Running benchmark...\n\n")
	runMicrobenchmarks(trials)
}