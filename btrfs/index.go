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
	"strings"

	"blichmann.eu/code/btrfscue/ordered"

	"fmt"
)

func init() { fmt.Printf("") }

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

type Index struct {
	items ordered.Set
}

func NewIndex() Index {
	return Index{ordered.NewSet(func(a, b interface{}) int {
		return KeyCompare(a.(*Item).Key, b.(*Item).Key)
	})}
}

func (fs *Index) Len() int { return fs.items.Len() }

func (fs *Index) Insert(item *Item) (int, bool) {
	return fs.items.Insert(item)
}

func (fs *Index) Find(k Key) int {
	if i := fs.LowerBound(k); i < fs.Len() && KeyCompare(fs.Key(i), k) == 0 {
		return i
	}
	return fs.Len()
}

func (fs *Index) LowerBound(k Key) int {
	return sort.Search(fs.Len(), func(i int) bool {
		return KeyCompare(fs.Key(i), k) >= 0
	})
}

// FindDirItem finds the position of the DirItem for a given FS path.
func (fs *Index) FindDirItem(rootDirID uint64, path string) int {
	result := fs.Len()
	dirID := rootDirID
	pathComps := strings.Split(strings.TrimLeft(path, "/"), "/")
	for i, comp := range pathComps {
		found := false
		last := i == len(pathComps)-1
		// Lookup items in directory (dirID, BTRFS_DIR_INDEX, ?)
		for j, end := fs.Range(KeyFirst(DirIndexKey, dirID),
			KeyLast(DirIndexKey, dirID)); j < end; j++ {
			item := fs.DirItem(j)
			found = item.Name == comp && (item.Type == FtDir || last)
			if found {
				result = j
				dirID = item.Location.ObjectID
				break
			}
		}
		if !found {
			return fs.Len()
		}
	}
	return result
}

func (fs *Index) FindInodeItem(inode uint64) int {
	return fs.Find(KeyFirst(InodeItemKey, inode))
}

func (fs *Index) FindInodeRefItem(inode uint64) int {
	if i, end := fs.Range(KeyFirst(InodeRefKey, inode),
		KeyLast(InodeRefKey, inode)); i < end {
		return i
	}
	return fs.Len()
}

// Item returns the item at a specific position in the index.
func (fs *Index) Item(i int) *Item { return fs.items.At(i).(*Item) }

// Key returns the BTRFS key for an index position.
func (fs *Index) Key(i int) Key { return fs.Item(i).Key }

// DirItem returns the directory item for an index.
func (fs *Index) DirItem(i int) *DirItem { return fs.Item(i).Data.(*DirItem) }

// InodeItem returns the inode item for an index.
func (fs *Index) InodeItem(i int) *InodeItem {
	return fs.Item(i).Data.(*InodeItem)
}

func (fs *Index) InodeRefItem(i int) *InodeRefItem {
	return fs.Item(i).Data.(*InodeRefItem)
}

func (fs *Index) Chunk(i int) *Chunk {
	return fs.Item(i).Data.(*Chunk)
}

func (fs *Index) Physical(logical uint64) (devID uint64, offset uint64) {
	if low, high := fs.Range(KeyFirst(ChunkItemKey, FirstChunkTreeObjectId),
		KeyLast(ChunkItemKey, FirstChunkTreeObjectId)); low < high {
		if i := low + sort.Search(high-low, func(i int) bool {
			i += low
			return int64(logical-fs.Key(i).Offset) < int64(fs.Chunk(i).Length)
		}); i < high {
			stripe := fs.Chunk(i).Stripes[0]
			return stripe.DevID, stripe.Offset + logical - fs.Key(i).Offset
		}
	}
	return
}

// Range returns an index range [low, high) for the given keys.
func (fs *Index) Range(first, last Key) (low, high int) {
	low = fs.LowerBound(first)
	high = low + sort.Search(fs.Len()-low, func(i int) bool {
		return KeyCompare(fs.Key(low+i), last) > 0
	})
	return
}

// RangeSubvolumes returns an index range containing all of the subvolumes.
// Note that this excludes the FS tree root itself, which is not considered
// to be a subvolume. To access the FS tree root, use
//   Find(KeyFirst(RootItemKey, FSTreeObjectId))
func (fs *Index) RangeSubvolumes() (low, high int) {
	return fs.Range(KeyFirst(RootItemKey, FirstFreeObjectId),
		KeyLast(RootItemKey, LastFreeObjectId))
}

func keyFirst(k *Key, args []uint64) {
	switch len(args) {
	case 3:
		k.Offset = args[2]
		fallthrough
	case 2:
		k.ObjectID = args[1]
		fallthrough
	case 1:
		k.Type = uint8(args[0])
	case 0:
	default:
		panic("invalid number of arguments")
	}
}

// KeyFirst returns a new key useful for searching an index. To enable more
// natural queries, the order of the arguments is different from the one in
// the Key struct: Type, ObjectID, Offset.
func KeyFirst(args ...uint64) (k Key) {
	keyFirst(&k, args)
	return
}

func KeyLast(args ...uint64) (k Key) {
	k.ObjectID = LastFreeObjectId
	k.Type = ^uint8(0)
	k.Offset = ^uint64(0)
	keyFirst(&k, args)
	return
}

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// NameHash computes the CRC32C value of an FS object name. The initial CRC
// in BTRFS is ^uint32(1) and the end result is _not_ inverted.
func NameHash(name string) uint32 {
	return ^crc32.Update(^uint32(0xFFFFFFFE), castagnoliTable, []byte(name))
}
