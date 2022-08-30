/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
 *
 * Tests for the identify sub-command
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

package identify

import (
	"testing"

	"blichmann.eu/code/btrfscue/pkg/btrfs"
)

func TestMakeSampleOffsets(t *testing.T) {
	const (
		numSamples = 3000
		devSize    = 320 << 20 /*320MiB*/
		blockSize  = btrfs.DefaultBlockSize
	)
	samples := MakeSampleOffsets(devSize, blockSize,
		numSamples)
	if len(samples) != numSamples {
		t.Errorf("expected %d, actual %d", numSamples, len(samples))
	}
	last := uint64(devSize)
	for i, o := range samples {
		if o >= devSize-blockSize {
			t.Fatalf("index out of range %d > %d", o, devSize-blockSize)
		}
		if i > 0 && o <= last {
			t.Fatalf("samples need to be sorted, expected %d < %d", last, o)
		}
		last = o
	}
}
