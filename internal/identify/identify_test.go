// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Tests for the identify sub-command

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
