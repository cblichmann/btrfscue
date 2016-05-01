package btrfs // import "blichmann.eu/code/btrfscue/btrfs"

import (
	"bytes"
	"encoding/binary"
)

const (
	// "_BHRfS_M" in little-endian
	Magic = 0x4D5F53665248425F

	DefaultBlockSize = 1 << 12
)

// Offsets of all superblock copies
const (
	SuperInfoOffset  = 0x10000         // 64 KiB
	SuperInfoOffset2 = 0x4000000       // 64 MiB
	SuperInfoOffset3 = 0x4000000000    // 256 GiB
	SuperInfoOffset4 = 0x4000000000000 // 1 PiB
)

// Object ids
const (
	// Holds pointers to all of the tree roots
	RootTreeObjectId = 1

	// Stores information about which extents are in use, and reference counts.
	ExtentTreeObjectId = 2

	// The chunk tree stores translations from logical -> physical block numbering the super block points to the chunk tree.
	ChunkTreeObjectId = 3

	// Stores information about which areas of a given device are in use. One per device. The tree of tree roots points to the device tree.
	DevTreeObjectId = 4

	// One per subvolume, storing files and directories
	FsTreeObjectId = 5

	// Directory objectid inside the root tree
	RootTreeDirObjectId = 6

	// Holds checksums of all the data extents
	CsumTreeObjectId = 7

	// Orhpan objectid for tracking unlinked/truncated files
	OrphanObjectId = ^uint64(5) + 1

	// Does write ahead logging to speed up fsyncs
	TreeLogObjectId      = ^uint64(6) + 1
	TreeLogFixupObjectId = ^uint64(7) + 1

	// For space balancing
	TreeRelocObjectId     = ^uint64(8) + 1
	DataRelocTreeObjectId = ^uint64(9) + 1

	// Extent checksums all have this objectid. This allows them to share the logging tree for fsyncs.
	ExtentCsumObjectId = ^uint64(10) + 1

	// For storing free space cache */
	FreeSpaceObjectId = ^uint64(11) + 1

	// Dummy objectid represents multiple objectids
	MultipleObjectIdS = ^uint64(255) + 1

	// All files have objectids in this range
	FirstFreeObjectId      = 256
	LastFreeObjectId       = ^uint64(256) + 1
	FirstChunkTreeObjectId = 256

	// The device items go into the chunk tree. The key is in the form [ 1 BTRFS_DEV_ITEM_KEY device_id ]
	DevItemsObjectId = 1

	BtreeInodeObjectId = 1

	EmptySubvolDirObjectId = 2
)

// Entity sizes
const (
	UuidSize             = 16
	CsumSize             = 32
	LabelSize            = 256
	SystemChunkArraySize = 2048
)

type Uuid [UuidSize]byte

func (u Uuid) String() string {
	const hexChars = "0123456789abcdef"
	buf := [UuidSize*2 + 4]byte{}
	p := 0
	for i, v := range u {
		buf[p] = hexChars[v>>4]
		p++
		buf[p] = hexChars[v&0xF]
		if i == 3 || i == 5 || i == 7 || i == 9 {
			p++
			buf[p] = '-'
		}
		p++
	}
	return string(buf[:])
}

type Csum [CsumSize]byte

type ParseBuffer struct {
	*bytes.Buffer
	ByteOrder binary.ByteOrder
}

func NewParseBuffer(buf []byte) *ParseBuffer {
	return &ParseBuffer{
		Buffer:    bytes.NewBuffer(buf),
		ByteOrder: binary.LittleEndian,
	}
}

func (b *ParseBuffer) NextUint8() uint8 {
	return b.Next(1)[0]
}

func (b *ParseBuffer) NextUint16() uint16 {
	return b.ByteOrder.Uint16(b.Next(2))
}

func (b *ParseBuffer) NextUint32() uint32 {
	return b.ByteOrder.Uint32(b.Next(4))
}

func (b *ParseBuffer) NextUint64() uint64 {
	return b.ByteOrder.Uint64(b.Next(8))
}

type Header struct {
	Csum Csum
	// The following three fields must match struct SuperBlock
	// File system specific UUID
	FsId Uuid
	// The start of this block relative to the begining of the backing device
	ByteNr uint64
	Flags  uint64
	// Allowed to be different from SuperBlock from here on
	ChunkTreeUuid Uuid
	Generation    uint64
	Owner         uint64
	NrItems       uint32
	Level         uint8
}

func NewHeader(block []byte) *Header {
	h := &Header{}
	h.Parse(block)
	return h
}

func (h *Header) Parse(block []byte) {
	b := NewParseBuffer(block)
	copy(h.Csum[:], b.Next(CsumSize))
	copy(h.FsId[:], b.Next(UuidSize))
	h.ByteNr = b.NextUint64()
	h.Flags = b.NextUint64()
	copy(h.ChunkTreeUuid[:], b.Next(UuidSize))
	h.Generation = b.NextUint64()
	h.Owner = b.NextUint64()
	h.NrItems = b.NextUint32()
	h.Level = b.NextUint8()
}

func (h *Header) IsLeaf() bool {
	return h.Level == 0
}
