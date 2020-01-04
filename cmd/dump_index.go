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

package cmd // import "blichmann.eu/code/btrfscue/cmd"

import (
	"fmt"

	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"

	"github.com/spf13/cobra"
)

func init() {
	dumpIndexCmd := &cobra.Command{
		Use:   "dump-index",
		Short: "for debugging, dump the index in text format",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(btrfscue.Options.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doDumpIndex(btrfscue.Options.Metadata)
		},
	}

	rootCmd.AddCommand(dumpIndexCmd)
}

func doDumpIndex(metadata string) {
	ix, err := index.OpenReadOnly(metadata)
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
