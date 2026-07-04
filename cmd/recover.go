// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Sub-command to restore data

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"blichmann.eu/code/btrfscue/cmd/btrfscue/app"
	cliutil "blichmann.eu/code/btrfscue/cmd/btrfscue/app/util"
	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
)

type recoverFilesOptions struct {
	clobber bool
}

func init() {
	options := recoverFilesOptions{}
	recoverCmd := &cobra.Command{
		Use:   "recover DEV/IMAGE DESTDIR",
		Short: "try to restore files from a damaged filesystem",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(app.Global.Metadata) == 0 {
				cliutil.Fatalf("missing metadata option\n")
			}
			doRecoverFiles(args[0], args[1], app.Global.Metadata, options)
		},
	}

	fs := recoverCmd.PersistentFlags()
	fs.BoolVar(&options.clobber, "clobber", false,
		"overwrite existing files")

	rootCmd.AddCommand(recoverCmd)
}

func doRecoverFiles(imagePath, destDir, metadata string, options recoverFilesOptions) {
	ix, err := index.OpenReadOnly(metadata)
	cliutil.ReportError(err)
	defer ix.Close()

	f, err := os.Open(imagePath)
	cliutil.ReportError(err)
	defer f.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		cliutil.Fatalf("failed to create destination directory: %v\n", err)
	}

	visited := make(map[[2]uint64]bool)

	// Start recovering from the main FS Tree root directory
	owner := uint64(btrfs.FSTreeObjectID)
	dirID := uint64(btrfs.FirstFreeObjectID)

	cliutil.Verbosef("Recovering root directory tree from subvolume %d...\n", owner)
	if err := recoverDir(ix, f, owner, dirID, destDir, options, visited); err != nil {
		cliutil.Warnf("failed to recover root directory: %v\n", err)
	}

		// Also find and recover any other subvolumes that weren't visited
	for r, _ := ix.Subvolumes(); r.HasNext(); _ = r.Next() {
		subOwner := r.Key().ObjectID
		key := [2]uint64{subOwner, btrfs.FirstFreeObjectID}
		if !visited[key] {
			subvolName := fmt.Sprintf("subvol_%d", subOwner)
			subvolDest := filepath.Join(destDir, subvolName)
			cliutil.Verbosef("Recovering unreferenced subvolume %d to %s...\n", subOwner, subvolDest)
			if err := recoverDir(ix, f, subOwner, btrfs.FirstFreeObjectID, subvolDest, options, visited); err != nil {
				cliutil.Warnf("failed to recover subvolume %d: %v\n", subOwner, err)
			}
		}
	}
}

func recoverDir(ix *index.Index, devFile *os.File, owner, dirID uint64, currentDest string, options recoverFilesOptions, visited map[[2]uint64]bool) error {
	key := [2]uint64{owner, dirID}
	if visited[key] {
		return nil
	}
	visited[key] = true

	if err := os.MkdirAll(currentDest, 0755); err != nil {
		return err
	}

	dis := []btrfs.DirItem{}
	for r, v := ix.DirItems(owner, dirID); r.HasNext(); v = r.Next() {
		dis = append(dis, v)
	}

	for _, di := range dis {
		name := di.Name()
		if name == "." || name == ".." || strings.Contains(name, "/") || strings.Contains(name, "\\") {
			continue
		}

		targetPath := filepath.Join(currentDest, name)

		if di.IsDir() || di.IsSubvolume() {
			subOwner := owner
			subDirID := di.Location().ObjectID
			if di.IsSubvolume() {
				subOwner = di.Location().ObjectID
				subDirID = btrfs.FirstFreeObjectID
			}
			if err := recoverDir(ix, devFile, subOwner, subDirID, targetPath, options, visited); err != nil {
				cliutil.Warnf("failed to recover directory %s: %v\n", targetPath, err)
			}
		} else if di.Type() == btrfs.FtRegFile {
			if err := recoverFile(ix, devFile, owner, di.Location().ObjectID, targetPath, options); err != nil {
				cliutil.Warnf("failed to recover file %s: %v\n", targetPath, err)
			}
		} else if di.Type() == btrfs.FtSymlink {
			if err := recoverSymlink(ix, devFile, owner, di.Location().ObjectID, targetPath, options); err != nil {
				cliutil.Warnf("failed to recover symlink %s: %v\n", targetPath, err)
			}
		} else {
			cliutil.Verbosef("skipping special file %s of type %d\n", targetPath, di.Type())
		}
	}
	return nil
}

func recoverFile(ix *index.Index, devFile *os.File, owner, inode uint64, targetPath string, options recoverFilesOptions) error {
	if _, err := os.Stat(targetPath); err == nil && !options.clobber {
		cliutil.Verbosef("file %s already exists, skipping (--clobber not specified)\n", targetPath)
		return nil
	}

	ii := ix.FindInodeItem(owner, inode)
	var fileSize uint64
	if ii != nil {
		fileSize = ii.Size()
	}

	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	cliutil.Verbosef("recovering file: %s (size: %d)\n", targetPath, fileSize)

	r, e := ix.FileExtentItems(owner, inode)
	if !r.HasNext() {
		if ii != nil {
			_ = f.Truncate(int64(fileSize))
		}
		return nil
	}

	for ; r.HasNext(); e = r.Next() {
		fileOffset := r.Key().Offset

		if e.IsInline() {
			inlineData := []byte(e.Data())
			limit := len(inlineData)
			if ii != nil && int64(fileOffset)+int64(limit) > int64(fileSize) {
				limit = int(fileSize - fileOffset)
			}
			if limit > 0 {
				if _, err := f.WriteAt(inlineData[:limit], int64(fileOffset)); err != nil {
					return fmt.Errorf("write inline data failed: %w", err)
				}
			}
		} else {
			numBytes := e.NumBytes()
			if e.DiskByteNr() == 0 {
				zeros := make([]byte, numBytes)
				if _, err := f.WriteAt(zeros, int64(fileOffset)); err != nil {
					return fmt.Errorf("write zeros failed: %w", err)
				}
			} else {
				_, physOffset := ix.Physical(e.DiskByteNr())
				srcOffset := physOffset + e.Offset()

				const chunkSize = 1024 * 1024
				buf := make([]byte, chunkSize)

				var bytesCopied uint64 = 0
				for bytesCopied < numBytes {
					toCopy := numBytes - bytesCopied
					if toCopy > chunkSize {
						toCopy = chunkSize
					}

					n, err := devFile.ReadAt(buf[:toCopy], int64(srcOffset+bytesCopied))
					if n > 0 {
						if _, wErr := f.WriteAt(buf[:n], int64(fileOffset+bytesCopied)); wErr != nil {
							return fmt.Errorf("write file data failed: %w", wErr)
						}
						bytesCopied += uint64(n)
					}
					if err != nil {
						if err == io.EOF && bytesCopied < numBytes {
							return fmt.Errorf("unexpected EOF reading device/image: %w", err)
						}
						break
					}
				}
			}
		}
	}

	if ii != nil {
		if err := f.Truncate(int64(fileSize)); err != nil {
			return fmt.Errorf("truncate failed: %w", err)
		}
	}
	return nil
}

func recoverSymlink(ix *index.Index, devFile *os.File, owner, inode uint64, targetPath string, options recoverFilesOptions) error {
	if _, err := os.Lstat(targetPath); err == nil {
		if !options.clobber {
			cliutil.Verbosef("symlink %s already exists, skipping (--clobber not specified)\n", targetPath)
			return nil
		}
		_ = os.Remove(targetPath)
	}

	e := ix.FindFileExtentItem(owner, inode)
	if e == nil {
		return fmt.Errorf("symlink extent not found")
	}

	var target string
	if e.IsInline() {
		target = e.Data()
	} else {
		if e.DiskByteNr() > 0 && devFile != nil {
			_, physOffset := ix.Physical(e.DiskByteNr())
			srcOffset := physOffset + e.Offset()
			buf := make([]byte, e.NumBytes())
			if _, err := devFile.ReadAt(buf, int64(srcOffset)); err == nil {
				target = string(buf)
			} else {
				return fmt.Errorf("failed to read symlink target: %w", err)
			}
		} else {
			return fmt.Errorf("symlink target is not inline and device file is missing/invalid")
		}
	}

	cliutil.Verbosef("recovering symlink: %s -> %s\n", targetPath, target)
	return os.Symlink(target, targetPath)
}
