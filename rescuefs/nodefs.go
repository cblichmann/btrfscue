// +build linux darwin

/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * A "rescue" FS that provides a read-only view backed by index metadata
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
	"os"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"

	"fmt" //DBG!!!

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfs/index"
)

func init() { fmt.Printf("") } //DBG!!!

func (r *rescueFS) newNode() *basicNode {
	n := &basicNode{Node: nodefs.NewDefaultNode(), fs: r}
	now := time.Now()
	n.info.SetTimes(&now, &now, &now)
	n.info.Mode = 0555 | fuse.S_IFDIR
	return n
}

func (r *rescueFS) OnMount(c *nodefs.FileSystemConnector) {
	i := r.root.Inode()
	i.NewChild("rescue", true, newRescueNode(r, btrfs.FSTreeObjectID,
		btrfs.FirstFreeObjectID))
	i.NewChild("metadata", false, newMetadataNode(r))
}

type basicNode struct {
	nodefs.Node
	fs   *rescueFS
	info fuse.Attr
}

func (n *basicNode) GetAttr(fi *fuse.Attr, file nodefs.File,
	context *fuse.Context) (code fuse.Status) {
	*fi = n.info
	return fuse.OK
}

func (n *basicNode) OnMount(c *nodefs.FileSystemConnector) {
	n.fs.OnMount(c)
}

func (n *basicNode) Deletable() bool { return false }

func (n *basicNode) StatFs() *fuse.StatfsOut { return &fuse.StatfsOut{} }

type metadataNode struct {
	basicNode
}

func newMetadataNode(fs *rescueFS) *metadataNode {
	n := &metadataNode{basicNode: *fs.newNode()}

	fi := &syscall.Stat_t{}
	syscall.Stat(fs.metadata, fi)

	n.info.FromStat(fi)
	n.info.Mode = n.info.Mode&0555 | fuse.S_IFREG
	return n
}

func (n *metadataNode) Open(flags uint32, context *fuse.Context) (
	nodefs.File, fuse.Status) {
	if f, err := os.OpenFile(n.fs.metadata, int(flags), 0555); err != nil {
		return nil, fuse.ToStatus(err)
	} else {
		return nodefs.NewReadOnlyFile(nodefs.NewLoopbackFile(f)), fuse.OK
	}
}

type rescueNode struct {
	nodefs.Node
	fs       *rescueFS
	ix       *index.Index
	owner    uint64
	ixInode  uint64
	attr     *fuse.Attr // Cached attributes
	dirItems map[string]btrfs.DirItem
}

func newRescueNode(fs *rescueFS, owner, inode uint64) *rescueNode {
	return &rescueNode{
		Node:     nodefs.NewDefaultNode(),
		fs:       fs,
		ix:       fs.ix,
		owner:    owner,
		ixInode:  inode,
		dirItems: make(map[string]btrfs.DirItem),
	}
}

func (n *rescueNode) GetAttr(fi *fuse.Attr, file nodefs.File,
	context *fuse.Context) (code fuse.Status) {
	if n.attr == nil {
		ii := n.ix.FindInodeItem(n.owner, n.ixInode)
		if ii == nil {
			fmt.Println("No data for", n.owner, n.ixInode) //DBG!!!
			return fuse.ENOATTR
		}
		n.attr = &fuse.Attr{
			Ino:       n.ixInode,
			Size:      ii.Size(),
			Atime:     uint64(ii.Atime().Unix()),
			Mtime:     uint64(ii.Mtime().Unix()),
			Ctime:     uint64(ii.Ctime().Unix()),
			Atimensec: uint32(ii.Atime().Nanosecond()),
			Mtimensec: uint32(ii.Mtime().Nanosecond()),
			Ctimensec: uint32(ii.Ctime().Nanosecond()),
			Mode:      ii.Mode(),
			Nlink:     ii.Nlink(),
			Owner:     fuse.Owner{Uid: ii.UID(), Gid: ii.GID()},
			Rdev:      uint32(ii.Rdev()),
		}
	}
	*fi = *n.attr
	return fuse.OK
}

func (n *rescueNode) ensureDirItems() {
	if len(n.dirItems) > 0 {
		return
	}
	for r, d := n.ix.DirItems(n.owner, n.ixInode); r.HasNext(); d = r.Next() {
		n.dirItems[d.Name()] = d
	}
}

func (n *rescueNode) Lookup(out *fuse.Attr, name string,
	context *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.ensureDirItems()
	d, ok := n.dirItems[name]
	if !ok {
		return nil, fuse.ENOENT
	}
	var owner, inode uint64
	if !d.IsSubvolume() {
		owner = n.owner
		inode = d.Location().ObjectID
	} else {
		owner = d.Location().ObjectID
		inode = btrfs.FirstFreeObjectID
	}
	node := newRescueNode(n.fs, owner, inode)
	ch := n.Inode().NewChild(name, d.IsDir(), node)
	return ch, node.GetAttr(out, nil, context)
}

func (n *rescueNode) OpenDir(context *fuse.Context) ([]fuse.DirEntry,
	fuse.Status) {
	n.ensureDirItems()
	var s []fuse.DirEntry
	for _, di := range n.dirItems {
		entry := fuse.DirEntry{Name: di.Name(), Mode: fuse.S_IFREG}
		if di.IsDir() {
			entry.Mode = fuse.S_IFDIR
		}
		s = append(s, entry)
	}
	return s, fuse.OK
}

func (n *rescueNode) Open(flags uint32, context *fuse.Context) (
	nodefs.File, fuse.Status) {
	// TODO(cblichmann): HACK HACK HACK!
	r, e := n.ix.FileExtentItems(n.owner, n.ixInode)
	if !r.HasNext() {
		return nil, fuse.ENOENT
	}
	if e.IsInline() {
		return nodefs.NewReadOnlyFile(nodefs.NewDataFile(
			[]byte(e.Data()))), fuse.OK
	}

	if f := newExtentFile(n.fs, n.owner, n.ixInode); f != nil {
		return f, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (n *rescueNode) GetXAttr(attribute string, context *fuse.Context) (
	data []byte, code fuse.Status) {
	// TODO(cblichmann): This should use btrfs.NameHash() for lookup
	for r, x := n.ix.XAttrItems(n.owner, n.ixInode); r.HasNext(); x = r.Next() {
		if x.Name() == attribute {
			return []byte(x.Data()), fuse.OK
		}
	}
	return nil, fuse.ENOATTR
}

func (n *rescueNode) ListXAttr(context *fuse.Context) ([]string, fuse.Status) {
	attrs := []string{}
	for r, x := n.ix.XAttrItems(n.owner, n.ixInode); r.HasNext(); x = r.Next() {
		attrs = append(attrs, x.Name())
	}
	return attrs, fuse.OK
}

func (n *rescueNode) Readlink(c *fuse.Context) ([]byte, fuse.Status) {
	// Link data is stored in inline extent.
	if e := n.ix.FindFileExtentItem(n.owner, n.ixInode); e != nil &&
		e.IsInline() {
		return []byte(e.Data()), fuse.OK
	}
	return nil, fuse.ENODATA
}
