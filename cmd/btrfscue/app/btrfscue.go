// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Global options and top-level app utility package

package app

type Options struct {
	Verbose   bool
	Progress  bool
	Machine   bool // Display machine parseable output
	BlockSize uint
	Metadata  string
}

var Global Options
