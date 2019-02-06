/*
 * sysbench.c
 * 
 * This program benchmarks the clock_gettime kernel syscall on Unix systems by
 * reading the CLOCK_MONOTONIC_RAW value. This is usually the fastest value to
 * read so we can ensure minimal in-kernel execution, leaving just the time
 * taken by the actual syscall.
 *
 * Syscalls are typically relatively slow operations, requiring a full context
 * switch, register saving, and more. If a system is using a vDSO, this should
 * be made much faster. Usually, time-related calls are part of the vDSO because
 * they do not require special privileges and can be safely executed in
 * userspace. This program should be an effective method of testing that.
 * This program is also able to benchmark the true syscall, as well as the
 * implicit call (provided by the libc implementation) which can be provided by
 * the vDSO, if any.
 *
 * Licensed under the MIT License (MIT)
 *
 * Copyright (c) 2019 Khronodragon "kdrag0n" <kdrag0n@pm.me>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#include <sys/syscall.h>
#include <sys/time.h>
#include <sys/mman.h>
#include <time.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <limits.h>
#include <unistd.h>
#include <fcntl.h>
#include <ctype.h>

#define TEST_FILE_PATH "/proc/sys/kernel/ostype"
#define NS_PER_SEC 1000000000
#define US_PER_SEC 1000000
#define true 1
#define false 0

typedef _Bool bool;
typedef unsigned long ulong;
typedef long (*bench_impl)(void);

static inline long true_ns(struct timespec ts) {
    return ts.tv_nsec + (ts.tv_sec * NS_PER_SEC);
}

static long time_syscall_mb(void) {
    struct timespec ts;
    syscall(__NR_clock_gettime, CLOCK_MONOTONIC, &ts);
}

static long time_implicit_mb(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
}

static long mmap_mb(void) {
    int fd = open(TEST_FILE_PATH, O_RDONLY); // open file: read-only
    int len = lseek(fd, 0, SEEK_END); // seek to end to get length

    void *data = mmap(NULL, len, PROT_READ, MAP_PRIVATE | MAP_POPULATE, fd, 0); // mmap it: read-only, preload contents at kernel's discretion
	void *copy = malloc(len);
	memcpy(copy, data, len); // copy data (effectively reading it)
	munmap(data, len); // unmap it

	close(fd);
}

static long file_mb(void) {
	FILE *f = fopen(TEST_FILE_PATH, "rb"); // open file: read + binary handling

	fseek(f, 0, SEEK_END); // seek to end for length
	long len = ftell(f); // get current position at end -> length
	fseek(f, 0, SEEK_SET); // seek back to beginning to read

	void *buf = malloc(len);
	fread(buf, 1, len, f); // read 1 * len bytes from f into buf

	fclose(f);
}

static ulong get_arg(int argc, char** argv, int index, ulong default_value) {
    ulong value = 0;

    if (argc > index)
        value = atoi(argv[index]);
    if (value == 0)
        value = default_value;

    return value;
}

static long run_bench_ns(bench_impl inner_call, ulong calls, ulong iters, ulong reps) {
    long best_ns1 = LONG_MAX;
    for (unsigned int rep = 0; rep < reps; rep++) {
        long best_ns2 = LONG_MAX;

        for (unsigned int i = 0; i < iters; i++) {
            struct timespec before;
            clock_gettime(CLOCK_MONOTONIC_RAW, &before);

            for (unsigned int call = 0; call < calls; call++) {
                inner_call();
            }

            struct timespec after;
            clock_gettime(CLOCK_MONOTONIC_RAW, &after);

            long elapsed_ns = true_ns(after) - true_ns(before);
            if (elapsed_ns < best_ns2) {
                best_ns2 = elapsed_ns;
            }
        }
        best_ns2 /= calls; // per call in the loop

        if (best_ns2 < best_ns1) {
            best_ns1 = best_ns2;
        }

        putchar('.');
        fflush(stdout);
        usleep(US_PER_SEC / 8); // 125 ms
    }

    return best_ns1;
}

int bench_time(int argc, char** argv) {
    ulong calls = get_arg(argc, argv, 2, 100000);
    ulong iters = get_arg(argc, argv, 3, 32);
    ulong reps = get_arg(argc, argv, 4, 5);

    printf("Time benchmark:\n"
           "%lu calls for %lu iterations with %lu repetitions\n"
           "The implicit call may be backed by vDSO.\n"
           "\n"
           "\n", calls, iters, reps);
    
    long best_ns_syscall = run_bench_ns(time_syscall_mb, calls, iters, reps);
    long best_ns_implicit = run_bench_ns(time_implicit_mb, calls, iters, reps);

    putchar('\n');

    printf("Time syscall: %ld ns\n", best_ns_syscall);
    printf("Time implicit: %ld ns\n", best_ns_implicit);

    return 0;
}

int bench_file(int argc, char** argv) {
    ulong calls = get_arg(argc, argv, 2, 100);
    ulong iters = get_arg(argc, argv, 3, 128);
    ulong reps = get_arg(argc, argv, 4, 5);

    printf("File benchmark:\n"
           "%lu calls for %lu iterations with %lu repetitions\n"
           "\n"
           "\n", calls, iters, reps);

    long best_ns_mmap = run_bench_ns(mmap_mb, calls, iters, reps);
    long best_ns_file = run_bench_ns(file_mb, calls, iters, reps);

    putchar('\n');

    printf("File via mmap: %ld ns\n", best_ns_mmap);
    printf("File via fd I/O: %ld ns\n", best_ns_file);

    return 0;
}

int main(int argc, char** argv) {
	int ret;
	bool do_time = false;
	bool do_file = false;

	printf("Sysbench: syscall benchmark by kdrag0n\n\n");
    if (argc == 1) { // No arguments supplied
        printf("Optional usage: %s [mode: [t]ime, [f]ile, [a]ll] [# of calls] [# of iterations] [# of repetitions]\n"
               "\n", argv[0]);
    }

	char mode = 'a';
	if (argc >= 2) { // 1+ arguments
		mode = tolower(argv[1][0]); // First letter of 1st argument
		if (mode != 't' && mode != 'f' && mode != 'a') {
			fprintf(stderr, "Invalid mode '%c'! Valid modes are: [t]ime, [f]ile, [a]ll\n", mode);
			return 1;
		}
	}

	switch (mode) {
	case 't':
		do_time = true;
		break;
	case 'f':
		do_file = true;
		break;
	case 'a':
		do_time = true;
		do_file = true;
		break;
	}

	if (do_time) {
		ret = bench_time(argc, argv);
		if (ret)
			return ret;
	}

	if (do_file) {
		ret = bench_file(argc, argv);
		if (ret)
			return ret;
	}

	return 0;
}