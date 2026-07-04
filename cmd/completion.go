// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Sub-command to generate tab completion scripts

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "generates bash completion scripts",
		Long: `Generates bash completion scripts.

To load completion run

. <(btrfscue completion)

To configure your bash shell to load completions for each session add to your
.bashrc:

# ~/.bashrc or ~/.profile
. <(btrfscue completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	rootCmd.AddCommand(completionCmd)
}
