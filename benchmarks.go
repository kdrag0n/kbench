package main

import (
	"regexp"
)

// RefScore is the score to normalize each reference to
const RefScore = 500

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
	Benchmark{
		Name:           "Time syscall",
		RefValue:       229,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time syscall: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Time vDSO call",
		RefValue:       35,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time implicit: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Null block I/O",
		RefValue:       5.9,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`copied, ([\d.]+) s,`),
		Program:        "dd",
		Arguments:      []string{"if=/dev/zero", "of=/dev/null", "count=10000000"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "IPC messaging",
		RefValue:       18.7,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Total time: ([\d.]+) \[sec\]`),
		Program:        "perf",
		Arguments:      []string{"bench", "sched", "messaging", "-ptl", "8000"},
		Speed:          Slow,
	},
	Benchmark{
		Name:           "Pipe IPC",
		RefValue:       87784,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`(\d+) ops/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "sched", "pipe"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Futex hashing",
		RefValue:       1814540,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "hash"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Futex wakeup",
		RefValue:       124.6,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Wokeup \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "wake", "-w", "8", "-t", "2048"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Futex requeuing",
		RefValue:       249.4,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Requeued \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "requeue", "-t", "2048"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "PI futex locking",
		RefValue:       208,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "lock-pi"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "VFS mmap",
		RefValue:       8486,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`mmap: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "VFS I/O syscalls",
		RefValue:       9548,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`syscalls: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Scheduler wakeup",
		RefValue:       36928,
		Unit:           "µs",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`99.0th: (\d+)`),
		Program:        "schbench",
		Speed:          Slow,
	},
}
