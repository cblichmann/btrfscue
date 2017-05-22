/*
 * btrfscue version 0.4
 * Copyright (c)2011-2017 Christian Blichmann
 *
 * Sub-command to restore data
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

	"blichmann.eu/code/btrfscue/btrfs/index"
	_ "blichmann.eu/code/btrfscue/subcommand"
)

type recoverCommand struct {
	clobber *bool
}

func (c *recoverCommand) DefineFlags(fs *flag.FlagSet) {
	c.clobber = fs.Bool("clobber", false,
		"overwrite existing files")
}

func (c *recoverCommand) Run([]string) {
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}

	ix, err := index.OpenReadOnly(*metadata)
	reportError(err)
	defer ix.Close()

	for r, v := ix.Subvolumes(); r.HasNext(); v = r.Next() {
		//verbosef("ID %d gen %d cgen %d top level - parent_uuid %s received_uuid %s uuid %s\n",
		//	r.Key().ObjectID, v.OTransID(), v.Generation(), v.ParentUUID(), v.ReceivedUUID(), v.UUID())
		verbosef("	item %s itemoff %d itemsize %d\n", r.Key(), r.Item().Offset(), r.Item().Size())
		verbosef("		root data bytenr %d level %d dirid %d refs %d gen %d lastsnap %d\n",
			v.ByteNr(), v.Level(), v.RootDirID(), v.Refs(), v.Generation(), v.LastSnapshot())
	}

	return
	//inode := uint64(264) // src.zip
	//ii := ix.InodeItem(ix.FindInodeItem(inode))
	//verbosef("file size: %d\n", ii.Size)

	//lo := uint64(0)
	//for i, end := ix.Range(
	//	btrfs.KF(btrfs.ExtentDataKey, inode),
	//	btrfs.KL(btrfs.ExtentDataKey, inode)); i < end; i++ {
	//	fe := &btrfs.FileExtentItem{} //ix.Item(i).Data.(*btrfs.FileExtentItem)
	//	lo = fe.DiskByteNr
	//	verbosef("file extent: %s %d %d %d %d\n", ix.Key(i),
	//		lo, fe.DiskNumBytes, fe.NumBytes, fe.Offset)
	//	_, of := ix.Physical(lo)
	//	verbosef("%d\n", of)
	//}

	//for i, end := ix.Range(
	//	btrfs.KF(btrfs.ChunkItemKey),
	//	btrfs.KL(btrfs.ChunkItemKey)); i < end; i++ {
	//	c := ix.Chunk(i)
	//	verbosef("%s %d 0x%x\n", ix.Key(i), c.Length, c.Stripes[0].Offset)
	//}
}

func init() {
	// TODO(cblichmann): This command is not implemented
	//subcommand.Register("recover",
	//	"try to restore files from a damaged filesystem", &recoverCommand{})
}
