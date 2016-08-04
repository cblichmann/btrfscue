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
	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/ordered"
	"blichmann.eu/code/btrfscue/uuid"
	"flag"
	"fmt"
	"io"
	"os"
)

type index struct {
	ordered.Set
}

func keyCompare(a, b interface{}) int {
	ai := a.(*btrfs.Item)
	bi := b.(*btrfs.Item)
	if r := int(ai.Type - bi.Type); r != 0 {
		return r
	}
	if ai.ObjectId < bi.ObjectId {
		return -1
	}
	if ai.ObjectId > bi.ObjectId {
		return 1
	}
	if ai.Offset < bi.Offset {
		return -1
	}
	if ai.Offset > bi.Offset {
		return 1
	}
	return 0
}

func NewIndex() *index {
	return &index{ordered.NewSet(keyCompare)}
}

type reconCommand struct {
	id uuid.UUID
}

func (c *reconCommand) DefineFlags(fs *flag.FlagSet) {
	fs.Var(&c.id, "id", "UUID of the filesystem (see identify)")
}

func (c *reconCommand) Run(args []string) {
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

	index := NewIndex()

	// Start right after the superblock
	for offset := uint64(btrfs.SuperInfoOffset) + bs; offset < devSize &&
		err != io.EOF; offset += bs {
		reportError(ReadBlockAt(f, buf, offset, bs))
		b.Rewind()
		l.Header.Parse(b)
		headerEnd := uint32(b.Offset())
		// Skip this header if it has the wrong FSID or is empty. Also skip
		// all non-leaves (although they should never be stored on disk).
		if l.Header.FSID != c.id || l.Header.NrItems == 0 {
			continue
		}
		if !l.Header.IsLeaf() {
			warnf("found non-leaf at offset %d\n", offset)
			continue
		}
		if l.Header.ByteNr != offset {
			// TODO(cblichmann): Are these backup leaves?
			//warnf("expected leaf offset %d, got %d\n", offset, l.Header.ByteNr)
		}
		l.Parse(b)
		for i, _ := range l.Items {
			item := &l.Items[i]
			b.SetOffset(int(headerEnd + item.Offset))

			_, found := index.Insert(item)
			fmt.Printf("%s %t\n", item.Key, found)

			switch data := item.Data.(type) {
			case *btrfs.InodeItem:
				verbosef("inode %d\n", data.Generation)
			case *btrfs.InodeRefItem:
				verbosef("inode ref %s\n", data.Name)
			case *btrfs.DirItem:
				verbosef("dir item %s\n", data.Name)
			case *btrfs.BlockGroupItem:
				verbosef("block %d\n", data.ChunkObjectId)
			default:
				//warnf("unknown item key type at offset %d: %d\n", offset, item.Type)
			}
		}
		//break //DBG!!!
	}
}
