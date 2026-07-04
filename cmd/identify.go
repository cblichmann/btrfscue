// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Identify sub-command

package cmd

import (
	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	"blichmann.eu/code/btrfscue/internal/identify"

	"github.com/spf13/cobra"
)

func init() {
	options := identify.IdentifyFSOptions{}
	identifyCmd := &cobra.Command{
		Use:   "identify",
		Short: "identify a BTRFS filesystem on a device",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.BlockSize = app.Global.BlockSize
			identify.IdentifyFS(args[0], options)
		},
	}

	fs := identifyCmd.PersistentFlags()
	fs.Float64Var(&options.SampleFraction, "sample-fraction", 0.0001,
		"fraction of blocks to sample for filesystem ids")
	fs.UintVar(&options.MinBlocks, "min-blocks", 1000,
		"minimum number of blocks to scan")
	fs.UintVar(&options.MaxBlocks, "max-blocks", 1000000,
		"maximum number of blocks to scan")
	fs.UintVar(&options.MinOccurrence, "min-occurrence", 4,
		"number of occurrences of an id required to report a filesystem")

	rootCmd.AddCommand(identifyCmd)
}
