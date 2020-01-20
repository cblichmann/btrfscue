/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
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

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "btrfscue",
	Short:   "Recover data from damaged BTRFS filesystems.",
	Version: "0.0", // Set via SetVersionTemplate()
}

func init() {
	options := &btrfscue.Options
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cliutil.SetVerbose(options.Verbose)
	}

	rootCmd.SetVersionTemplate(`btrfscue 0.6
Copyright (c)2011-2020 Christian Blichmann
This software is BSD licensed, see the source for copying conditions.
`)
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
For bug reporting instructions, please see:
<https://github.com/cblichmann/btrfscue/issues>
`)

	fs := rootCmd.PersistentFlags()
	fs.BoolVar(&options.Verbose, "verbose", false, "explain what is being done")
	fs.UintVar(&options.BlockSize, "block-size", btrfs.DefaultBlockSize,
		"filesystem block size")
	fs.StringVar(&options.Metadata, "metadata", os.Getenv("BTRFSCUE_METADATA"),
		"metadata database to use")
}

// Execute adds all child commands to the root command and sets flags
// appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
