// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Recover data from damaged BTRFS filesystems

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/pkg/btrfs"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "btrfscue",
	Short:   "Recover data from damaged BTRFS filesystems.",
	Version: "0.0", // Set via SetVersionTemplate()
}

func init() {
	global := &app.Global
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cliutil.SetVerbose(global.Verbose)
	}

	rootCmd.SetVersionTemplate(`btrfscue 0.6
Copyright (c)2011-2026 Christian Blichmann
This software is BSD licensed, see the source for copying conditions.
`)
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
For bug reporting instructions, please see:
<https://github.com/cblichmann/btrfscue/issues>
`)

	fs := rootCmd.PersistentFlags()
	fs.BoolVarP(&global.Verbose, "verbose", "v", false,
		"explain what is being done")
	fs.BoolVarP(&global.Progress, "progress", "p", true,
		"display visual progress")
	fs.BoolVarP(&global.Machine, "machine", "m", false,
		"display machine parseable output")
	fs.UintVar(&global.BlockSize, "block-size", btrfs.DefaultBlockSize,
		"filesystem block size")
	fs.StringVar(&global.Metadata, "metadata", os.Getenv("BTRFSCUE_METADATA"),
		"metadata database to use")
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main().
func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
