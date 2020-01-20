/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Sub-command to restore data
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

package cmd

import (
	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/btrfs/index"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
)

type recoverFilesOptions struct {
	clobber bool
}

func init() {
	options := recoverFilesOptions{}
	// TODO(cblichmann): This command is not implemented
	recoverCmd := &cobra.Command{
		Use:   "recover",
		Short: "try to restore files from a damaged filesystem",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(btrfscue.Options.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doRecoverFiles(btrfscue.Options.Metadata)
		},
	}

	fs := recoverCmd.PersistentFlags()
	fs.BoolVar(&options.clobber, "clobber", false,
		"overwrite existing files")

	rootCmd.AddCommand(recoverCmd)
}

func doRecoverFiles(metadata string) {
	ix, err := index.OpenReadOnly(metadata)
	cliutil.ReportError(err)
	defer ix.Close()

	for r, v := ix.Subvolumes(); r.HasNext(); v = r.Next() {
		//cliutil.Verbosef("ID %d gen %d cgen %d top level - parent_uuid %s received_uuid %s uuid %s\n",
		//	r.Key().ObjectID, v.OTransID(), v.Generation(), v.ParentUUID(), v.ReceivedUUID(), v.UUID())
		cliutil.Verbosef("	item %s itemoff %d itemsize %d\n", r.Key(), r.Item().Offset(), r.Item().Size())
		cliutil.Verbosef("		root data bytenr %d level %d dirid %d refs %d gen %d lastsnap %d\n",
			v.ByteNr(), v.Level(), v.RootDirID(), v.Refs(), v.Generation(), v.LastSnapshot())
	}

	return
	//inode := uint64(264) // src.zip
	//ii := ix.InodeItem(ix.FindInodeItem(inode))
	//cliutil.Verbosef("file size: %d\n", ii.Size)

	//lo := uint64(0)
	//for i, end := ix.Range(
	//	btrfs.KF(btrfs.ExtentDataKey, inode),
	//	btrfs.KL(btrfs.ExtentDataKey, inode)); i < end; i++ {
	//	fe := &btrfs.FileExtentItem{} //ix.Item(i).Data.(*btrfs.FileExtentItem)
	//	lo = fe.DiskByteNr
	//	cliutil.Verbosef("file extent: %s %d %d %d %d\n", ix.Key(i),
	//		lo, fe.DiskNumBytes, fe.NumBytes, fe.Offset)
	//	_, of := ix.Physical(lo)
	//	cliutil.Verbosef("%d\n", of)
	//}

	//for i, end := ix.Range(
	//	btrfs.KF(btrfs.ChunkItemKey),
	//	btrfs.KL(btrfs.ChunkItemKey)); i < end; i++ {
	//	c := ix.Chunk(i)
	//	cliutil.Verbosef("%s %d 0x%x\n", ix.Key(i), c.Length, c.Stripes[0].Offset)
	//}
}
