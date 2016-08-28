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
	"encoding/gob"
	"flag"
	"io"
	"os"

	"blichmann.eu/code/btrfscue/btrfs"
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

// WriteIndex writes the filesystem metadata index to the specified Writer.
// It uses the Gob enconding to write out the actual elements.
func WriteIndex(w io.Writer, fs *btrfs.Index) error {
	enc := gob.NewEncoder(w)
	for i := 0; i < fs.Len(); i++ {
		if err := enc.Encode(fs.Item(i)); err != nil {
			fatalf("%s\n", err)
			return err
		}
	}
	return nil
}

// ReadIndex reads a Gob encoded filesystem metadata index from a Reader.
func ReadIndex(r io.Reader, fs *btrfs.Index) error {
	dec := gob.NewDecoder(r)
	for {
		item := &btrfs.Item{}
		if err := dec.Decode(&item); err == nil {
			fs.Insert(item)
		} else if err != io.EOF {
			return err
		} else {
			break
		}
	}
	return nil
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

	filename := args[0]
	f, err := os.Open(filename)
	reportError(err)
	defer f.Close()

	mode := os.O_RDWR | os.O_CREATE
	if !c.append {
		mode |= os.O_TRUNC
	}
	m, err := os.OpenFile(*metadata, mode, 0666)
	reportError(err)
	defer m.Close()

	bs := uint64(*blockSize)

	devSize, err := CheckedBtrfsDeviceSize(f, bs)
	reportError(err)

	buf := make([]byte, bs)
	b := btrfs.NewParseBuffer(buf)
	l := btrfs.Leaf{}

	fs := btrfs.NewIndex()

	// Start right after the first superblock
	for offset := uint64(btrfs.SuperInfoOffset) + bs; offset < devSize &&
		err != io.EOF; offset += bs {
		reportError(ReadBlockAt(f, buf, offset, bs))
		b.Rewind()
		l.Header.Parse(b)
		// Skip this header if it has the wrong FSID or is empty.
		if l.FSID != c.id || l.NrItems == 0 {
			continue
		}
		// Also skip all non-leaves (although they should never be stored on
		// disk).
		if !l.IsLeaf() {
			// TODO(cblichmann): Find out why this sometimes happens
			//warnf("found non-leaf at offset %d, level %d\n", offset, l.Level)
			continue
		}
		if l.Header.ByteNr != offset {
			// TODO(cblichmann): Are these backup leaves?
			//warnf("expected leaf offset %d, got %d\n", offset,
			//	l.Header.ByteNr)
		}
		// Given he := uint32(b.Offset()), after the following call to
		// Parse(), the free space of a leaf is between offsets
		// [ he, l.Items[Len(l.Items) - 1].Offset ).
		l.Parse(b)
		for i := range l.Items {
			fs.Insert(&l.Items[i])
			//item := &l.Items[i]
			//if ii, ok := item.Data.(*btrfs.InodeItem); ok {
			//	if item.Key.ObjectID != 264 {
			//		continue
			//	}
			//	verbosef("%s %d %d %d %d %d\n", item.Key, ii.Size, ii.BlockGroup, ii.Generation, ii.Transid, ii.Sequence)
			//}
		}
	}

	// TODO(cblichmann): Add callback to allow to display progress.
	reportError(WriteIndex(m, &fs))
	for i := 0; i < fs.Len(); i++ {
		verbosef("%s\n", fs.Key(i))
	}
}

func init() {
	subcommand.Register("recon", "gather metadata for later use",
		&reconCommand{})
}
