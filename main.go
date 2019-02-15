package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"syscall"
	"time"
)

// GChargeStopLevel is the path to the kernel battery charge limit pseudo-file.
const GChargeStopLevel = "/sys/devices/platform/soc/soc:google,charger/charge_stop_level"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func cmdMain() int {
	fmt.Println("KBench by @kdrag0n")

	var trials uint
	flag.UintVar(&trials, "trials", 3, "The number of times to run all the microbenchmarks. The geometric mean of each trial's score is calculated for the final score.")
	var monitorPower bool
	flag.BoolVar(&monitorPower, "power", true, "Whether to monitor system power usage during the test. Only works accurately on Google Pixel devices.")
	var powerInterval uint
	flag.UintVar(&powerInterval, "power-interval", 250, "The interval in milliseconds at which to sample power usage during benchmarks.")
	var stopAndroid bool
	flag.BoolVar(&stopAndroid, "stop-android", true, "Whether to stop most of the Android system to prevent interference and reduce variables. Android will be restarted automatically when the benchmarks finish.")
	var rawSpeed uint
	flag.UintVar(&rawSpeed, "speed", 0, "The speed at which to run at, skipping slower benchmarks if necessary. Benchmark results from one speed level are not comparable to others. Available speed classes: 0 = slow (all), 1 = medium (most), and 2 = fast (some).")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [options]
Option format: -[name]=[value]
Example usage: kbench -trials=5 -power=false -stop-android=false

Supported options:
`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if trials == 0 {
		fmt.Fprintf(os.Stderr, "Trial count must be non-zero.\n")
		return 1
	}

	if rawSpeed >= uint(MaxSpeed) {
		fmt.Fprintf(os.Stderr, "Invalid speed %d. Available classes: 0 (slow), 1 (medium), 2 (fast)\n", rawSpeed)
		return 1
	}
	speed := Speed(rawSpeed)

	user, err := user.Current()
	check(err)
	if user.Uid != "0" {
		fmt.Fprintf(os.Stderr, "This program must be run as root.\n")
		return 1
	}

	deferredFuncs := make([]func(), 0, 2)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		for _, fun := range deferredFuncs {
			fun()
		}
	}()

	if monitorPower {
		voltNow := PsyPrefix + "voltage_now"
		_, err = os.Stat(voltNow)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to stat voltage_now: %v; disabling power monitor\n", err)
			monitorPower = false
			goto skipChargeLimit
		}

		_, err = os.Stat(GChargeStopLevel)
		if !os.IsNotExist(err) {
			before, err := ioutil.ReadFile(GChargeStopLevel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to back up charge limit: %v\n", err)
			}

			err = ioutil.WriteFile(GChargeStopLevel, []byte("2"), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to disable charging and power source: %v\n", err)
			}

			cLimitRestoreFunc := func() {
				err = ioutil.WriteFile(GChargeStopLevel, before, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to restore backed up charge limit: %v\n", err)
				}
			}
			defer cLimitRestoreFunc()
			deferredFuncs = append(deferredFuncs, cLimitRestoreFunc)
		} else {
			fmt.Fprintf(os.Stderr, "Cannot disable charging: %v; power usage may not be accurate\n", err)
		}
	}

skipChargeLimit:
	if stopAndroid {
		fmt.Println("Stopping Android...")
		exec.Command("/system/bin/stop").Run()

		startAndroidFunc := func() {
			fmt.Println("Restarting Android...")
			exec.Command("/system/bin/start").Run()
		}
		defer startAndroidFunc()
		deferredFuncs = append(deferredFuncs, startAndroidFunc)

		fmt.Println("Waiting for processes to stop...")
		time.Sleep(2 * time.Second)
	}

	os.Stderr.Sync()
	fmt.Print("\n")
	runMicrobenchmarks(trials, speed, monitorPower, powerInterval)

	return 0
}

func main() {
	os.Exit(cmdMain())
}
