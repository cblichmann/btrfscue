// +build linux darwin

/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
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

package rescuefs

import (
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
)

// Similar to stripeMapEntry in Index.Physical()
type extentMapEntry struct {
	devID  uint64 // Not yet used
	offset uint64
	length uint64
	// TODO(cblichmann): Compression
}

type extentFile struct {
	nodefs.File
	fs        *rescueFS
	extentMap []extentMapEntry
}

func newExtentFile(fs *rescueFS, owner, id uint64) nodefs.File {
	r, e := fs.ix.FileExtentItems(owner, id)
	if !r.HasNext() {
		return nil
	}
	f := &extentFile{
		File:      nodefs.NewReadOnlyFile(nodefs.NewDefaultFile()),
		fs:        fs,
		extentMap: make([]extentMapEntry, 0, 1)}
	for ; r.HasNext(); e = r.Next() {
		dev, phys := fs.ix.Physical(e.DiskByteNr())
		f.extentMap = append(f.extentMap, extentMapEntry{
			devID:  dev,
			offset: phys,
			length: e.DiskNumBytes()})
	}
	return f
}

func (f *extentFile) String() string {
	return "extentFile"
}

func (f *extentFile) GetAttr(out *fuse.Attr) fuse.Status {
	// TODO(cblichmann): Return the real attributes, esp. size
	out.Mode = fuse.S_IFREG | 0644
	out.Size = f.extentMap[0].length
	return fuse.OK
}

func (f *extentFile) Read(buf []byte, off int64) (res fuse.ReadResult,
	code fuse.Status) {
	// TODO(cblichmann): Read more than the first extent
	start := int64(f.extentMap[0].offset) + off
	if uint64(off+int64(len(buf))) <= f.extentMap[0].length {
		if _, err := f.fs.dev.ReadAt(buf, start); err != nil {
			cliutil.Warnf("read error: \n", err)
		}
	} else {
		cliutil.Warnf("read out of bounds: only first extent supported (up "+
			"to %d bytes)\n", f.extentMap[0].length)
	}
	return fuse.ReadResultData(buf), fuse.OK
}
