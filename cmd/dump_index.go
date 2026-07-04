// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Sub-command to dump the index contents.

package cmd

import (
	"fmt"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"

	"github.com/spf13/cobra"
)

func init() {
	dumpIndexCmd := &cobra.Command{
		Use:   "dump-index",
		Short: "for debugging, dump the index in text format",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(app.Global.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doDumpIndex(app.Global.Metadata)
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
