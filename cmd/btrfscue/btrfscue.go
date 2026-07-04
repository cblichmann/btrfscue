// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Recover data from damaged BTRFS filesystems, CLI entry point

package main

import (
	"blichmann.eu/code/btrfscue/cmd"
)

func main() {
	startProfiling()
	defer stopProfiling()

	cmd.Execute()
}
