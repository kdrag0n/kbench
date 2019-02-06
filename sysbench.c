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
#include <time.h>
#include <stdio.h>
#include <stdlib.h>
#include <limits.h>
#include <unistd.h>

#define NS_PER_SEC 1000000000
#define US_PER_SEC 1000000

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
        usleep(US_PER_SEC / 4); // 250 ms
    }

    return best_ns1;
}

int main(int argc, char** argv) {
    ulong calls = get_arg(argc, argv, 1, 100000);
    ulong iters = get_arg(argc, argv, 2, 32);
    ulong reps = get_arg(argc, argv, 3, 5);

    if (argc == 1) {
        printf("Usage: %s [# of calls] [# of iterations] [# of repetitions]\n"
               "All arguments are optional.\n"
               "\n", argv[0]);
    }

    printf("Sysbench: syscall benchmark by kdrag0n\n"
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
