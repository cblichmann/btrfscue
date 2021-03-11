/*
 * btrfscue version 0.6
 * Copyright (c)2011-2021 Christian Blichmann
 *
 * Command-line application utilities
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package util

import (
	"fmt"
	"os"
)

const warnPrefix = "btrfscue: "

var verbose = false

// Warnf prints a formatted warning message to stderr.
func Warnf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, warnPrefix+format, v...)
}

// Fatalf prints a formatted error message to stderr and exits the program
// with exit code 1.
func Fatalf(format string, v ...interface{}) {
	Warnf(format, v...)
	os.Exit(1)
}

// SetVerbose enables or disables verbose messages.
func SetVerbose(v bool) { verbose = v }

// Verbosef prints a formatted message to stdout if in verbose mode. Use
// SetVerbose to enable/disable verbose mode.
func Verbosef(format string, v ...interface{}) {
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
