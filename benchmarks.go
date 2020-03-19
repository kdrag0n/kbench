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
		RefValue:       0,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time syscall: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Time vDSO call",
		RefValue:       0,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Time implicit: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"t", "100000", "25", "3"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "In-memory I/O",
		RefValue:       0,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`copied, ([\d.]+) s,`),
		Program:        "dd",
		Arguments:      []string{"if=/dev/zero", "of=/dev/null", "count=10000000"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "IPC messaging",
		RefValue:       0,
		Unit:           "s",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Total time: ([\d.]+) \[sec\]`),
		Program:        "perf",
		Arguments:      []string{"bench", "sched", "messaging", "-ptl", "8000"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Pipes",
		RefValue:       0,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`(\d+) ops/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "sched", "pipe"},
		Speed:          Slow,
	},
	Benchmark{
		Name:           "Futex hashing",
		RefValue:       0,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "hash"},
		Speed:          Slow,
	},
	Benchmark{
		Name:           "Serial futex wakeups",
		RefValue:       0,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Wokeup \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "wake", "-w", "8", "-t", "2048"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Parallel futex wakeups",
		RefValue:       0,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Avg per-thread latency \(waking 1/\d+ threads\) in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "wake-parallel", "-t", "2048"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "Futex requeuing",
		RefValue:       0,
		Unit:           "ms",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Requeued \d+ of \d+ threads in ([\d.]+) ms`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "requeue", "-t", "2048"},
		Speed:          Fast,
	},
	Benchmark{
		Name:           "PI futex locking",
		RefValue:       0,
		Unit:           "op/s",
		HigherIsBetter: true,
		Pattern:        regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:        "perf",
		Arguments:      []string{"bench", "futex", "lock-pi"},
		Speed:          Slow,
	},
	Benchmark{
		Name:           "VFS mmap",
		RefValue:       0,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`mmap: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "VFS I/O syscalls",
		RefValue:       0,
		Unit:           "ns",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`syscalls: (\d+) ns`),
		Program:        "callbench",
		Arguments:      []string{"f"},
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Scheduler wakeups",
		RefValue:       0,
		Unit:           "µs",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`99.0th: (\d+)`),
		Program:        "schbench",
		Speed:          Medium,
	},
	Benchmark{
		Name:           "Timer jitter",
		RefValue:       0,
		Unit:           "µs",
		HigherIsBetter: false,
		Pattern:        regexp.MustCompile(`Avg:\s+(\d+)`),
		Program:        "cyclictest",
		Arguments:      []string{"-qD", "5"},
		Speed:          Medium,
	},
}
