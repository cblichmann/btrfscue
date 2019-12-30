/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Sub-command to dump the index contents.
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

	_ "blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
	"blichmann.eu/code/btrfscue/subcommand"
)

type dumpIndexCommand struct {
}

func (c *dumpIndexCommand) DefineFlags(fs *flag.FlagSet) {
}

func (c *dumpIndexCommand) Run(args []string) {
	if len(args) > 0 {
		cliutil.Fatalf("extra operand: %s\n", args[0])
	}
	if len(*btrfscue.Metadata) == 0 {
		cliutil.Fatalf("missing metadata option\n")
	}

	ix, err := index.OpenReadOnly(*btrfscue.Metadata)
	cliutil.ReportError(err)
	defer ix.Close()

	last := ^uint64(0)
	for r, v := ix.FullRange(); r.HasNext(); v = r.Next() {
		if o := r.Owner(); o != last {
			fmt.Printf("owner %d\n", o)
			last = o
		}
		k := r.Key()
		fmt.Printf("%s @ %d\n", k, r.Generation())
		_ = v
	}
}

func init() {
	subcommand.Register("dump-index",
		"for debugging, dump the index in text format",
		&dumpIndexCommand{})
}
