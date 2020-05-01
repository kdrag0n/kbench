package main

import (
	"regexp"
)

// RefScore is the score to normalize each reference to
const RefScore = 1000

// A Speed represents a class of benchmark speeds.
type Speed uint16

// Speeds at which benchmarks are classified as
const (
	Slow Speed = iota
	Medium
	Fast
	MaxSpeed
)

// Benchmark describes the details of a single benchmark.
type Benchmark struct {
	Name           string
	HigherIsBetter bool
	RefValue       float64
	Unit           string
	Pattern        *regexp.Regexp
	Program        string
	Arguments      []string
	Speed          Speed
}

var benchmarks = []Benchmark{
	{
		Name:           "Time syscall",
		RefValue:       229,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time syscall: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Fast,
	},
	{
		Name:           "Time vDSO call",
		RefValue:       35,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time implicit: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Fast,
	},
	{
		Name:           "Null block I/O",
		RefValue:       5.9,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`([\d.]+) s`),
		Program:        "dd",
		Arguments:      []string{"if=/dev/zero", "of=/dev/null", "count=10000000"},
		Speed:          Fast,
	},
	{
		Name:           "IPC messaging",
		RefValue:       18.7,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Total time: ([\d.]+) \[sec\]`),
		Program:        "perf",
		Arguments:      []string{"bench", "sched", "messaging", "-ptl", "8000"},
		Speed:          Slow,
	},
	{
		Name:           "Futex hashing",
		RefValue:       1814540,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "hash"},
		Speed:          Medium,
	},
	{
		Name:           "Futex wakeup",
		RefValue:       124.6,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Wokeup \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "wake", "-w", "8", "-t", "2048"},
		Speed:          Medium,
	},
	{
		Name:           "Futex requeuing",
		RefValue:       249.4,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Requeued \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "requeue", "-t", "2048"},
		Speed:          Medium,
	},
	{
		Name:           "VFS mmap",
		RefValue:       8486,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`mmap: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Medium,
	},
	{
		Name:           "VFS I/O syscalls",
		RefValue:       9548,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`syscalls: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Fast,
	},
	{
		Name:           "Scheduler wakeup",
		RefValue:       36928,
		Unit:           "µs",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`99.0th: (\d+)`),
		Program:        "schbench",
		Speed:          Slow,
	},
}
