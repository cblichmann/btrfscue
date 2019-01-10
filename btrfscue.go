/*
 * btrfscue version 0.5
 * Copyright (c)2011-2019 Christian Blichmann
 *
 * Recover data from damaged BTRFS filesystems
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

package main

import (
	"flag"
	"fmt"
	"os"

	"blichmann.eu/code/btrfscue/cliutil"
	_ "blichmann.eu/code/btrfscue/identify"
	"blichmann.eu/code/btrfscue/subcommand"
)

var (
	// Global options
	help    = flag.Bool("help", false, "display this help and exit")
	verbose = flag.Bool("verbose", false, "explain what is being done")
	version = flag.Bool("version", false, "display version and exit")
)

type helpCommand struct{}

func (hc *helpCommand) DefineFlags(fs *flag.FlagSet) {}
func (hc *helpCommand) Run(args []string)            { printUsage() }

func init() {
	subcommand.Commands.RegisterHidden("help", &helpCommand{})
}

// printUsage prints GNU-looking usage text.
func printUsage() {
	fmt.Printf("Usage: %s COMMAND [OPTION]...\n"+
		"Recover data from damaged BTRFS filesystems.\n\n"+
		"Commands:\n", os.Args[0])
	subcommand.VisitAll(func(name, desc string, cmd subcommand.Command) {
		fmt.Printf("  %-9s %s\n", name, desc)
	})
	fmt.Printf("\nCommon options:\n")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("      --%-23s %s\n", f.Name, f.Usage)
	})
	fmt.Printf("\nFor bug reporting instructions, please see:\n" +
		"<https://github.com/cblichmann/btrfscue/issues>\n")
}

func main() {
	flag.Usage = func() { /* Disable */ }
	fatalHelp := fmt.Sprintf("Try '%s' --help for more information.",
		os.Args[0])

	flag.Parse()
	if *help {
		printUsage()
		os.Exit(0)
	}
	if *version {
		fmt.Printf("btrfscue 0.5\n" +
			"Copyright (c)2011-2019 Christian Blichmann\n" +
			"This software is BSD licensed, see the source for copying " +
			"conditions.\n\n")
		os.Exit(0)
	}
	cliutil.SetVerbose(*verbose)
	if flag.NArg() == 0 {
		cliutil.Fatalf("missing command\n%s\n", fatalHelp)
	}

	startProfiling()
	defer stopProfiling()

	subcommand.Commands.SetGlobalFlags(flag.CommandLine)
	subcommand.Parse(flag.Args())
	subcommand.Run()
}
