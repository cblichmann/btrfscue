// +build linux darwin

// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Device extent backed file implementation

package rescuefs

import (
	"io"

	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
)

type extentMapEntry struct {
	fileOffset uint64
	length     uint64
	physOffset uint64
	diskByteNr uint64
	devID      uint64
}

type extentFile struct {
	nodefs.File
	fs        *rescueFS
	owner     uint64
	inode     uint64
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
		owner:     owner,
		inode:     id,
		extentMap: make([]extentMapEntry, 0, 1),
	}
	for ; r.HasNext(); e = r.Next() {
		var phys uint64
		var dev uint64
		if e.DiskByteNr() > 0 {
			dev, phys = fs.ix.Physical(e.DiskByteNr())
		}
		f.extentMap = append(f.extentMap, extentMapEntry{
			fileOffset: r.Key().Offset,
			length:     e.NumBytes(),
			physOffset: phys + e.Offset(),
			diskByteNr: e.DiskByteNr(),
			devID:      dev,
		})
	}
	return f
}

func (f *extentFile) String() string {
	return "extentFile"
}

func (f *extentFile) GetAttr(out *fuse.Attr) fuse.Status {
	ii := f.fs.ix.FindInodeItem(f.owner, f.inode)
	if ii != nil {
		out.Mode = ii.Mode()
		out.Size = ii.Size()
		out.Atime = uint64(ii.Atime().Unix())
		out.Mtime = uint64(ii.Mtime().Unix())
		out.Ctime = uint64(ii.Ctime().Unix())
		out.Atimensec = uint32(ii.Atime().Nanosecond())
		out.Mtimensec = uint32(ii.Mtime().Nanosecond())
		out.Ctimensec = uint32(ii.Ctime().Nanosecond())
		out.Nlink = ii.Nlink()
		out.Owner = fuse.Owner{Uid: ii.UID(), Gid: ii.GID()}
		out.Rdev = uint32(ii.Rdev())
		return fuse.OK
	}

	// Fallback if inode item is not found in metadata
	out.Mode = fuse.S_IFREG | 0644
	var size uint64
	if len(f.extentMap) > 0 {
		last := f.extentMap[len(f.extentMap)-1]
		size = last.fileOffset + last.length
	}
	out.Size = size
	return fuse.OK
}

func (f *extentFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	if off < 0 {
		return nil, fuse.EINVAL
	}

	bytesRead := 0
	bufLen := len(buf)

	for _, entry := range f.extentMap {
		if off >= int64(entry.fileOffset+entry.length) {
			continue
		}
		if off+int64(bufLen-bytesRead) <= int64(entry.fileOffset) {
			break
		}

		// Calculate overlap between the requested range and the current extent
		extentStart := int64(entry.fileOffset)
		extentEnd := int64(entry.fileOffset + entry.length)

		readStart := off
		if readStart < extentStart {
			readStart = extentStart
		}

		readEnd := off + int64(bufLen-bytesRead)
		if readEnd > extentEnd {
			readEnd = extentEnd
		}

		if readEnd <= readStart {
			continue
		}

		chunkSize := readEnd - readStart
		destOffset := readStart - off

		if entry.diskByteNr == 0 {
			// Sparse hole: fill with zeros
			for i := int64(0); i < chunkSize; i++ {
				buf[destOffset+i] = 0
			}
		} else {
			// Read from physical device
			physReadOffset := int64(entry.physOffset) + (readStart - extentStart)
			if f.fs.dev == nil {
				// No device file provided: return zeroes
				for i := int64(0); i < chunkSize; i++ {
					buf[destOffset+i] = 0
				}
			} else {
				if _, err := f.fs.dev.ReadAt(buf[destOffset:destOffset+chunkSize], physReadOffset); err != nil && err != io.EOF {
					cliutil.Warnf("read error at physical offset %d: %v\n", physReadOffset, err)
					return nil, fuse.EIO
				}
			}
		}
		bytesRead += int(chunkSize)
	}

	if bytesRead < bufLen {
		return fuse.ReadResultData(buf[:bytesRead]), fuse.OK
	}
	return fuse.ReadResultData(buf), fuse.OK
}
