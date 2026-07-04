// +build !linux,!darwin

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Null implementation for non-Linux, non-Darwin systems

package rescuefs

import (
	"errors"
	"io"

	"blichmann.eu/code/btrfscue/pkg/btrfs"
)

type rescueFS struct{}

func New(metadata string, ix *btrfs.Index, reader io.ReaderAt) rescueFS {
	// Do nothing
	return rescueFS{}
}

var errNotSupported = errors.New(
	"FUSE mount is only supported on Linux and macOS")

func (r *rescueFS) Mount(on string) error { return errNotSupported }
func (r *rescueFS) Unmount() error        { return nil }
func (r *rescueFS) Serve() error          { return errNotSupported }
