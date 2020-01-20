/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Identify sub-command
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

package cmd

import (
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/identify"

	"github.com/spf13/cobra"
)

func init() {
	options := identify.IdentifyFSOptions{}
	identifyCmd := &cobra.Command{
		Use:   "identify",
		Short: "identify a BTRFS filesystem on a device",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.BlockSize = btrfscue.Options.BlockSize
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
