package main

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func readFileLine(path string) (cleaned string, err error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	cleaned = strings.TrimSuffix(string(raw), "\n")
	return
}

func getTickRate() float64 {
	uptimeStr, err := readFileLine("/proc/uptime")
	check(err)
	uptimeSec, err := strconv.ParseFloat(strings.Fields(uptimeStr)[0], 64)
	check(err)

	statStr, err := readFileLine("/proc/self/stat")
	check(err)
	statTimeoutJiffies, err := strconv.Atoi(strings.Fields(statStr)[21])
	check(err)

	return uptimeSec / float64(statTimeoutJiffies)
}
