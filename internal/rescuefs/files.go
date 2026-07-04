// +build linux darwin

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Device extent backed file implementation

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
