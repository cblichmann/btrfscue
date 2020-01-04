// +build linux darwin

/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Sub-command to provide and mount a "rescue fs"
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
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
	"blichmann.eu/code/btrfscue/rescuefs"
)

func init() {
	mountCmd := &cobra.Command{
		Use:   "mount",
		Short: "provide a 'rescue' filesystem backed by metadata",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(btrfscue.Options.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doMountRescueFS(args, btrfscue.Options.Metadata)
		},
	}

	rootCmd.AddCommand(mountCmd)
}

func doMountRescueFS(args []string, metadata string) {
	ix, err := index.OpenReadOnly(metadata)
	cliutil.ReportError(err)
	defer ix.Close()

	mountPoint := args[len(args)-1]
	var dev *os.File
	defer func() {
		if dev != nil {
			dev.Close()
		}
	}()
	if len(args) == 2 {
		dev, err = os.Open(args[0])
		cliutil.ReportError(err)
	} else {
		cliutil.Warnf("no device file given, only inline file data will be " +
			"visible\n")
	}

	fs := rescuefs.New(btrfscue.Options.Metadata, ix, dev)
	cliutil.ReportError(fs.Mount(mountPoint))
	cliutil.Verbosef("mounted rescue FS on %s\n", mountPoint)
	go fs.Serve()

	// Break and unmount on CTRL+C or TERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	_ = <-ch
	cliutil.Warnf("got signal, unmounting...\n")
	cliutil.ReportError(fs.Unmount())
}
