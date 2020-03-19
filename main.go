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

// SysPwrSuspend is the path to the kernel sysfs node that disables charging.
const SysPwrSuspend = "/sys/class/power_supply/battery/input_suspend"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func cmdMain() int {
	var trials int
	flag.IntVar(&trials, "trials", 3, "The number of times to run all the benchmarks. The geometric mean of each trial's score is calculated for the final score.")
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

	if _, err = os.Stat(PsyPrefix + "voltage_now"); monitorPower && err != nil {
		_, err = os.Stat(SysPwrSuspend)
		if !os.IsNotExist(err) {
			err = ioutil.WriteFile(SysPwrSuspend, []byte("1"), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to disable charging: %v\n", err)
			}

			chgEnableFunc := func() {
				err = ioutil.WriteFile(SysPwrSuspend, []byte("0"), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to enable charging: %v\n", err)
				}
			}
			defer chgEnableFunc()
			deferredFuncs = append(deferredFuncs, chgEnableFunc)
		} else {
			fmt.Fprintf(os.Stderr, "Cannot disable charging: %v; power usage will be inaccurate\n", err)
		}
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to stat voltage_now: %v; disabling power monitor\n", err)
	}

	if stopAndroid {
		fmt.Println("Stopping Android...")
		exec.Command("/system/bin/start blank_screen").Run()
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
	runBenchmarks(trials, speed, monitorPower, powerInterval)

	return 0
}

func main() {
	os.Exit(cmdMain())
}
