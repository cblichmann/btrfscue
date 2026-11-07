// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Sub-command to gather metadata

package cmd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
	"blichmann.eu/code/btrfscue/pkg/uuid"
)

type scanFSOptions struct {
	id     uuid.UUID
	append bool
}

func init() {
	options := scanFSOptions{}
	reconCmd := &cobra.Command{
		Use:   "recon",
		Short: "gather metadata for later use",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(app.Global.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doScanFS(args[0], app.Global.Metadata, options)
		},
	}

	fs := reconCmd.PersistentFlags()
	fs.Var(&options.id, "id", "UUID of the filesystem (see identify)")
	fs.BoolVar(&options.append, "append", false, "append to metadata file")

	rootCmd.AddCommand(reconCmd)
}

func doScanFS(filename, metadata string, options scanFSOptions) {
	if options.id.IsZero() {
		cliutil.Fatalf("missing id option\n")
	}

	f, err := os.Open(filename)
	cliutil.ReportError(err)
	defer f.Close()

	bs := uint64(app.Global.BlockSize)

	devSize, err := btrfs.CheckDeviceSize(f, bs)
	cliutil.ReportError(err)
	devSize = devSize - (devSize % bs)

	startOff := uint64(btrfs.SuperInfoOffset) + bs
	_, err = f.Seek(int64(startOff), 0)
	cliutil.ReportError(err)

	reader := bufio.NewReaderSize(f, 64*1024*1024) // 64MB buffer

	buf := make([]byte, bs)

	ix, err := index.Open(app.Global.Metadata, 0644, &index.Options{
		BlockSize:  uint(bs),
		FSID:       options.id,
		Generation: ^uint64(0),
	})
	cliutil.ReportError(err)
	defer func() {
		cliutil.ReportError(ix.Commit())
		ix.Close()
	}()

	bar := pb.New64(int64(devSize)) //.SetUnits(pb.U_BYTES)
	bar.SetMaxWidth(120)
	if app.Global.Progress {
		bar.Start()
	}
	defer bar.Finish()

	for off := startOff; off < devSize; off += bs {
		_, err = io.ReadFull(reader, buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			cliutil.ReportError(err)
		}
		if (off / bs) % 10000 == 0 {
			bar.SetCurrent(int64(off))
		}

		// Offset 32:48 is FSID. options.id is a uuid.UUID which is [16]byte.
		if !bytes.Equal(buf[32:48], options.id[:]) {
			continue
		}

		// Offset 96:100 is NrItems (uint32). Little endian.
		nrItems := binary.LittleEndian.Uint32(buf[96:100])
		if nrItems == 0 {
			continue
		}

		// Offset 100 is Level (uint8).
		level := buf[100]
		if level != 0 {
			continue
		}

		l := btrfs.Leaf(buf)
		h := l.Header()

		for i := 0; i < l.Len(); i++ {
			cliutil.ReportError(ix.InsertItem(l.Key(i), h, l.Item(i),
				l.Data(i)))
		}
	}
	bar.SetCurrent(int64(devSize))

	bar.Finish()
}
