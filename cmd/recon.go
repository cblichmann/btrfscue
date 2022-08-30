/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
 *
 * Sub-command to gather metadata
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
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
	"blichmann.eu/code/btrfscue/pkg/ioutil"
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

	// Start right after the first superblock
	for off := uint64(btrfs.SuperInfoOffset) + bs; off < devSize; off += bs {
		if err = ioutil.ReadBlockAt(f, buf, off); err == io.EOF {
			break
		} else if err != nil {
			cliutil.ReportError(err)
		}
		bar.SetCurrent(int64(off))
		l := btrfs.Leaf(buf)
		h := l.Header()

		// Skip this header if it has the wrong FSID or is empty.
		if h.FSID() != options.id || h.NrItems() == 0 {
			continue
		}
		// Also skip all non-leaves (= nodes)
		if !h.IsLeaf() {
			//cliutil.Warnf("non-leaf %d at offset %d, level %d\n", h.ByteNr(),
			//	off, h.Level())
			continue
		}
		// The free space of a leaf is between offsets
		// [ btrfs.HeaderSize, l.Items(l.Len() - 1).Offset() ).
		for i := 0; i < l.Len(); i++ {
			cliutil.ReportError(ix.InsertItem(l.Key(i), h, l.Item(i),
				l.Data(i)))
		}
	}
	bar.SetCurrent(int64(devSize))

	bar.Finish()
}
