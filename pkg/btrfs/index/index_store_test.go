/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
 *
 * Tests for BTRFS index
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
