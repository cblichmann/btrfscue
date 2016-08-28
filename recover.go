/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Sub-command to recover data from filesystem images using metadata
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
	"os"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/subcommand"
)

type recoverCommand struct {
	clobber *bool
}

func (rc *recoverCommand) DefineFlags(fs *flag.FlagSet) {
	rc.clobber = fs.Bool("clobber", false,
		"overwrite existing files")
}

func (rc *recoverCommand) Run([]string) {
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}

	m, err := os.Open(*metadata)
	reportError(err)
	defer m.Close()

	fs := btrfs.NewIndex()
	reportError(ReadIndex(m, &fs))

	inode := uint64(264) // src.zip
	ii := fs.InodeItem(fs.FindInodeItem(inode))
	verbosef("file size: %d\n", ii.Size)

	lo := uint64(0)
	for i, end := fs.Range(
		btrfs.KeyFirst(btrfs.ExtentDataKey, inode),
		btrfs.KeyLast(btrfs.ExtentDataKey, inode)); i < end; i++ {
		fe := fs.Item(i).Data.(*btrfs.FileExtentItem)
		lo = fe.DiskByteNr
		verbosef("file extent: %s %d %d %d %d\n", fs.Key(i),
			lo, fe.DiskNumBytes, fe.NumBytes, fe.Offset)
		_, of := fs.Physical(lo)
		verbosef("%d\n", of)
	}

	for i, end := fs.Range(
		btrfs.KeyFirst(btrfs.ChunkItemKey),
		btrfs.KeyLast(btrfs.ChunkItemKey)); i < end; i++ {
		c := fs.Chunk(i)
		verbosef("%s %d 0x%x\n", fs.Key(i), c.Length, c.Stripes[0].Offset)
	}
}

func init() {
	subcommand.Register("recover",
		"try to restore files from a damaged filesystem", &recoverCommand{})
}
