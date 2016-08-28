// +build linux darwin

/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Sub-command to provide and mount a "rescue fs"
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

package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/subcommand"
)

type btrfscueFS struct {
	metadata  string
	index     *btrfs.Index
	rootDirID uint64
	device    *os.File

	root *basicNode
}

func newBtrfscueFSRoot(metadata string, index *btrfs.Index, rootDirID uint64,
	device *os.File) nodefs.Node {
	b := &btrfscueFS{
		metadata:  metadata,
		index:     index,
		rootDirID: rootDirID,
		device:    device,
	}
	b.root = b.newNode()
	return b.root
}

func (b *btrfscueFS) newNode() *basicNode {
	n := &basicNode{
		Node: nodefs.NewDefaultNode(),
		fs:   b,
	}
	now := time.Now()
	n.info.SetTimes(&now, &now, &now)
	n.info.Mode = 0555 | fuse.S_IFDIR
	return n
}

func (b *btrfscueFS) OnMount(c *nodefs.FileSystemConnector) {
	i := b.root.Inode()
	i.NewChild("rescue", true, newRescueNode(b.index, b.rootDirID))
	//i.NewChild("undelete", true, b.newNode())
	i.NewChild("metadata", false, newMetadataNode(b))
}

type basicNode struct {
	nodefs.Node
	fs   *btrfscueFS
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

func newMetadataNode(fs *btrfscueFS) *metadataNode {
	n := &metadataNode{}
	n.Node = nodefs.NewDefaultNode()
	n.fs = fs

	fi := &syscall.Stat_t{}
	syscall.Stat(fs.metadata, fi)

	n.info.FromStat(fi)
	n.info.Mode = n.info.Mode&0555 | fuse.S_IFREG
	return n
}

func (n *metadataNode) Open(flags uint32, context *fuse.Context) (
	nodefs.File, fuse.Status) {
	f, err := os.OpenFile(n.fs.metadata, int(flags), 0555)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return nodefs.NewReadOnlyFile(nodefs.NewLoopbackFile(f)),
		fuse.OK
}

type rescueNode struct {
	nodefs.Node
	index      *btrfs.Index
	indexInode uint64
	dirItems   map[string]*btrfs.DirItem
}

func newRescueNode(index *btrfs.Index, indexInode uint64) *rescueNode {
	return &rescueNode{
		Node:       nodefs.NewDefaultNode(),
		index:      index,
		indexInode: indexInode,
		dirItems:   make(map[string]*btrfs.DirItem),
	}
}

func (n *rescueNode) GetAttr(fi *fuse.Attr, file nodefs.File,
	context *fuse.Context) (code fuse.Status) {
	i := n.index.FindInodeItem(n.indexInode)
	if i == n.index.Len() {
		return fuse.ENOATTR
	}
	ino := n.index.InodeItem(i)

	*fi = fuse.Attr{
		Ino:       n.indexInode,
		Size:      ino.Size,
		Atime:     uint64(ino.Atime.Unix()),
		Mtime:     uint64(ino.Mtime.Unix()),
		Ctime:     uint64(ino.Ctime.Unix()),
		Atimensec: uint32(ino.Atime.Nanosecond()),
		Mtimensec: uint32(ino.Mtime.Nanosecond()),
		Ctimensec: uint32(ino.Ctime.Nanosecond()),
		Mode:      ino.Mode,
		Nlink:     ino.Nlink,
		Owner:     fuse.Owner{Uid: ino.UID, Gid: ino.GID},
		Rdev:      uint32(ino.Rdev),
	}
	return fuse.OK
}

func (n *rescueNode) ensureDirItems() {
	if len(n.dirItems) > 0 {
		return
	}
	for i, end := n.index.Range(btrfs.KeyFirst(btrfs.DirIndexKey, n.indexInode),
		btrfs.KeyLast(btrfs.DirIndexKey, n.indexInode)); i < end; i++ {
		di := n.index.DirItem(i)
		de := fuse.DirEntry{Name: di.Name}
		if di.Type == btrfs.FtDir {
			de.Mode = fuse.S_IFDIR
		} else {
			de.Mode = fuse.S_IFREG
		}
		n.dirItems[di.Name] = di
	}
}

func (n *rescueNode) Lookup(out *fuse.Attr, name string,
	context *fuse.Context) (*nodefs.Inode, fuse.Status) {
	n.ensureDirItems()
	di, ok := n.dirItems[name]
	if !ok {
		return nil, fuse.ENOENT
	}
	ch := n.Inode().NewChild(name, di.IsDir(), newRescueNode(n.index,
		di.Location.ObjectID))
	return ch, ch.Node().GetAttr(out, nil, context)
}

func (n *rescueNode) OpenDir(context *fuse.Context) ([]fuse.DirEntry,
	fuse.Status) {
	n.ensureDirItems()
	var s []fuse.DirEntry
	for _, di := range n.dirItems {
		entry := fuse.DirEntry{Name: di.Name, Mode: fuse.S_IFREG}
		if di.IsDir() {
			entry.Mode = fuse.S_IFDIR
		}
		s = append(s, entry)
	}
	return s, fuse.OK
}

func (n *rescueNode) Open(flags uint32, context *fuse.Context) (
	nodefs.File, fuse.Status) {

	for i, end := n.index.Range(
		btrfs.KeyFirst(btrfs.ExtentDataKey, n.indexInode),
		btrfs.KeyLast(btrfs.ExtentDataKey, n.indexInode)); i < end; i++ {
		fe := n.index.Item(i).Data.(*btrfs.FileExtentItem)
		fmt.Printf("%s (%d %d) (%d %d)\n", n.index.Key(i), fe.DiskByteNr,
			fe.DiskNumBytes, fe.Offset, fe.NumBytes)
		if fe.Type == btrfs.FileExtentInline {
			return nodefs.NewReadOnlyFile(nodefs.NewDataFile(
				[]byte(fe.Data))), fuse.OK
		}
	}
	return nil, fuse.ENOENT
}

func (n *rescueNode) GetXAttr(attribute string, context *fuse.Context) (
	data []byte, code fuse.Status) {
	// TODO(cblichmann): This should use btrfs.NameHash() for lookup
	for i, end := n.index.Range(
		btrfs.KeyFirst(btrfs.XAttrItemKey, n.indexInode),
		btrfs.KeyLast(btrfs.XAttrItemKey, n.indexInode)); i < end; i++ {
		di := n.index.DirItem(i)
		if attribute == di.Name {
			return []byte(di.Data), fuse.OK
		}
	}
	return nil, fuse.ENOENT
}

func (n *rescueNode) ListXAttr(context *fuse.Context) ([]string, fuse.Status) {
	attrs := []string{}
	for i, end := n.index.Range(
		btrfs.KeyFirst(btrfs.XAttrItemKey, n.indexInode),
		btrfs.KeyLast(btrfs.XAttrItemKey, n.indexInode)); i < end; i++ {
		attrs = append(attrs, n.index.DirItem(i).Name)
	}
	return attrs, fuse.OK
}

type mountCommand struct {
}

func (c *mountCommand) DefineFlags(fs *flag.FlagSet) {
}

func (c *mountCommand) Run(args []string) {
	if len(args) == 0 {
		fatalf("missing mount point\n")
	}
	if len(args) > 1 {
		fatalf("extra operand '%s'\n", args[1])
	}
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}

	m, err := os.Open(*metadata)
	reportError(err)
	defer m.Close()

	index := btrfs.NewIndex()
	reportError(ReadIndex(m, &index))

	dirID := uint64(btrfs.FirstFreeObjectId)

	root := newBtrfscueFSRoot(*metadata, &index, dirID, nil)

	server, _, err := nodefs.MountRoot(args[0], root, &nodefs.Options{})
	reportError(err)
	// TODO(cblichmann): Daemonize all the things!
	server.Serve()
}

func init() {
	subcommand.Register("mount",
		"provide a 'rescue' filesystem backed by metadata", &mountCommand{})
}
