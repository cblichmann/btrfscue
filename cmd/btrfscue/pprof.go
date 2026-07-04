// +build pprof

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Enables CPU profiling if build with -tags 'pprof'

package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
)

func startProfiling() {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		reportError(err)
		reportError(pprof.StartCPUProfile(f))
	}
	// Enable local profiler at http://localhost:6060/debug/pprof/. See
	// https://blog.golang.org/profiling-go-programs on how to use it.
	go func() { reportError(http.ListenAndServe(":6060", nil)) }()
}

func stopProfiling() {
	pprof.StopCPUProfile()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		defer f.Close()
		reportError(err)
		runtime.GC() // Update statistics
		reportError(pprof.WriteHeapProfile(f))
	}
}
