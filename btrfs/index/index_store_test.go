/*
 * btrfscue version 0.3
 * Copyright (c)2011-2017 Christian Blichmann
 *
 * Tests for the index data structure
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

package index // import "blichmann.eu/code/btrfscue/btrfs/index"

import (
	"math/rand"
	"testing"
	"time"
)

func TestKeyCompare(t *testing.T) {
	low := Key{ObjectID: RootTreeDirObjectId, Type: DirItemKey, Offset: 0}
	high := Key{ObjectID: FirstFreeObjectId, Type: ExtentItemKey, Offset: 100}
	if c := KeyCompare(low, low); c != 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}
	if c := KeyCompare(low, high); c >= 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}
	if c := KeyCompare(high, low); c <= 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}
	if c := KeyCompare(high, high); c != 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}

	low = Key{RootTreeObjectId, InodeRefKey, 6}
	high = Key{FSTreeObjectId, InodeRefKey, 6}
	if c := KeyCompare(low, high); c >= 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}
}

// Data from a small real-world filesystem with all extent data removed.
var indexItems []*Item = []*Item{
	&Item{Key{RootTreeObjectId, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{257, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{258, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{259, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{260, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{261, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{262, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{263, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{264, InodeItemKey, 0}, 0, 0, nil},
	&Item{Key{FSTreeObjectId, InodeRefKey, 6}, 0, 0, nil},
	&Item{Key{RootTreeDirObjectId, InodeRefKey, 6}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, InodeRefKey, 256}, 0, 0, nil},
	&Item{Key{257, InodeRefKey, 256}, 0, 0, nil},
	&Item{Key{258, InodeRefKey, 257}, 0, 0, nil},
	&Item{Key{259, InodeRefKey, 256}, 0, 0, nil},
	&Item{Key{260, InodeRefKey, 259}, 0, 0, nil},
	&Item{Key{261, InodeRefKey, 260}, 0, 0, nil},
	&Item{Key{262, InodeRefKey, 256}, 0, 0, nil},
	&Item{Key{263, InodeRefKey, 256}, 0, 0, nil},
	&Item{Key{264, InodeRefKey, 259}, 0, 0, nil},
	&Item{Key{RootTreeObjectId, DirItemKey, 0x8dbfc2d2}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirItemKey, 0x7c67c22b}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirItemKey, 0x8244f324}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirItemKey, 0xc53a6a73}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirItemKey, 0xcc4ebf03}, 0, 0, nil},
	&Item{Key{257, DirItemKey, 0xc4b0d86b}, 0, 0, nil},
	&Item{Key{259, DirItemKey, 0xe125a1}, 0, 0, nil},
	&Item{Key{259, DirItemKey, 0xf66986c2}, 0, 0, nil},
	&Item{Key{260, DirItemKey, 0x8b806efd}, 0, 0, nil},
	&Item{Key{262, DirItemKey, 0xb349a3fd}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirIndexKey, 2}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirIndexKey, 3}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirIndexKey, 5}, 0, 0, nil},
	&Item{Key{FirstFreeObjectId, DirIndexKey, 6}, 0, 0, nil},
	&Item{Key{257, DirIndexKey, 2}, 0, 0, nil},
	&Item{Key{259, DirIndexKey, 2}, 0, 0, nil},
	&Item{Key{259, DirIndexKey, 3}, 0, 0, nil},
	&Item{Key{260, DirIndexKey, 2}, 0, 0, nil},
	&Item{Key{262, DirIndexKey, 3}, 0, 0, nil},
	&Item{Key{FSTreeObjectId, RootItemKey, 0}, 0, 0, nil},
	&Item{Key{257, RootItemKey, 0}, 0, 0, nil},
	&Item{Key{257, RootBackRefKey, 5}, 0, 0, nil},
	&Item{Key{FSTreeObjectId, RootRefKey, 257}, 0, 0, nil},
	// Real chunk data, including stripes
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0x0}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0x0}}}},
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0x400000}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0x400000}}}},
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0xc00000}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0xc00000}}}},
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0x1400000}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0x1400000}, {Offset: 0x1c00000}}}},
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0x1c00000}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0x2400000}, {Offset: 0x4400000}}}},
	&Item{Key{FirstFreeObjectId, ChunkItemKey, 0x3c00000}, 0, 0,
		&Chunk{Stripes: []Stripe{{Offset: 0x6400000}}}},
}

func makeIndex(items []*Item) Index {
	index := NewIndex()
	for _, item := range items {
		index.Insert(item)
	}
	return index
}

func TestInsert(t *testing.T) {
	index := makeIndex(indexItems)
	for i, item := range indexItems {
		if KeyCompare(item.Key, index.Key(i)) != 0 {
			t.Fatalf("%s vs %s", item.Key, index.Key(i))
		}
	}
}

func TestIndexRandomInsert(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := NewIndex()
	for index.Len() < len(indexItems) {
		index.Insert(indexItems[r.Intn(len(indexItems))])
	}
	for i, item := range indexItems {
		if KeyCompare(item.Key, index.Key(i)) != 0 {
			t.Fatalf("%s vs %s", item.Key, index.Key(i))
		}
	}
}

func TestFind(t *testing.T) {
	index := makeIndex(indexItems)

	// TODO(cblichmann): Instead of logging, actually test
	var i int
	i = index.LowerBound(Key{})
	t.Logf("%d %d", i, index.Len())
	i = index.LowerBound(Key{^uint64(0), 255, ^uint64(0)})
	t.Logf("%d %d", i, index.Len())
	i = index.LowerBound(Key{258, DirIndexKey, 0})
	t.Logf("%d %d %s", i, index.Len(), index.Key(i))
	t.Log("")

	i = index.Find(Key{})
	t.Logf("%d %d", i, index.Len())
	i = index.Find(Key{^uint64(0), 255, ^uint64(0)})
	t.Logf("%d %d", i, index.Len())
	i = index.Find(Key{258, DirIndexKey, 0})
	t.Logf("%d %d", i, index.Len())
	i, end := index.Range(KeyFirst(DirIndexKey, 258),
		KeyLast(DirIndexKey, 258))
	t.Logf("%d %d %t %d", i, index.Len(), i < end, end)
}

func TestPhysical(t *testing.T) {
	index := makeIndex(indexItems)

	// Exact chunk offsets
	chunkMap := map[uint64]uint64{
		0x0:       0x0,
		0x400000:  0x400000,
		0xc00000:  0xc00000,
		0x1400000: 0x1c00000,
		0x1c00000: 0x2400000,
		0x3c00000: 0x6400000,
	}
	for logical, physical := range chunkMap {
		if _, p := index.Physical(logical); p != physical {
			t.Logf("0x%x => 0x%x: 0x%x", logical, physical, p)
		}
	}
}
