/*
 * btrfscue version 0.5
 * Copyright (c)2011-2019 Christian Blichmann
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

package main

import (
	"flag"
	"io"
	"os"

	"gopkg.in/cheggaaa/pb.v1"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/subcommand"
	"blichmann.eu/code/btrfscue/uuid"
)

type reconCommand struct {
	id     uuid.UUID
	append bool
}

func (c *reconCommand) DefineFlags(fs *flag.FlagSet) {
	fs.Var(&c.id, "id", "UUID of the filesystem (see identify)")
	fs.BoolVar(&c.append, "append", false, "append to metadata file")
}

func (c *reconCommand) Run(args []string) {
	if len(args) == 0 {
		fatalf("missing device file\n")
	}
	if len(args) > 1 {
		fatalf("extra operand '%s'\n", args[1])
	}
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}
	if c.id.IsZero() {
		fatalf("missing id option\n")
	}

	filename := args[0]
	f, err := os.Open(filename)
	reportError(err)
	defer f.Close()

	bs := uint64(*blockSize)

	devSize, err := CheckedBtrfsDeviceSize(f, bs)
	reportError(err)
	devSize = devSize - (devSize % bs)

	buf := make([]byte, bs)

	ix, err := index.Open(*metadata, 0644, &index.Options{
		BlockSize:  uint(bs),
		FSID:       c.id,
		Generation: ^uint64(0),
	})
	reportError(err)
	defer func() {
		reportError(ix.Commit())
		ix.Close()
	}()

	bar := pb.New64(int64(devSize)).SetUnits(pb.U_BYTES)
	bar.SetMaxWidth(120)
	bar.Start()
	defer bar.Finish()

	// Start right after the first superblock
	for off := uint64(btrfs.SuperInfoOffset) + bs; off < devSize; off += bs {
		if err = ReadBlockAt(f, buf, off, bs); err == io.EOF {
			break
		} else if err != nil {
			reportError(err)
		}
		bar.Set64(int64(off))
		l := btrfs.Leaf(buf)
		h := l.Header()

		// Skip this header if it has the wrong FSID or is empty.
		if h.FSID() != c.id || h.NrItems() == 0 {
			continue
		}
		// Also skip all non-leaves (= nodes)
		if !h.IsLeaf() {
			//warnf("non-leaf %d at offset %d, level %d\n", h.ByteNr(), off,
			//	h.Level())
			continue
		}
		// The free space of a leaf is between offsets
		// [ btrfs.HeaderSize, l.Items(l.Len() - 1).Offset() ).
		for i := 0; i < l.Len(); i++ {
			reportError(ix.InsertItem(l.Key(i), h, l.Item(i), l.Data(i)))
		}
	}
	bar.Set64(int64(devSize))

	bar.Finish()
	ix.Experimental()
}

func init() {
	subcommand.Register("recon", "gather metadata for later use",
		&reconCommand{})
}
