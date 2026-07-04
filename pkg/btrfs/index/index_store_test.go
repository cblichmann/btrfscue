// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Tests for BTRFS index

package index

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"blichmann.eu/code/btrfscue/pkg/uuid"
)

func TestCreation(t *testing.T) {
	td, err := ioutil.TempDir("", "index_store_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	testFile := filepath.Join(td, "index")

	const (
		blockSize  = 4096
		generation = 424242
	)
	fsid := uuid.UUID{0xa7, 0xf3, 0x26, 0x75, 0xa3, 0x26, 0x04, 0xf9, 0x2c,
		0xd1, 0xe4, 0x8b, 0x6f, 0x93, 0x98, 0xe0}

	ix, err := Open(testFile, 0644, &Options{BlockSize: blockSize, FSID: fsid,
		Generation: generation})
	if err != nil {
		t.Fatal(err)
	}
	ix.Close()

	ix, err = OpenReadOnly(testFile)
	if err != nil {
		t.Fatal(err)
	}
	m := ix.Metadata()
	if m.BlockSize() != blockSize {
		t.Fatalf("%d vs. %d", blockSize, m.BlockSize())
	}
	if m.FSID() != fsid {
		t.Fatalf("%s vs. %s", fsid, m.FSID())
	}
	if m.Generation() != generation {
		t.Fatalf("%d vs. %d", generation, m.Generation())
	}
	ix.Close()
}
