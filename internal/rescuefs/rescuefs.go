// +build linux darwin

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Rescue FS API

package rescuefs

import (
	"io"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"

	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
)

type rescueFS struct {
	metadata string
	ix       *index.Index

	dev io.ReaderAt

	root   *basicNode
	server *fuse.Server
}

func New(metadata string, ix *index.Index, dev io.ReaderAt) rescueFS {
	r := rescueFS{metadata: metadata, ix: ix, dev: dev}
	r.root = r.newNode()
	return r
}

func (r *rescueFS) Mount(on string) error {
	var err error
	r.server, _, err = nodefs.MountRoot(on, r.root, &nodefs.Options{})
	return err
}

func (r *rescueFS) Unmount() error { return r.server.Unmount() }
func (r *rescueFS) Serve()         { r.server.Serve() }
