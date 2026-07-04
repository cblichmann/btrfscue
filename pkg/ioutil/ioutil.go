// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// I/O utility routines

package ioutil

import (
	"io"
)

// ReadBlockAt reads a block of data from an io.ReaderAt.
//
// It always either reads a full block or reports an error.
func ReadBlockAt(r io.ReaderAt, block []byte, offset uint64) error {
	if read, err := r.ReadAt(block, int64(offset)); err != nil {
		return err
	} else if read == len(block) {
		return nil // Never return EOF if read succeeded
	}
	return nil
}
