/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Tests for I/O utility routines
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
