// +build linux darwin

/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Rescue FS API
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
