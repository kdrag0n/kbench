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
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
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
typedef long (*loop_impl)(ulong, ulong);

static long true_ns(struct timespec ts) {
    return ts.tv_nsec + (ts.tv_sec * NS_PER_SEC);
}

static long syscall_loop(ulong calls, ulong iters) {
    long best_ns = LONG_MAX;

    for (unsigned int i = 0; i < iters; i++) {
        struct timespec before;
        clock_gettime(CLOCK_MONOTONIC_RAW, &before);

        struct timespec holder;
        for (unsigned int call = 0; call < calls; call++) {
            syscall(__NR_clock_gettime, CLOCK_MONOTONIC, &holder);
        }

        struct timespec after;
        clock_gettime(CLOCK_MONOTONIC_RAW, &after);

        long elapsed_ns = true_ns(after) - true_ns(before);
        if (elapsed_ns < best_ns) {
            best_ns = elapsed_ns;
        }
    }

    best_ns /= calls; // per syscall in the loop
    return best_ns;
}

static long implicit_loop(ulong calls, ulong iters) {
    long best_ns = LONG_MAX;

    for (unsigned int i = 0; i < iters; i++) {
        struct timespec before;
        clock_gettime(CLOCK_MONOTONIC_RAW, &before);

        struct timespec holder;
        for (unsigned int call = 0; call < calls; call++) {
            clock_gettime(CLOCK_MONOTONIC, &holder);
        }

        struct timespec after;
        clock_gettime(CLOCK_MONOTONIC_RAW, &after);

        long elapsed_ns = true_ns(after) - true_ns(before);
        if (elapsed_ns < best_ns) {
            best_ns = elapsed_ns;
        }
    }

    best_ns /= calls; // per call in the loop
    return best_ns;
}

static ulong get_arg(int argc, char** argv, int index, ulong default_value) {
    ulong value = 0;

    if (argc > index)
        value = atoi(argv[index]);
    if (value == 0)
        value = default_value;

    return value;
}

static long test_loop_best_ns(loop_impl call_loop, ulong calls, ulong iters, ulong reps) {
    long best_ns = LONG_MAX;
    for (unsigned int rep = 0; rep < reps; rep++) {
        long elapsed_ns = call_loop(calls, iters);
        if (elapsed_ns < best_ns) {
            best_ns = elapsed_ns;
        }

        putchar('.');
        fflush(stdout);
        usleep(US_PER_SEC / 2); // 500 ms
    }

    return best_ns;
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

    printf("Sysbench - syscall tester by kdrag0n\n"
           "%lu calls for %lu iterations with %lu repetitions\n"
           "The implicit call can be backed by a true syscall (in lieu of faster routines), vDSO (Linux), or commpage (macOS).\n"
           "\n"
           "\n", calls, iters, reps);
    
    long best_ns_syscall = test_loop_best_ns(syscall_loop, calls, iters, reps);
    long best_ns_implicit = test_loop_best_ns(implicit_loop, calls, iters, reps);

    putchar('\n');

    printf("Syscall: %ld ns\n", best_ns_syscall);
    printf("Implicit: %ld ns\n", best_ns_implicit);

    return 0;
}
