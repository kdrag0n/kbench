package main

import (
	"io/ioutil"
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
