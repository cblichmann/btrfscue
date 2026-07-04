// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Global flags applicable to all/most commands

package btrfs

import (
	"fmt"
	"io"
)

func CheckDeviceSize(rs io.ReadSeeker, blockSize uint64) (uint64, error) {
	size, err := rs.Seek(0, io.SeekEnd)
	if err == nil {
		if uint64(size) < blockSize {
			err = fmt.Errorf("device smaller than block size: %d < %d", size,
				blockSize)
		} else if uint64(size) < SuperInfoOffset2+blockSize*100 {
			// Sanity check: BTRFS minimum filesystem size is 64MiB plus a few
			// blocks
			err = fmt.Errorf("device too small, must be > 64MiB: %d", size)
		}
	}
	return uint64(size), err
}
