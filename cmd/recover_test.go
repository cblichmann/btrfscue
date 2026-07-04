// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

package cmd

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/btrfs/index"
	"blichmann.eu/code/btrfscue/pkg/uuid"
)

func makeHeader(owner, generation uint64, fsid uuid.UUID) btrfs.Header {
	h := make([]byte, btrfs.HeaderLen)
	copy(h[32:], fsid[:])
	binary.LittleEndian.PutUint64(h[80:], generation)
	binary.LittleEndian.PutUint64(h[88:], owner)
	return btrfs.Header(h)
}

func makeItem(k btrfs.Key, offset, size uint32) btrfs.Item {
	item := make([]byte, btrfs.ItemLen)
	binary.LittleEndian.PutUint64(item[0:], k.ObjectID)
	item[8] = k.Type
	binary.LittleEndian.PutUint64(item[9:], k.Offset)
	binary.LittleEndian.PutUint32(item[17:], offset)
	binary.LittleEndian.PutUint32(item[21:], size)
	return btrfs.Item(item)
}

func makeInodeItem(size uint64, mode uint32) []byte {
	ii := make([]byte, btrfs.InodeItemLen)
	binary.LittleEndian.PutUint64(ii[16:], size)
	binary.LittleEndian.PutUint32(ii[52:], mode)
	return ii
}

func makeDirItem(loc btrfs.Key, typeVal uint8, name string) []byte {
	nameBytes := []byte(name)
	di := make([]byte, 30+len(nameBytes))
	binary.LittleEndian.PutUint64(di[0:], loc.ObjectID)
	di[8] = loc.Type
	binary.LittleEndian.PutUint64(di[9:], loc.Offset)
	binary.LittleEndian.PutUint16(di[25:], 0)
	binary.LittleEndian.PutUint16(di[27:], uint16(len(nameBytes)))
	di[29] = typeVal
	copy(di[30:], nameBytes)
	return di
}

func makeInlineFileExtentItem(data []byte) []byte {
	fe := make([]byte, 21+len(data))
	binary.LittleEndian.PutUint64(fe[8:], uint64(len(data)))
	fe[20] = 0 // Type = FileExtentInline
	copy(fe[21:], data)
	return fe
}

func makeRegFileExtentItem(diskByteNr, diskNumBytes, offset, numBytes uint64) []byte {
	fe := make([]byte, 53)
	binary.LittleEndian.PutUint64(fe[8:], numBytes) // RAMBytes
	fe[20] = 1 // Type = FileExtentReg
	binary.LittleEndian.PutUint64(fe[21:], diskByteNr)
	binary.LittleEndian.PutUint64(fe[29:], diskNumBytes)
	binary.LittleEndian.PutUint64(fe[37:], offset)
	binary.LittleEndian.PutUint64(fe[45:], numBytes)
	return fe
}

func makeChunk(length uint64, devID uint64, offset uint64) []byte {
	c := make([]byte, 48+32)
	binary.LittleEndian.PutUint64(c[0:], length)
	binary.LittleEndian.PutUint16(c[44:], 1) // NumStripes = 1
	binary.LittleEndian.PutUint64(c[48:], devID)
	binary.LittleEndian.PutUint64(c[56:], offset)
	return c
}

func TestRecoverCommand(t *testing.T) {
	td, err := ioutil.TempDir("", "btrfscue_recover_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	metadataPath := filepath.Join(td, "metadata.db")
	imagePath := filepath.Join(td, "disk.img")
	destDir := filepath.Join(td, "recovered")

	fsid := uuid.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	ix, err := index.Open(metadataPath, 0644, &index.Options{
		BlockSize:  4096,
		FSID:       fsid,
		Generation: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	chunkKey := btrfs.Key{ObjectID: btrfs.FirstFreeObjectID, Type: btrfs.ChunkItemKey, Offset: 1000000}
	chunkData := makeChunk(4096, 1, 10000)
	hChunk := makeHeader(btrfs.ChunkTreeObjectID, 1, fsid)
	itemChunk := makeItem(chunkKey, 0, uint32(len(chunkData)))
	if err := ix.InsertItem(chunkKey, hChunk, itemChunk, chunkData); err != nil {
		t.Fatal(err)
	}

	fileInode := uint64(257)
	dirKey := btrfs.Key{ObjectID: btrfs.FirstFreeObjectID, Type: btrfs.DirItemKey, Offset: uint64(index.NameHash("test.txt"))}
	dirData := makeDirItem(btrfs.Key{ObjectID: fileInode, Type: btrfs.InodeItemKey, Offset: 0}, btrfs.FtRegFile, "test.txt")
	hDir := makeHeader(btrfs.FSTreeObjectID, 1, fsid)
	itemDir := makeItem(dirKey, 0, uint32(len(dirData)))
	if err := ix.InsertItem(dirKey, hDir, itemDir, dirData); err != nil {
		t.Fatal(err)
	}

	fileInodeKey := btrfs.Key{ObjectID: fileInode, Type: btrfs.InodeItemKey, Offset: 0}
	fileInodeData := makeInodeItem(12, 0644) // size = 12 bytes
	itemFileInode := makeItem(fileInodeKey, 0, uint32(len(fileInodeData)))
	if err := ix.InsertItem(fileInodeKey, hDir, itemFileInode, fileInodeData); err != nil {
		t.Fatal(err)
	}

	fileExtentKey := btrfs.Key{ObjectID: fileInode, Type: btrfs.ExtentDataKey, Offset: 0}
	fileExtentData := makeRegFileExtentItem(1000000, 4096, 0, 12)
	itemFileExtent := makeItem(fileExtentKey, 0, uint32(len(fileExtentData)))
	if err := ix.InsertItem(fileExtentKey, hDir, itemFileExtent, fileExtentData); err != nil {
		t.Fatal(err)
	}

	inlineInode := uint64(258)
	inlineDirKey := btrfs.Key{ObjectID: btrfs.FirstFreeObjectID, Type: btrfs.DirItemKey, Offset: uint64(index.NameHash("inline.txt"))}
	inlineDirData := makeDirItem(btrfs.Key{ObjectID: inlineInode, Type: btrfs.InodeItemKey, Offset: 0}, btrfs.FtRegFile, "inline.txt")
	itemInlineDir := makeItem(inlineDirKey, 0, uint32(len(inlineDirData)))
	if err := ix.InsertItem(inlineDirKey, hDir, itemInlineDir, inlineDirData); err != nil {
		t.Fatal(err)
	}

	inlineInodeKey := btrfs.Key{ObjectID: inlineInode, Type: btrfs.InodeItemKey, Offset: 0}
	inlineInodeData := makeInodeItem(5, 0644)
	itemInlineInode := makeItem(inlineInodeKey, 0, uint32(len(inlineInodeData)))
	if err := ix.InsertItem(inlineInodeKey, hDir, itemInlineInode, inlineInodeData); err != nil {
		t.Fatal(err)
	}

	inlineExtentKey := btrfs.Key{ObjectID: inlineInode, Type: btrfs.ExtentDataKey, Offset: 0}
	inlineExtentData := makeInlineFileExtentItem([]byte("hello"))
	itemInlineExtent := makeItem(inlineExtentKey, 0, uint32(len(inlineExtentData)))
	if err := ix.InsertItem(inlineExtentKey, hDir, itemInlineExtent, inlineExtentData); err != nil {
		t.Fatal(err)
	}

	if err := ix.Commit(); err != nil {
		t.Fatal(err)
	}
	ix.Close()

	imageContent := make([]byte, 20000)
	copy(imageContent[10000:], []byte("world wide!!"))
	if err := ioutil.WriteFile(imagePath, imageContent, 0644); err != nil {
		t.Fatal(err)
	}

	options := recoverFilesOptions{clobber: true}
	doRecoverFiles(imagePath, destDir, metadataPath, options)

	recoveredFile := filepath.Join(destDir, "test.txt")
	if data, err := ioutil.ReadFile(recoveredFile); err != nil {
		t.Fatalf("failed to read recovered file: %v", err)
	} else if string(data) != "world wide!!" {
		t.Errorf("expected 'world wide!!', got '%s'", string(data))
	}

	recoveredInline := filepath.Join(destDir, "inline.txt")
	if data, err := ioutil.ReadFile(recoveredInline); err != nil {
		t.Fatalf("failed to read recovered inline file: %v", err)
	} else if string(data) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(data))
	}
}
