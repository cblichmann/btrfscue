/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Sub-command to list files, directories and subvolumes/snapshots.
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
	"io"
	"os"
	"path"
	"sort"
	"text/tabwriter"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/subcommand"
)

func dirItemTypeString(t uint8) string {
	switch t {
	case btrfs.FtRegFile:
		return "-"
	case btrfs.FtBlkdev:
		return "b"
	case btrfs.FtChrdev:
		return "c"
	case btrfs.FtDir:
		return "d"
	case btrfs.FtSymlink:
		return "l"
	case btrfs.FtFifo:
		return "p"
	case btrfs.FtSock:
		return "s"
	default:
		return "?"
	}
}

func inodeModeString(mode uint32) string {
	const FilePerms string = "" +
		"---" + // 0
		"--x" + // 1
		"-w-" + // 2
		"-wx" + // 3
		"r--" + // 4
		"r-x" + // 5
		"rw-" + // 6
		"rwx" //   7
	u, g, o := 3*((mode>>6)&0x7), 3*((mode>>3)&0x7), 3*(mode&0x7)
	return FilePerms[u:u+3] + FilePerms[g:g+3] + FilePerms[o:o+3]
}

func listDirItem(w io.Writer, fs *btrfs.Index, di *btrfs.DirItem, showInode bool) {
	inode := di.Location.ObjectID
	if showInode {
		fmt.Fprintf(w, "%d\t", di.Location.ObjectID)
	}
	fmt.Fprintf(w, dirItemTypeString(di.Type))
	if i := fs.FindInodeItem(inode); i >= fs.Len() {
		fmt.Fprintf(w, "?????????\t?\t?\t?\t??????\t?????")
	} else {
		ii := fs.InodeItem(i)
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d", inodeModeString(ii.Mode),
			ii.Nlink, ii.UID, ii.GID, ii.Size)
	}

	fmt.Fprintf(w, "\t%s\n", di.Name)
}

type dirItemSlice []*btrfs.DirItem

func (a dirItemSlice) Len() int { return len(a) }
func (a dirItemSlice) Less(i, j int) bool {
	if !a[i].IsDir() && a[j].IsDir() {
		return true
	}
	if a[i].IsDir() && !a[j].IsDir() {
		return false
	}
	return a[i].Name < a[j].Name
}
func (a dirItemSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a dirItemSlice) Sort()         { sort.Sort(a) }

func listDirectory(w io.Writer, fs *btrfs.Index, dirID uint64, recursive,
	showInode bool) {
	dis := dirItemSlice{}
	for i, end := fs.Range(btrfs.KeyFirst(btrfs.DirIndexKey, dirID),
		btrfs.KeyLast(btrfs.DirIndexKey, dirID)); i < end; i++ {
		dis = append(dis, fs.DirItem(i))
	}
	sort.Sort(dis)
	todo := dirItemSlice{}
	for _, di := range dis {
		listDirItem(w, fs, di, showInode)
		if recursive && di.IsDir() {
			todo = append(todo, di)
		}
	}

	for _, di := range todo {
		fmt.Fprintf(w, "%s:\n", di.Name)
		listDirectory(w, fs, di.Location.ObjectID, true, showInode)
	}
}

type lsCommand struct {
	recursive bool
	inode     bool
}

func (c *lsCommand) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.recursive, "recursive", false,
		"recurse into sub-directories")
	fs.BoolVar(&c.inode, "inode", false,
		"show inode numbers")
}

func (c *lsCommand) Run(args []string) {
	if len(args) == 0 {
		args = append(args, "/")
	}
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}

	m, err := os.Open(*metadata)
	reportError(err)
	defer m.Close()

	fs := btrfs.NewIndex()
	reportError(ReadIndex(m, &fs))

	dirID := uint64(btrfs.FirstFreeObjectId)

	w := tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	var todo []uint64
	for _, p := range args {
		p = path.Clean(p)
		if p == "/" {
			todo = append(todo, dirID)
			continue
		}

		var i int
		if i = fs.FindDirItem(dirID, p); i >= fs.Len() {
			warnf("cannot lookup '%s': No such file or directory\n", p)
			continue
		}

		di := fs.DirItem(i)
		if di.IsDir() {
			listDirItem(w, &fs, di, c.inode)
			continue
		}

		todo = append(todo, di.Location.ObjectID)
	}
	w.Flush()

	w = tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	for _, i := range todo {
		listDirectory(w, &fs, i, c.recursive, c.inode)
	}
	w.Flush()
}

func init() {
	subcommand.Register("ls",
		"list information about files, directories and subvolumes/snapshots",
		&lsCommand{})
}
