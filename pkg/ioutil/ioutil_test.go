// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Tests for I/O utility routines

package ioutil

import (
	"bytes"
	"io"
	"testing"
)

func TestReadBlockAt(t *testing.T) {
	r := bytes.NewReader([]byte("0123456789abcdefghij")) // 20 bytes
	buf := make([]byte, 10)
	var expected []byte

	if err := ReadBlockAt(r, buf, 0); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	expected = []byte("0123456789")
	if !bytes.Equal(buf, expected) {
		t.Fatalf("%s vs %s", expected, buf)
	}

	if err := ReadBlockAt(r, buf, 10); err != nil {
		// Ensure we don't get EOF at the end
		t.Fatalf("nil vs %s", err)
	}
	expected = []byte("abcdefghij")
	if !bytes.Equal(buf, expected) {
		t.Fatalf("%s vs %s", expected, buf)
	}

	if err := ReadBlockAt(r, buf, 11); err != io.EOF {
		t.Fatalf("nil vs %s", err)
	}
}
