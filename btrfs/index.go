/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * BTRFS filesystem index data structure
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

package btrfs // import "blichmann.eu/code/btrfscue/btrfs"

import (
	"hash/crc32"
	"sort"

	"blichmann.eu/code/btrfscue/ordered"
)

// KeyCompare compares two BTRFS keys lexicographically. It returns 0 if
// a==b, -1 if a < b and +1 if a > b.
func KeyCompare(a, b Key) int {
	if r := int(a.Type) - int(b.Type); r != 0 {
		return r
	}
	if a.ObjectID < b.ObjectID {
		return -1
	}
	if a.ObjectID > b.ObjectID {
		return 1
	}
	if a.Offset < b.Offset {
		return -1
	}
	if a.Offset > b.Offset {
		return 1
	}
	return 0
}

type index struct {
	items ordered.Set
}

func NewIndex() *index {
	return &index{ordered.NewSet(func(a, b interface{}) int {
		return KeyCompare(a.(*Item).Key, b.(*Item).Key)
	})}
}

func (idx *index) Len() int { return idx.items.Len() }

func (idx *index) Insert(item *Item) *Item {
	if _, ok := idx.items.Insert(item); ok {
		return item
	}
	return nil
}

func (idx *index) Find(key Key) int {
	return sort.Search(idx.Len(), func(i int) bool {
		return KeyCompare(idx.KeyAt(i), key) >= 0
	})
}

func (idx *index) ItemAt(i int) *Item { return idx.items.At(i).(*Item) }
func (idx *index) KeyAt(i int) Key    { return idx.ItemAt(i).Key }

// Range returns an index range [low, high) for the given keys.
func (idx *index) Range(first, last Key) (low, high int) {
	low = idx.Find(first)
	high = low + sort.Search(idx.Len()-low, func(i int) bool {
		return KeyCompare(idx.KeyAt(low+i), last) > 0
	})
	return
}

// RangeSubvolumes returns an index range containing all of the subvolumes.
// Note that this excludes the FS tree root itself, which is not considered
// to be a subvolume. To access the FS tree root, use
//   Find(KeyFirst(RootItemKey, FSTreeObjectId))
func (idx *index) RangeSubvolumes() (low, high int) {
	return idx.Range(KeyFirst(RootItemKey, FirstFreeObjectId),
		KeyLast(RootItemKey, LastFreeObjectId))
}

func keyFirst(key *Key, args []uint64) {
	switch len(args) {
	case 3:
		key.Offset = args[2]
		fallthrough
	case 2:
		key.ObjectID = args[1]
		fallthrough
	case 1:
		key.Type = uint8(args[0])
	case 0:
	default:
		panic("invalid number of arguments")
	}
}

// KeyFirst returns a new key useful for searching an index. To enable more
// natural queries, the order of the arguments is different from the one in
// the Key struct: Type, ObjectID, Offset.
func KeyFirst(args ...uint64) (key Key) {
	keyFirst(&key, args)
	return
}

func KeyLast(args ...uint64) (key Key) {
	key.ObjectID = LastFreeObjectId
	key.Type = ^uint8(0)
	key.Offset = ^uint64(0)
	keyFirst(&key, args)
	return
}

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// NameHash computes the CRC32C value of an FS object name. The initial CRC
// in BTRFS is ^uint32(1) and the end result is _not_ inverted.
func NameHash(name string) uint32 {
	return ^crc32.Update(^uint32(0xFFFFFFFE), castagnoliTable, []byte(name))
}
