/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
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

package cmd // import "blichmann.eu/code/btrfscue/cmd"

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
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

func shortTime(t time.Time) string {
	return t.Format("Jan _2 15:04")
}

func listDirItem(w io.Writer, ix *index.Index, owner uint64, di btrfs.DirItem,
	showInode bool) {
	inode := di.Location().ObjectID
	if showInode {
		fmt.Fprintf(w, "%d\t", di.Location().ObjectID)
	}
	fmt.Fprintf(w, dirItemTypeString(di.Type()))
	if ii := ix.FindInodeItem(owner, inode); ii == nil {
		fmt.Fprintf(w, "?????????\t?\t?\t?\t?\t?")
	} else {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%s", inodeModeString(ii.Mode()),
			ii.Nlink(), ii.UID(), ii.GID(), ii.Size(), shortTime(ii.Ctime()))
	}

	fmt.Fprintf(w, "\t%s\n", di.Name())
	// TODO(cblichmann): Print symlinks
}

type dirItemSlice []btrfs.DirItem

func (a dirItemSlice) Len() int { return len(a) }
func (a dirItemSlice) Less(i, j int) bool {
	if !a[i].IsDir() && a[j].IsDir() {
		return true
	}
	if a[i].IsDir() && !a[j].IsDir() {
		return false
	}
	return a[i].Name() < a[j].Name()
}
func (a dirItemSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a dirItemSlice) Sort()         { sort.Sort(a) }

func listDirectory(w io.Writer, ix *index.Index, owner, dirID uint64,
	recursive, showInode bool) {
	dis := dirItemSlice{}
	for r, v := ix.DirItems(owner, dirID); r.HasNext(); v = r.Next() {
		dis = append(dis, v)
	}
	sort.Sort(dis)
	todo := dirItemSlice{}
	for _, di := range dis {
		listDirItem(w, ix, owner, di, showInode)
		// TODO(cblichmann): Subvolumes/Snapshots, add owner
		if recursive && di.IsDir() {
			todo = append(todo, di)
		}
	}

	for _, di := range todo {
		fmt.Fprintf(w, "%s:\n", di.Name())
		listDirectory(w, ix, owner, di.Location().ObjectID, true, showInode)
	}
}

type listFilesOptions struct {
	Recursive bool
	Inode     bool
}

func init() {
	options := listFilesOptions{}
	lsCmd := &cobra.Command{
		Use: "ls",
		Short: "list information about files, directories and " +
			"subvolumes/snapshots",
		Run: func(cmd *cobra.Command, args []string) {
			if len(btrfscue.Options.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doListFiles(args, btrfscue.Options.Metadata, options)
		},
	}

	fs := lsCmd.PersistentFlags()
	fs.BoolVar(&options.Recursive, "recursive", false,
		"recurse into sub-directories")
	fs.BoolVar(&options.Inode, "inode", false,
		"show inode numbers")

	rootCmd.AddCommand(lsCmd)
}

func doListFiles(args []string, metadata string, options listFilesOptions) {
	if len(args) == 0 {
		args = append(args, "/")
	}

	ix, err := index.OpenReadOnly(btrfscue.Options.Metadata)
	cliutil.ReportError(err)
	defer ix.Close()

	owner := uint64(btrfs.FSTreeObjectID)
	dirID := uint64(btrfs.FirstFreeObjectID)

	w := tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	type Todo struct{ owner, dirID uint64 }
	var todo []Todo
	for _, p := range args {
		p = path.Clean(p)
		if p == "/" {
			todo = append(todo, Todo{owner, dirID})
			continue
		}

		if di := ix.FindDirItemForPath(owner, p); di == nil {
			cliutil.Warnf("cannot lookup '%s': No such file or directory\n", p)
			continue
		} else if !di.IsDir() {
			listDirItem(w, ix, owner, di, options.Inode)
			continue
		} else {
			todo = append(todo, Todo{owner, di.Location().ObjectID})
		}
	}
	w.Flush()

	w = tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	for _, t := range todo {
		listDirectory(w, ix, t.owner, t.dirID, options.Recursive, options.Inode)
	}
	w.Flush()
}
