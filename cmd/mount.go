// +build linux darwin

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Sub-command to provide and mount a "rescue fs"

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/internal/rescuefs"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
)

func init() {
	mountCmd := &cobra.Command{
		Use:   "mount",
		Short: "provide a 'rescue' filesystem backed by metadata",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(app.Global.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doMountRescueFS(args, app.Global.Metadata)
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

	fs := rescuefs.New(app.Global.Metadata, ix, dev)
	cliutil.ReportError(fs.Mount(mountPoint))
	cliutil.Verbosef("mounted rescue FS on %s\n", mountPoint)
	go fs.Serve()

	// Break and unmount on CTRL+C or TERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	cliutil.Warnf("got signal, unmounting...\n")
	cliutil.ReportError(fs.Unmount())
}
