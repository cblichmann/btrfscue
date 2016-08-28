/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * BTRFS filesystem structures
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

package btrfs // import "blichmann.eu/code/btrfscue/btrfs"

import (
	"encoding/gob"
	"time"

	"blichmann.eu/code/btrfscue/uuid"

	"fmt" //DBG!!!
)

const (
	// Magic spells "_BHRfS_M" in little-endian
	Magic = 0x4d5f53665248425f

	// DefaultBlockSize is the default block size for BTRFS. It is the size
	// of a single page on x86 (4096 bytes).
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

	// Stores information about which extents are in use, and reference
	// counts
	ExtentTreeObjectId = 2

	// The chunk tree stores translations from logical -> physical block
	// numbering the super block points to the chunk tree
	ChunkTreeObjectId = 3

	// Stores information about which areas of a given device are in use. One
	// per device. The tree of tree roots points to the device tree.
	DevTreeObjectId = 4

	// One per subvolume, storing files and directories
	FSTreeObjectId = 5

	// Directory objectid inside the root tree
	RootTreeDirObjectId = 6

	// Holds checksums of all the data extents
	CSumTreeObjectId = 7

	// Holds quota configuration and tracking
	QuotaTreeObjectId = 8

	// For storing items that use the BTRFS_UUID_KEY* types
	UuidTreeObjectId = 9

	// Tracks free space in block groups
	FreeSpaceTreeObjectId = 10

	// Device stats in the device tree
	DevStatsObjectId = 0

	// For storing balance parameters in the root tree
	BalanceObjectId = ^uint64(4) + 1

	// Orphan objectid for tracking unlinked/truncated files
	OrphanObjectId = ^uint64(5) + 1

	// Does write ahead logging to speed up fsyncs
	TreeLogObjectId      = ^uint64(6) + 1
	TreeLogFixupObjectId = ^uint64(7) + 1

	// For space balancing
	TreeRelocObjectId     = ^uint64(8) + 1
	DataRelocTreeObjectId = ^uint64(9) + 1

	// Extent checksums all have this objectid. This allows them to share the
	// logging tree for fsyncs.
	ExtentCSumObjectId = ^uint64(10) + 1

	// For storing free space cache
	FreeSpaceObjectId = ^uint64(11) + 1

	// The inode number assigned to the special inode for storing free inode
	// cache
	FreeInoObjectId = ^uint64(12) + 1

	// Dummy objectid represents multiple objectids
	MultipleObjectIds = ^uint64(255) + 1

	// All files have objectids in this range
	FirstFreeObjectId      = 256
	LastFreeObjectId       = ^uint64(256) + 1
	FirstChunkTreeObjectId = 256

	// The device items go into the chunk tree. The key is in the form
	// [ 1 DevItemKey device_id ]
	DevItemsObjectId = 1

	BtreeInodeObjectId = 1

	EmptySubvolDirObjectId = 2

	// Maximum value of an objectid
	LastObjectId = ^uint64(0)
)

// Entity sizes
const (
	CSumSize             = 32
	LabelSize            = 256
	SystemChunkArraySize = 2048
)

// Key types
const (
	// Inode items have the data typically returned from stat and store other
	// info about object characteristics. There is one for every file and dir
	// in the FS.
	InodeItemKey = 1

	InodeRefKey    = 12
	InodeExtrefKey = 13
	XAttrItemKey   = 24
	OrphanItemKey  = 48

	// dir items are the name -> inode pointers in a directory. There is one
	// for every name in a directory.
	DirLogItemKey  = 60
	DirLogIndexKey = 72
	DirItemKey     = 84
	DirIndexKey    = 96

	// Extent data is for file data.
	ExtentDataKey = 108

	// Extent csums are stored in a separate tree and hold csums for
	// an entire extent on disk.
	ExtentCSumKey = 128

	// Root items point to tree roots. They are typically in the root
	// tree used by the super block to find all the other trees.
	RootItemKey = 132

	// Root backrefs tie subvols and snapshots to the directory entries that
	// reference them.
	RootBackRefKey = 144

	// Root refs make a fast index for listing all of the snapshots and
	// subvolumes referenced by a given root. They point directly to the
	// directory item in the root that references the subvol.
	RootRefKey = 156

	// Extent items are in the extent map tree. These record which blocks
	// are used, and how many references there are to each block.
	ExtentItemKey = 168

	// The same as the ExtentItemKey, except it's metadata we already know
	// the length, so we save the level in key->offset instead of the
	// length.
	MetadataItemKey = 169

	// These are contained within an extent item.
	// TODO(cblichmann): Parse these
	TreeBlockRefKey   = 176
	ExtentDataRefKey  = 178
	ExtentRefV0Key    = 180
	SharedBlockRefKey = 182
	SharedDataRefKey  = 184

	// Block groups give us hints into the extent allocation trees. Which
	// blocks are free etc.
	BlockGroupItemKey = 192

	// Every block group is represented in the free space tree by a free space
	// info item, which stores some accounting information. It is keyed on
	// (block_group_start, FREE_SPACE_INFO, block_group_length).
	FreeSpaceInfoKey = 198

	// A free space extent tracks an extent of space that is free in a block
	// group. It is keyed on (start, FREE_SPACE_EXTENT, length).
	FreeSpaceExtentKey = 199

	// When a block group becomes very fragmented, we convert it to use bitmaps
	// instead of extents. A free space bitmap is keyed on
	// (start, FREE_SPACE_BITMAP, length); the corresponding item is a bitmap
	// with (length / sectorsize) bits.
	FreeSpaceBitmapKey = 200

	DevExtentKey = 204
	DevItemKey   = 216
	ChunkItemKey = 228

	// Records the overall state of the qgroups.
	// There's only one instance of this key present,
	// (0, BTRFS_QGROUP_STATUS_KEY, 0)
	QgroupStatusKey = 240

	// Records the currently used space of the qgroup.
	// One key per qgroup, (0, BTRFS_QGROUP_INFO_KEY, qgroupid).
	QgroupInfoKey = 242

	// Contains the user configured limits for the qgroup.
	// One key per qgroup, (0, BTRFS_QGROUP_LIMIT_KEY, qgroupid).
	QgroupLimitKey = 244

	// Records the child-parent relationship of qgroups. For each relation, 2
	// keys are present:
	// (childid, BTRFS_QGROUP_RELATION_KEY, parentid)
	// (parentid, BTRFS_QGROUP_RELATION_KEY, childid)
	QgroupRelationKey = 246

	// Obsolete name, see BTRFS_TEMPORARY_ITEM_KEY
	BalanceItemKey = TemporaryItemKey

	// The key type for tree items that are stored persistently, but do not need
	// to exist for extended period of time. The items can exist in any tree.
	//
	// [subtype, BTRFS_TEMPORARY_ITEM_KEY, data]
	//
	// Existing items:
	//
	// - balance status item
	//   (BTRFS_BALANCE_OBJECTID, BTRFS_TEMPORARY_ITEM_KEY, 0)
	TemporaryItemKey = 248

	// Obsolete name, see BTRFS_PERSISTENT_ITEM_KEY
	DevStatsKey = PersistentItemKey

	// The key type for tree items that are stored persistently and usually
	// exist for a long period, eg. filesystem lifetime. The item kinds can be
	// status information, stats or preference values. The item can exist in
	// any tree.
	//
	// [subtype, BTRFS_PERSISTENT_ITEM_KEY, data]
	//
	// Existing items:
	//
	// - device statistics, store IO stats in the device tree, one key for all
	//   stats
	//   (BTRFS_DEV_STATS_OBJECTID, BTRFS_DEV_STATS_KEY, 0)
	PersistentItemKey = 249

	// Persistantly stores the device replace state in the device tree.
	// The key is built like this: (0, BTRFS_DEV_REPLACE_KEY, 0).
	DevReplaceKey = 250

	// Stores items that allow to quickly map UUIDs to something else.
	// These items are part of the filesystem UUID tree.
	// The key is built like this:
	// (UUID_upper_64_bits, BTRFS_UUID_KEY*, UUID_lower_64_bits).
	//
	// For UUIDs assigned to subvols
	UUIDKeySubvol = 251

	// For UUIDs assigned to received subvols
	UUIDKeyReceivedSubvol = 252

	// String items are for debugging. They just store a short string of data
	// in the FS.
	StringItemKey = 253
)

// CSum holds raw checksum bytes
type CSum [CSumSize]byte

type Header struct {
	CSum CSum
	// The following three fields must match struct SuperBlock
	// File system specific UUID
	FSID uuid.UUID
	// The start of this block relative to the begining of the backing device
	ByteNr uint64
	Flags  uint64
	// Allowed to be different from SuperBlock from here on
	ChunkTreeUUID uuid.UUID
	Generation    uint64
	Owner         uint64
	NrItems       uint32
	Level         uint8
}

func (h *Header) Parse(b *ParseBuffer) {
	copy(h.CSum[:], b.Next(CSumSize))
	copy(h.FSID[:], b.Next(uuid.UUIDSize))
	h.ByteNr = b.NextUint64()
	h.Flags = b.NextUint64()
	copy(h.ChunkTreeUUID[:], b.Next(uuid.UUIDSize))
	h.Generation = b.NextUint64()
	h.Owner = b.NextUint64()
	h.NrItems = b.NextUint32()
	h.Level = b.NextUint8()
}

func (h *Header) IsLeaf() bool {
	return h.Level == 0
}

type Key struct {
	ObjectID uint64
	Type     uint8
	Offset   uint64
}

func (k *Key) Parse(b *ParseBuffer) {
	k.ObjectID = b.NextUint64()
	k.Type = b.NextUint8()
	k.Offset = b.NextUint64()
}

type parseable interface {
	Parse(b *ParseBuffer)
}

type Item struct {
	Key
	Offset uint32
	Size   uint32
	Data   parseable
}

func (i *Item) Parse(b *ParseBuffer) {
	i.Key.Parse(b)
	i.Offset = b.NextUint32()
	i.Size = b.NextUint32()
}

func (i *Item) ParseData(b *ParseBuffer) {
	switch i.Type {
	case InodeItemKey:
		i.Data = &InodeItem{}
	case InodeRefKey:
		i.Data = &InodeRefItem{}
	case XAttrItemKey:
		fallthrough
	case DirItemKey:
		fallthrough
	case DirIndexKey:
		i.Data = &DirItem{}
	case ExtentDataKey:
		i.Data = &FileExtentItem{}
	case ExtentCSumKey:
		i.Data = &CSumItem{}
	case RootItemKey:
		i.Data = &RootItem{}
	case RootBackRefKey:
		fallthrough
	case RootRefKey:
		i.Data = &RootRef{}
	case ExtentItemKey:
		i.Data = &ExtentItem{compatV0: i.Size < 8+8+8}
	case MetadataItemKey:
		// TODO(cblichmann): Special metadata handling for extents
		i.Data = &ExtentItem{}
	case BlockGroupItemKey:
		i.Data = &BlockGroupItem{}
	case DevExtentKey:
		i.Data = &DevExtent{}
	case DevItemKey:
		i.Data = &DevItem{}
	case ChunkItemKey:
		i.Data = &Chunk{}
	default:
		return
	}
	i.Data.(parseable).Parse(b)
}

// Dev extents record free space on individual devices. The owner field
// points back to the chunk allocation mapping tree that allocated the
// extent. The chunk tree uuid field is a way to double check the owner.
type DevExtent struct {
	ChunkTree     uint64
	ChunkObjectID uint64
	ChunkOffset   uint64
	Length        uint64
	ChunkTreeUUID uuid.UUID
}

func (i *DevExtent) Parse(b *ParseBuffer) {
	i.ChunkTree = b.NextUint64()
	i.ChunkObjectID = b.NextUint64()
	i.Length = b.NextUint64()
	copy(i.ChunkTreeUUID[:], b.Next(uuid.UUIDSize))
}

type DevItem struct {
	// The internal BTRFS device id
	DevID uint64

	// Size of the device
	TotalBytes uint64

	// Bytes used
	BytesUsed uint64

	// Optimal I/O alignment for this device
	IOAlign uint32

	// Optimal I/O width for this device
	IOWidth uint32

	// Minimal I/O size for this device
	SectorSize uint32

	// Type and info about this device
	Type uint64

	// Expected generation for this device
	Generation uint64

	// Starting byte of this partition on the device, to allow for stripe
	// alignment in the future
	StartOffset uint64

	// Grouping information for allocation decisions
	DevGroup uint32

	// Seek speed 0-100 where 100 is fastest
	SeekSpeed uint8

	// Bandwidth 0-100 where 100 is fastest
	Bandwidth uint8

	// BTRFS generated UUID for this device
	UUID uuid.UUID

	// UUID of FS that owns this device
	FSID uuid.UUID
}

func (i *DevItem) Parse(b *ParseBuffer) {
	i.DevID = b.NextUint64()
	i.TotalBytes = b.NextUint64()
	i.BytesUsed = b.NextUint64()
	i.IOAlign = b.NextUint32()
	i.IOWidth = b.NextUint32()
	i.SectorSize = b.NextUint32()
	i.Type = b.NextUint64()
	i.Generation = b.NextUint64()
	i.StartOffset = b.NextUint64()
	i.DevGroup = b.NextUint32()
	i.SeekSpeed = b.NextUint8()
	i.Bandwidth = b.NextUint8()
	copy(i.UUID[:], b.Next(uuid.UUIDSize))
	copy(i.FSID[:], b.Next(uuid.UUIDSize))
}

type Stripe struct {
	DevID   uint64
	Offset  uint64
	DevUUID uuid.UUID
}

func (i *Stripe) Parse(b *ParseBuffer) {
	i.DevID = b.NextUint64()
	i.Offset = b.NextUint64()
	copy(i.DevUUID[:], b.Next(uuid.UUIDSize))
}

type Chunk struct {
	// Size of this chunk in bytes
	Length uint64

	// ObjectID of the root referencing this chunk
	Owner uint64

	StripeLen uint64
	Type      uint64

	// Optimal IO alignment for this chunk
	IOAlign uint32

	// Optimal IO width for this chunk
	IOWidth uint32

	// Minimal IO size for this chunk
	SectorSize uint32

	// 2^16 stripes is quite a lot, a second limit is the size of a single
	// item in the btree
	NumStripes uint16

	// Sub stripes only matter for raid10
	SubStripes uint16
	Stripes    []Stripe
}

func (i *Chunk) Parse(b *ParseBuffer) {
	i.Length = b.NextUint64()
	i.Owner = b.NextUint64()
	i.StripeLen = b.NextUint64()
	i.Type = b.NextUint64()
	i.IOAlign = b.NextUint32()
	i.IOWidth = b.NextUint32()
	i.SectorSize = b.NextUint32()
	i.NumStripes = b.NextUint16()
	i.SubStripes = b.NextUint16()

	// Do not limit the number of stripes we allocated here. Worst case is
	// 64k stripes.
	i.Stripes = make([]Stripe, i.NumStripes)
	for s := 0; s < int(i.NumStripes); s++ {
		i.Stripes[s].Parse(b)
	}
}

type InodeItem struct {
	// NFS style generation number
	Generation uint64
	// Transid that last touched this inode
	Transid    uint64
	Size       uint64
	Nbytes     uint64
	BlockGroup uint64
	Nlink      uint32
	UID        uint32
	GID        uint32
	Mode       uint32
	Rdev       uint64
	Flags      uint64

	// Modification sequence number for NFS
	Sequence uint64

	// A little future expansion, for more than this we can just grow the
	// inode item and version it.
	Reserved [4]uint64
	Atime    time.Time
	Ctime    time.Time
	Mtime    time.Time
	Otime    time.Time
}

func (i *InodeItem) Parse(b *ParseBuffer) {
	i.Generation = b.NextUint64()
	i.Transid = b.NextUint64()
	i.Size = b.NextUint64()
	i.Nbytes = b.NextUint64()
	i.BlockGroup = b.NextUint64()
	i.Nlink = b.NextUint32()
	i.UID = b.NextUint32()
	i.GID = b.NextUint32()
	i.Mode = b.NextUint32()
	i.Rdev = b.NextUint64()
	i.Flags = b.NextUint64()
	i.Sequence = b.NextUint64()
	for j, _ := range i.Reserved {
		i.Reserved[j] = b.NextUint64()
	}
	i.Atime = b.NextTime().UTC()
	i.Ctime = b.NextTime().UTC()
	i.Mtime = b.NextTime().UTC()
	i.Otime = b.NextTime().UTC()
}

type InodeRefItem struct {
	Index   uint64
	NameLen uint16
	Name    string
}

func (i *InodeRefItem) Parse(b *ParseBuffer) {
	i.Index = b.NextUint64()
	i.NameLen = b.NextUint16()
	l := int(i.NameLen)
	if l > 255 {
		l = 255
	}
	i.Name = string(b.Next(l))
}

// Directory item type
const (
	FtUnknown = iota
	FtRegFile
	FtDir
	FtChrdev
	FtBlkdev
	FtFifo
	FtSock
	FtSymlink
	FtXattr
	FtMax
)

type DirItem struct {
	Location Key
	TransId  uint64
	DataLen  uint16
	NameLen  uint16
	Type     uint8
	Name     string
	Data     string
}

func (i *DirItem) Parse(b *ParseBuffer) {
	i.Location.Parse(b)
	i.TransId = b.NextUint64()
	i.DataLen = b.NextUint16()
	i.NameLen = b.NextUint16()
	i.Type = b.NextUint8()
	l := int(i.NameLen)
	if l > 255 {
		l = 255
	}
	i.Name = string(b.Next(l))
	l = int(i.DataLen)
	if l > DefaultBlockSize {
		l = DefaultBlockSize
	}
	i.Data = string(b.Next(l))
}

func (i *DirItem) IsDir() bool { return i.Type == FtDir }

const (
	BlockGroupData = 1 << iota
	BlockGroupSystem
	BlockGroupMetadata
	BlockGroupRaid0
	BlockGroupRaid1
	BlockGroupDup
	BlockGroupRaid10
	BlockGroupRaid5
	BlockGroupRaid6
	// TODO(cblichmann): More constants and the block group masks
	// BlockGroupReserved = AVAIL_ALLOC_BIT_SINGLE | SPACE_INFO_GLOBAL_RSV
)

type BlockGroupItem struct {
	Used          uint64
	ChunkObjectID uint64
	Flags         uint64
}

func (i *BlockGroupItem) Parse(b *ParseBuffer) {
	i.Used = b.NextUint64()
	i.ChunkObjectID = b.NextUint64()
	i.Flags = b.NextUint64()
}

const (
	FileExtentInline = iota
	FileExtentReg
	FileExtentPreAlloc
)

type FileExtentItem struct {
	// Transaction id that created this extent
	Generation uint64

	// Max number of bytes to hold this extent in ram when we split a
	// compressed extent we can't know how big each of the resulting pieces
	// will be. So, this is an upper limit on the size of the extent in ram
	// instead of an exact limit.
	RAMBytes uint64

	// 32 bits for the various ways we might encode the data, including
	// compression and encryption. If any of these are set to something a
	// given disk format doesn't understand it is treated like an incompat
	// flag for reading and writing, but not for stat.
	Compression   uint8
	Encryption    uint8
	OtherEncoding uint16 // For later use

	// Are we inline data or a real extent?
	Type uint8

	// Disk space consumed by the extent, checksum blocks are included in
	// these numbers.

	// At this offset in the structure, the inline extent data start.
	// The following fields are valid only if Type != FileExtentInline:
	DiskByteNr   uint64
	DiskNumBytes uint64

	// The logical offset in file blocks (no csums) this extent record is
	// for. This allows a file extent to point into the middle of an existing
	// extent on disk, sharing it between two snapshots (useful if some bytes
	// in the middle of the extent have changed.
	Offset uint64

	// The logical number of file blocks (no csums included). This always
	// reflects the size uncompressed and without encoding.
	NumBytes uint64

	// This field is only set if Type == FileExtentInline:
	Data string
}

func (i *FileExtentItem) Parse(b *ParseBuffer) {
	i.Generation = b.NextUint64()
	i.RAMBytes = b.NextUint64()
	i.Compression = b.NextUint8()
	i.Encryption = b.NextUint8()
	i.OtherEncoding = b.NextUint16()
	i.Type = b.NextUint8()
	if i.Type != FileExtentInline {
		i.DiskByteNr = b.NextUint64()
		i.DiskNumBytes = b.NextUint64()
		i.Offset = b.NextUint64()
		i.NumBytes = b.NextUint64()
	} else {
		l := int(i.RAMBytes)
		if l > DefaultBlockSize {
			l = DefaultBlockSize
		}
		i.Data = string(b.Next(l))
	}
}

type CSumItem struct {
	CSum uint8
}

func (i *CSumItem) Parse(b *ParseBuffer) {
	i.CSum = b.NextUint8()
	// TODO(cblichmann): Parse the actual checksums
}

type RootItem struct {
	Inode        InodeItem
	Generation   uint64
	RootDirID    uint64
	ByteNr       uint64
	ByteLimit    uint64
	BytesUsed    uint64
	LastSnapshot uint64
	Flags        uint64
	Refs         uint32
	DropProgress Key
	DropLevel    uint8
	Level        uint8

	// The following fields appear after subvol_uuids+subvol_times were
	// introduced.

	// This generation number is used to test if the new fields are valid
	// and up to date while reading the root item. Everytime the root item
	// is written out, the "generation" field is copied into this field. If
	// anyone ever mounted the fs with an older kernel, we will have
	// mismatching generation values here and thus must invalidate the
	// new fields.
	GenerationV2 uint64
	UUID         uuid.UUID
	ParentUUID   uuid.UUID
	ReceivedUUID uuid.UUID
	CTransID     uint64 // Updated when an inode changes
	OTransID     uint64 // Trans when created
	STransID     uint64 // Trans when sent. Non-zero for received subvol
	RTransID     uint64 // Trans when received. Non-zero for received subvol
	Ctime        time.Time
	Otime        time.Time
	Stime        time.Time
	Rtime        time.Time
	Reserved     [8]uint64
}

func (i *RootItem) Parse(b *ParseBuffer) {
	i.Inode.Parse(b)
	i.Generation = b.NextUint64()
	i.RootDirID = b.NextUint64()
	i.ByteNr = b.NextUint64()
	i.ByteLimit = b.NextUint64()
	i.LastSnapshot = b.NextUint64()
	i.Flags = b.NextUint64()
	i.Refs = b.NextUint32()
	i.DropProgress.Parse(b)
	i.DropLevel = b.NextUint8()
	i.Level = b.NextUint8()
	i.GenerationV2 = b.NextUint64()
	if i.Generation == i.GenerationV2 {
		copy(i.UUID[:], b.Next(uuid.UUIDSize))
		copy(i.ParentUUID[:], b.Next(uuid.UUIDSize))
		copy(i.ReceivedUUID[:], b.Next(uuid.UUIDSize))
		i.CTransID = b.NextUint64()
		i.OTransID = b.NextUint64()
		i.STransID = b.NextUint64()
		i.RTransID = b.NextUint64()
		i.Ctime = b.NextTime().UTC()
		i.Otime = b.NextTime().UTC()
		i.Stime = b.NextTime().UTC()
		i.Rtime = b.NextTime().UTC()
		for j := range i.Reserved {
			i.Reserved[j] = b.NextUint64()
		}
	}
}

// This is used for both forward and backward root refs
type RootRef struct {
	DirID    uint64
	Sequence uint64
	NameLen  uint16
	Name     string
}

func (i *RootRef) Parse(b *ParseBuffer) {
	i.DirID = b.NextUint64()
	i.Sequence = b.NextUint64()
	i.NameLen = b.NextUint16()
	l := int(i.NameLen)
	if l > 255 {
		l = 255
	}
	i.Name = string(b.Next(l))
}

const (
	ExtentFlagData = 1 << iota
	ExtentFlagTreeBlock
)

// Items in the extent btree are used to record the objectid of the
// owner of the block and the number of references.
type ExtentItem struct {
	Refs       uint64
	Generation uint64
	Flags      uint64
	compatV0   bool
}

func (i *ExtentItem) Parse(b *ParseBuffer) {
	if !i.IsCompatV0() {
		i.Refs = b.NextUint64()
		i.Generation = b.NextUint64()
		i.Flags = b.NextUint64()
	} else {
		i.Refs = uint64(b.NextUint32())
	}
}

func (i *ExtentItem) IsCompatV0() bool { return i.compatV0 }

type Leaf struct {
	Header
	Items []Item
}

func (l *Leaf) Parse(b *ParseBuffer) {
	if l.Header.NrItems == 0 {
		return
	}
	headerEnd := uint32(b.Offset())
	// Clamp maximum number of items to avoid running OOM in case NrItems is
	// corrupted. 0x19 is the typical item size without item data.
	maxItems := b.Unread() / 0x19
	numItems := l.Header.NrItems
	if numItems > uint32(maxItems) {
		numItems = uint32(maxItems)
	}

	l.Items = make([]Item, numItems)
	for i := range l.Items {
		l.Items[i].Parse(b)
	}
	for i := range l.Items {
		item := &l.Items[i]
		o := int(headerEnd) + int(item.Offset)
		if o >= b.Len() {
			continue
		}
		b.SetOffset(o)
		item.ParseData(b)
	}
}

func init() {
	fmt.Printf("") //DBG!!!
	// Register with custom short names
	gob.RegisterName("BlkG", &BlockGroupItem{})
	gob.RegisterName("CSum", &CSumItem{})
	gob.RegisterName("Chnk", &Chunk{})
	gob.RegisterName("DExt", &DevExtent{})
	gob.RegisterName("DevI", &DevItem{})
	gob.RegisterName("DirI", &DirItem{})
	gob.RegisterName("ExtI", &ExtentItem{})
	gob.RegisterName("FExt", &FileExtentItem{})
	gob.RegisterName("InoR", &InodeRefItem{})
	gob.RegisterName("Inod", &InodeItem{})
	gob.RegisterName("Root", &RootItem{})
	gob.RegisterName("RtRf", &RootRef{})
}
