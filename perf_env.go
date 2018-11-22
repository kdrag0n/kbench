package main

import (
	"os"
	"os/exec"
)

var perfEnvPath string
var perfEnvDevs = []string{"zero", "null"}

func setupPerfEnv() {
	dir, err := os.Getwd()
	check(err)
	perfEnvPath = dir + "/perfenv"

	devPath := perfEnvPath + "/dev/"
	os.Mkdir(devPath, 755)

	for _, device := range perfEnvDevs {
		err = exec.Command("/sbin/.core/busybox/mount", "--bind", "/dev/" + device, devPath + device).Run()
		check(err)
	}
}

func cleanupPerfEnv() {
	for _, device := range perfEnvDevs {
		err := exec.Command("/sbin/.core/busybox/umount", perfEnvPath + "/dev/" + device).Run()
		check(err)
	}
}

func execPerf(arguments ...string) (out []byte, err error) {
	oldPath, oldExist := os.LookupEnv("LD_LIBRARY_PATH")
	os.Setenv("LD_LIBRARY_PATH", "/lib")

	arguments = append([]string{perfEnvPath, "/perf"}, arguments...)
	out, err = exec.Command("chroot", arguments...).CombinedOutput()
	
	if oldExist {
		os.Setenv("LD_LIBRARY_PATH", oldPath)
	} else {
		os.Unsetenv("LD_LIBRARY_PATH")
	}

	return
}
