// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Command-line application utilities

package util

import (
	"fmt"
	"os"
)

const warnPrefix = "btrfscue: "

var verbose = false

// Warnf prints a formatted warning message to stderr.
func Warnf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, warnPrefix+format, v...)
}

// Fatalf prints a formatted error message to stderr and exits the program
// with exit code 1.
func Fatalf(format string, v ...any) {
	Warnf(format, v...)
	os.Exit(1)
}

// SetVerbose enables or disables verbose messages.
func SetVerbose(v bool) { verbose = v }

// Verbosef prints a formatted message to stdout if in verbose mode. Use
// SetVerbose to enable/disable verbose mode.
func Verbosef(format string, v ...any) {
	if verbose {
		fmt.Printf(format, v...)
	}
}

// ReportError checks if there was an error and conditionally reports it by
// calling Fatalf().
func ReportError(err error) {
	if err != nil {
		Fatalf("%s\n", err)
	}
}
