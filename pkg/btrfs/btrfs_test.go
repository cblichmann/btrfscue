// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Tests for the utility functions of the FS structures package

package btrfs

import (
	"testing"
)

func TestKeyCompare(t *testing.T) {
	low := Key{ObjectID: RootTreeDirObjectID, Type: DirItemKey, Offset: 0}
	high := Key{ObjectID: FirstFreeObjectID, Type: ExtentItemKey, Offset: 100}
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

	low = Key{RootTreeObjectID, InodeRefKey, 6}
	high = Key{FSTreeObjectID, InodeRefKey, 6}
	if c := KeyCompare(low, high); c >= 0 {
		t.Fatalf("%s vs %s: %d", low, high, c)
	}
}
