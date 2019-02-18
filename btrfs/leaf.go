/*
 * btrfscue version 0.5
 * Copyright (c)2011-2019 Christian Blichmann
 *
 * BTRFS filesystem structures - Leaf, Header and Items
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

package btrfs

import (
	"blichmann.eu/code/btrfscue/uuid"
)

type Header []byte

// Header offsets for parsing from byte slice
const (
	headerCSum = 0
	// The following three fields must match struct SuperBlock
	headerFSID   = headerCSum + CSumSize
	headerByteNr = headerFSID + uuid.UUIDSize
	headerFlags  = headerByteNr + 8 // Includes 1 byte backref rev.
	// Allowed to be different from SuperBlock from here on
	headerChunkTreeUUID = headerFlags + 8
	headerGeneration    = headerChunkTreeUUID + uuid.UUIDSize
	headerOwner         = headerGeneration + 8
	headerNrItems       = headerOwner + 8
	headerLevel         = headerNrItems + 4
	HeaderLen           = headerLevel + 1
)

func (h Header) CSum() CSum {
	c := CSum{}
	copy(c[:], h[headerCSum:headerCSum+CSumSize])
	return c
}

// FSID returns the filesystem specific UUID
func (h Header) FSID() uuid.UUID { return SliceUUID(h[headerFSID:]) }

// ByteNr returns the start of this block relative to the begining of the
// backing device
func (h Header) ByteNr() uint64           { return SliceUint64LE(h[headerByteNr:]) }
func (h Header) Flags() uint64            { return SliceUint64LE(h[headerFlags:]) }
func (h Header) ChunkTreeUUID() uuid.UUID { return SliceUUID(h[headerChunkTreeUUID:]) }
func (h Header) Generation() uint64       { return SliceUint64LE(h[headerGeneration:]) }
func (h Header) Owner() uint64            { return SliceUint64LE(h[headerOwner:]) }
func (h Header) NrItems() uint32          { return SliceUint32LE(h[headerNrItems:]) }
func (h Header) Level() uint8             { return uint8(h[headerLevel]) }

func (h Header) IsLeaf() bool { return h.Level() == 0 }

type Item []byte

const (
	itemKey    = 0
	itemOffset = itemKey + KeyLen
	itemSize   = itemOffset + 4
	ItemLen    = itemSize + 4
)

func (i Item) Key() Key       { return SliceKey(i[itemKey:]) }
func (i Item) Offset() uint32 { return SliceUint32LE(i[itemOffset:]) }
func (i Item) Size() uint32   { return SliceUint32LE(i[itemSize:]) }
func (i Item) Data() []byte   { return i[ItemLen : ItemLen+i.Size()] }

type Leaf []byte

func (l Leaf) Header() Header { return Header(l) }

// Len returns the number of items in this leaf.
func (l Leaf) Len() int {
	// Clamp maximum number of items to avoid running OOM in case NrItems is
	// corrupted.
	maxItems := cap(l) / ItemLen
	numItems := l.Header().NrItems()
	if numItems > uint32(maxItems) {
		numItems = uint32(maxItems)
	}
	return int(numItems)
}

func (l Leaf) Items() []Item {
	items := make([]Item, l.Len())
	for i := range items {
		items[i] = l.Item(i)
	}
	return items
}

func (l Leaf) Item(i int) Item {
	o := HeaderLen + i*ItemLen
	return Item(l[o : o+ItemLen])
}

func (l Leaf) Key(i int) Key {
	return SliceKey(l[HeaderLen+i*ItemLen+itemKey:])
}

func (l Leaf) Data(i int) []byte {
	item := l.Item(i)
	o := HeaderLen + item.Offset()
	// Guard against invalid Item lengths.
	leafLen := uint32(len(l))
	e := o + item.Size()
	if e > leafLen {
		e = leafLen
	}
	return l[o:e]
}
