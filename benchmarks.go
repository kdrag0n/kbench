package main

import (
	"regexp"
)

// A Speed represents a class of benchmark speeds.
type Speed uint16

// Speeds at which benchmarks are classified as
const (
	Slow Speed = iota
	Medium
	Fast
	MaxSpeed
)

// Microbenchmark describes the details of a single microbenchmark.
type Microbenchmark struct {
	Name         string
	MoreIsBetter bool
	Factor       float64
	Unit         string
	Pattern      *regexp.Regexp
	Program      string
	Arguments    []string
	Speed        Speed
}

var microbenchmarks = []Microbenchmark{
	Microbenchmark{
		Name:         "Basic syscall",
		MoreIsBetter: false,
		Factor:       2,
		Unit:         "ns",
		Pattern:      regexp.MustCompile(`Syscall: (\d+) ns`),
		Program:      "sysbench",
		Arguments:    []string{"100000", "25", "3"},
		Speed:        Medium,
	},
	Microbenchmark{
		Name:         "Basic vDSO call",
		MoreIsBetter: false,
		Factor:       2,
		Unit:         "ns",
		Pattern:      regexp.MustCompile(`Implicit: (\d+) ns`),
		Program:      "sysbench",
		Arguments:    []string{"100000", "25", "3"},
		Speed:        Medium,
	},
	Microbenchmark{
		Name:         "In-memory I/O",
		MoreIsBetter: false,
		Factor:       1,
		Unit:         "ms",
		Pattern:      regexp.MustCompile(`([\d.]+)\s+(?:msec\s+)?task-clock`),
		Program:      "perf",
		Arguments:    []string{"stat", "-B", "dd", "if=/dev/zero", "of=/dev/null", "count=1000000"},
		Speed:        Fast,
	},
	Microbenchmark{
		Name:         "IPC messaging",
		MoreIsBetter: false,
		Factor:       100,
		Unit:         "sec",
		Pattern:      regexp.MustCompile(`Total time: ([\d.]+) \[sec\]`),
		Program:      "perf",
		Arguments:    []string{"bench", "sched", "messaging"},
		Speed:        Fast,
	},
	Microbenchmark{
		Name:         "Pipe IPC",
		MoreIsBetter: true,
		Factor:       1 / 100.0,
		Unit:         "ops/sec",
		Pattern:      regexp.MustCompile(`(\d+) ops/sec`),
		Program:      "perf",
		Arguments:    []string{"bench", "sched", "pipe"},
		Speed:        Slow,
	},
	Microbenchmark{
		Name:         "Futex hashing",
		MoreIsBetter: true,
		Factor:       1 / 10000.0,
		Unit:         "ops/sec",
		Pattern:      regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:      "perf",
		Arguments:    []string{"bench", "futex", "hash"},
		Speed:        Slow,
	},
	Microbenchmark{
		Name:         "Serial futex wakeups",
		MoreIsBetter: false,
		Factor:       100,
		Unit:         "ms",
		Pattern:      regexp.MustCompile(`Wokeup 32 of 32 threads in ([\d.]+) ms`),
		Program:      "perf",
		Arguments:    []string{"bench", "futex", "wake", "-w", "8", "-t", "32"},
		Speed:        Fast,
	},
	Microbenchmark{
		Name:         "Parallel futex wakeups",
		MoreIsBetter: false,
		Factor:       10000,
		Unit:         "ms",
		Pattern:      regexp.MustCompile(`Avg per-thread latency \(waking 1/64 threads\) in ([\d.]+) ms`),
		Program:      "perf",
		Arguments:    []string{"bench", "futex", "wake-parallel", "-t", "64"},
		Speed:        Fast,
	},
	Microbenchmark{
		Name:         "Futex requeuing",
		MoreIsBetter: false,
		Factor:       1000,
		Unit:         "ms",
		Pattern:      regexp.MustCompile(`Requeued 32 of 32 threads in ([\d.]+) ms`),
		Program:      "perf",
		Arguments:    []string{"bench", "futex", "requeue", "-t", "32"},
		Speed:        Fast,
	},
	Microbenchmark{
		Name:         "PI futex locking",
		MoreIsBetter: true,
		Factor:       1 / 4.0,
		Unit:         "ops/sec",
		Pattern:      regexp.MustCompile(`Averaged (\d+) operations/sec`),
		Program:      "perf",
		Arguments:    []string{"bench", "futex", "lock-pi"},
		Speed:        Slow,
	},
}
