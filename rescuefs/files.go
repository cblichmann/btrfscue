// +build linux darwin

/*
 * btrfscue version 0.3
 * Copyright (c)2011-2017 Christian Blichmann
 *
 * Device extent backed file implementation
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

package rescuefs // import "blichmann.eu/code/btrfscue/rescuefs"

import (
	"fmt" // DBG!!!
	"github.com/hanwen/go-fuse/fuse/nodefs"

	_ "blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfs/index"
)

type extentFile struct {
	nodefs.File
}

func newExtentFile(ix *index.Index, owner, id uint64) nodefs.File {
	r, e := ix.FileExtentItems(owner, id)
	if !r.HasNext() {
		return nil
	}
	f := &extentFile{File: nodefs.NewReadOnlyFile(nodefs.NewDefaultFile())}
	for ; r.HasNext(); e = r.Next() {
		fmt.Printf("%s (%d %d) (%d %d)\n", r.Key(), e.DiskByteNr(),
			e.DiskNumBytes(), e.Offset(), e.NumBytes())
	}
	return f
}
