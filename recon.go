/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
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

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/uuid"
)

//func (idx *index) ItemAt(i int) *btrfs.Item {
//	return idx.At(i).(*btrfs.Item)
//}

//func (idx *index) KeyAt(i int) *btrfs.Key {
//	return &idx.At(i).(*btrfs.Item).Key
//}

type reconCommand struct {
	id uuid.UUID
}

func (c *reconCommand) DefineFlags(fs *flag.FlagSet) {
	fs.Var(&c.id, "id", "UUID of the filesystem (see identify)")
}

func (c *reconCommand) Run(args []string) {
	//i2 := ordered.NewMultiSet(ordered.IntCompare, 10, 20, 30, 30, 20, 10, 10, 20)
	//for i := range i2.Data() {
	//	verbosef("%d ", i2.IntAt(i))
	//}
	//verbosef("\n")
	//
	//low := i2.LowerBound(0, i2.Len(), 20)
	//high := i2.UpperBound(low, i2.Len(), 20)
	//verbosef("%d %d\n", low, high)
	//
	//return

	if len(args) == 0 {
		fatalf("missing device file\n")
	}
	if len(args) > 1 {
		fatalf("extra operand '%s'\n", args[1])
	}

	filename := args[0]
	f, err := os.Open(filename)
	reportError(err)
	defer f.Close()

	bs := uint64(*blockSize)

	devSize, err := CheckedBtrfsDeviceSize(f, bs)
	reportError(err)

	buf := make([]byte, bs)
	b := btrfs.NewParseBuffer(buf)
	l := btrfs.Leaf{}

	index := btrfs.NewIndex()

	// Start right after the superblock
	for offset := uint64(btrfs.SuperInfoOffset) + bs; offset < devSize &&
		err != io.EOF; offset += bs {
		reportError(ReadBlockAt(f, buf, offset, bs))
		b.Rewind()
		l.Header.Parse(b)
		//headerEnd := uint32(b.Offset())
		// Skip this header if it has the wrong FSID or is empty.
		if l.FSID != c.id || l.NrItems == 0 {
			continue
		}
		// Also skip all non-leaves (although they should never be stored on
		// disk).
		if !l.IsLeaf() {
			// TODO(cblichmann): Find out why this sometimes happens
			warnf("found non-leaf at offset %d, level %d\n", offset, l.Level)
			continue
		}
		if l.Header.ByteNr != offset {
			// TODO(cblichmann): Are these backup leaves?
			//warnf("expected leaf offset %d, got %d\n", offset, l.Header.ByteNr)
		}
		l.Parse(b)
		for i := range l.Items {
			index.Insert(&l.Items[i])
		}
		//switch data := item.Data.(type) {
		//case *btrfs.FileExtentItem:
		//verbosef("file extent item: %s", data.Data)
		//case *btrfs.RootRef:
		//	verbosef("root ref %s %s\n", btrfs.KeyTypeString(item.Type), data.Name)
		//case *btrfs.InodeItem:
		//	verbosef("inode %d\n", data.Generation)
		//case *btrfs.InodeRefItem:
		//	verbosef("inode ref %s\n", data.Name)
		//case *btrfs.DirItem:
		//	verbosef("dir item %s\n", data.Name)
		//case *btrfs.BlockGroupItem:
		//	verbosef("block %d\n", data.ChunkObjectID)
		//default:
		//	//warnf("unknown item key type at offset %d: %d\n", offset, item.Type)
		//}
	}

	for i, end := index.RangeSubvolumes(); i < end; i++ {
		item := index.ItemAt(i)
		data := item.Data.(*btrfs.RootItem)
		verbosef("root item %d %d %s\n", data.RootDirID, data.Generation,
			item.Key)
	}
	verbosef("\n")
	first := btrfs.KeyFirst(btrfs.DirItemKey, 256)
	last := btrfs.KeyLast(btrfs.DirItemKey, 256)
	verbosef("\n")
	for i, end := index.Range(first, last); i < end; i++ {
		item := index.ItemAt(i)
		data := item.Data.(*btrfs.DirItem)
		verbosef("dir item %s %s\n", data.Name, data.Location)
	}
	verbosef("\n")
}
