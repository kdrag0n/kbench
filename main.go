package main

import (
	"io/ioutil"
	"flag"
	"fmt"
	"os"
	"os/user"
	"os/exec"
)

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
	var stopAndroid bool
	flag.BoolVar(&stopAndroid, "stop-android", true, "Whether to stop most of the Android system to prevent interference and reduce variables. Android will be restarted automatically when the benchmarks finish.")
	flag.Parse()

	if trials == 0 {
		fmt.Fprintf(os.Stderr, "Trial count must be non-zero!\n")
		os.Exit(1)
	}

	user, err := user.Current()
	check(err)
	if user.Uid != "0" {
		fmt.Fprintf(os.Stderr, "Must be run as root!\n")
		os.Exit(1)
	}

	fmt.Println("Preparing environment...")
	setupPerfEnv()
	defer cleanupPerfEnv()

	fmt.Print("Running warmup round...")
	runWarmup()
	fmt.Print("\n")

	_, err = os.Stat(GChargeStopLevel)
	if monitorPower {
		if os.IsNotExist(err) {
			before, err := ioutil.ReadFile(GChargeStopLevel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to back up charge limit: %v\n", err)
			}

			err = ioutil.WriteFile(GChargeStopLevel, []byte("2"), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to disable charging and power source: %v\n", err)
			}

			defer func() {
				err = ioutil.WriteFile(GChargeStopLevel, before, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to restore backed up charge limit: %v\n", err)
				}
			}()
		} else {
			fmt.Fprintf(os.Stderr, "Cannot disable charging: %v; power usage may not be accurate\n", err)
		}
	}

	if stopAndroid {
		fmt.Println("Stopping Android...")
		exec.Command("/system/bin/stop").Run()
		defer func() {
			fmt.Println("Restarting Android...")
			exec.Command("/system/bin/start").Run()
		}()
	}

	fmt.Print("Running benchmark...\n\n")
	runMicrobenchmarks(trials, monitorPower)

	return 0
}

func main() {
	os.Exit(cmdMain())
}
