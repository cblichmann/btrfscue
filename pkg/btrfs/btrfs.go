/*
 * btrfscue version 0.6
 * Copyright (c)2011-2021 Christian Blichmann
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

package btrfs

import (
	"time"

	"blichmann.eu/code/btrfscue/pkg/uuid"
)

const (
	// Magic spells "_BHRfS_M" in little-endian
	Magic = 0x4d5f53665248425f

	// X86RegularPageSize is the size of a regular memory page on X86.
	X86RegularPageSize = 1 << 12

	// DefaultBlockSize is the default block size for BTRFS. It is the size
	// of four pages on x86 (16384 bytes).
	DefaultBlockSize = 4 * X86RegularPageSize
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
	RootTreeObjectID = 1

	// Stores information about which extents are in use, and reference
	// counts
	ExtentTreeObjectID = 2

	// The chunk tree stores translations from logical -> physical block
	// numbering the super block points to the chunk tree
	ChunkTreeObjectID = 3

	// Stores information about which areas of a given device are in use. One
	// per device. The tree of tree roots points to the device tree.
	DevTreeObjectID = 4

	// One per subvolume, storing files and directories
	FSTreeObjectID = 5

	// Directory objectid inside the root tree
	RootTreeDirObjectID = 6

	// Holds checksums of all the data extents
	CSumTreeObjectID = 7

	// Holds quota configuration and tracking
	QuotaTreeObjectID = 8

	// For storing items that use the BTRFS_UUID_KEY* types
	UuidTreeObjectID = 9

	// Tracks free space in block groups
	FreeSpaceTreeObjectID = 10

	// Device stats in the device tree
	DevStatsObjectID = 0

	// For storing balance parameters in the root tree
	BalanceObjectID = ^uint64(4) + 1

	// Orphan objectid for tracking unlinked/truncated files
	OrphanObjectID = ^uint64(5) + 1

	// Does write ahead logging to speed up fsyncs
	TreeLogObjectID      = ^uint64(6) + 1
	TreeLogFixupObjectID = ^uint64(7) + 1

	// For space balancing
	TreeRelocObjectID     = ^uint64(8) + 1
	DataRelocTreeObjectID = ^uint64(9) + 1

	// Extent checksums all have this objectid. This allows them to share the
	// logging tree for fsyncs.
	ExtentCSumObjectID = ^uint64(10) + 1

	// For storing free space cache
	FreeSpaceObjectID = ^uint64(11) + 1

	// The inode number assigned to the special inode for storing free inode
	// cache
	FreeInoObjectID = ^uint64(12) + 1

	// Dummy objectid represents multiple objectids
	MultipleObjectIDs = ^uint64(255) + 1

	// All files have objectids in this range
	FirstFreeObjectID = 256
	LastFreeObjectID  = ^uint64(256) + 1

	FirstChunkTreeObjectID = 256

	// The device items go into the chunk tree. The key is in the form
	// [ 1 DevItemKey device_id ]
	DevItemsObjectID = 1

	BtreeInodeObjectID = 1

	EmptySubvolDirObjectID = 2

	// Maximum value of an objectid
	LastObjectID = ^uint64(0)
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

type Key struct {
	ObjectID uint64
	Type     uint8
	Offset   uint64
}

const KeyLen = 8 + 1 + 8

// KeyCompare compares two BTRFS keys lexicographically. It returns 0 if
// a==b, -1 if a < b and +1 if a > b.
func KeyCompare(a, b Key) int {
	if r := int(a.Type) - int(b.Type); r != 0 {
		return r
	}
	if a.ObjectID < b.ObjectID {
		return -1
	}
	if a.ObjectID > b.ObjectID {
		return 1
	}
	if a.Offset < b.Offset {
		return -1
	}
	if a.Offset > b.Offset {
		return 1
	}
	return 0
}

// CSum holds raw checksum bytes
type CSum [CSumSize]byte

// Dev extents record free space on individual devices. The owner field
// points back to the chunk allocation mapping tree that allocated the
// extent. The chunk tree uuid field is a way to double check the owner.
type DevExtent []byte

// Dev extent offsets for parsing from byte slice
const (
	devExtentChunkTree     = 0
	devExtentChunkObjectID = devExtentChunkTree + 8
	devExtentChunkOffset   = devExtentChunkObjectID + 8
	devExtentLength        = devExtentChunkOffset + 8
	devExtentChunkTreeUUID = devExtentLength + 8
	DevExtentLen           = devExtentChunkTreeUUID + uuid.UUIDSize
)

func (e DevExtent) ChunkTree() uint64        { return SliceUint64LE(e[devExtentChunkTree:]) }
func (e DevExtent) ChunkObjectID() uint64    { return SliceUint64LE(e[devExtentChunkObjectID:]) }
func (e DevExtent) ChunkOffset() uint64      { return SliceUint64LE(e[devExtentChunkOffset:]) }
func (e DevExtent) Length() uint64           { return SliceUint64LE(e[devExtentLength:]) }
func (e DevExtent) ChunkTreeUUID() uuid.UUID { return SliceUUID(e[devExtentChunkTreeUUID:]) }

type DevItem []byte

const (
	devItemDevID       = 0
	devItemTotalBytes  = devItemDevID + 8
	devItemBytesUsed   = devItemTotalBytes + 8
	devItemIOAlign     = devItemBytesUsed + 8
	devItemIOWidth     = devItemIOAlign + 4
	devItemSectorSize  = devItemIOWidth + 4
	devItemType        = devItemSectorSize + 4
	devItemGeneration  = devItemType + 8
	devItemStartOffset = devItemGeneration + 8
	devItemDevGroup    = devItemStartOffset + 8
	devItemSeekSpeed   = devItemDevGroup + 4
	devItemBandwidth   = devItemSeekSpeed + 1
	devItemUUID        = devItemBandwidth + 1
	devItemFSID        = devItemUUID + uuid.UUIDSize
	DevItemLen         = devItemFSID + uuid.UUIDSize
)

// DevID returns the internal BTRFS device id
func (i DevItem) DevID() uint64 { return SliceUint64LE(i[devItemDevID:]) }

// TotalBytes returns the size of the device
func (i DevItem) TotalBytes() uint64 { return SliceUint64LE(i[devItemTotalBytes:]) }

// BytesUsed returns the number of bytes used
func (i DevItem) BytesUsed() uint64 { return SliceUint64LE(i[devItemBytesUsed:]) }

// IOAlign returns the optimal I/O alignment for this device
func (i DevItem) IOAlign() uint32 { return SliceUint32LE(i[devItemIOAlign:]) }

// IOWidth returns the optimal I/O width for this device
func (i DevItem) IOWidth() uint32 { return SliceUint32LE(i[devItemIOWidth:]) }

// SectorSize returns the minimal I/O size for this device
func (i DevItem) SectorSize() uint32 { return SliceUint32LE(i[devItemSectorSize:]) }

// Type returns the type and info about this device
func (i DevItem) Type() uint64 { return SliceUint64LE(i[devItemType:]) }

// Generation returns the expected generation for this device
func (i DevItem) Generation() uint64 { return SliceUint64LE(i[devItemGeneration:]) }

// StartOffset returns the starting byte of this partition on the device. This
// allows for stripe alignment in the future.
func (i DevItem) StartOffset() uint64 { return SliceUint64LE(i[devItemStartOffset:]) }

// DevGroup returns grouping information for allocation decisions
func (i DevItem) DevGroup() uint32 { return SliceUint32LE(i[devItemDevGroup:]) }

// SeekSpeed returns the device seek speed in range 0-100 where 100 is fastest
func (i DevItem) SeekSpeed() uint8 { return i[devItemSeekSpeed] }

// Bandwidth returns the device bandwidth in range 0-100 where 100 is fastest
func (i DevItem) Bandwidth() uint8 { return i[devItemBandwidth] }

// UUID returns the BTRFS generated UUID for this device
func (i DevItem) UUID() uuid.UUID { return SliceUUID(i[devItemUUID:]) }

// FSID returns the UUID of the FS that owns this device
func (i DevItem) FSID() uuid.UUID { return SliceUUID(i[devItemFSID:]) }

type Stripe []byte

// Stripe offsets for parsing from byte slice
const (
	stripeDevID   = 0
	stripeOffset  = stripeDevID + 8
	stripeDevUUID = stripeOffset + 8
	stripeEnd     = stripeDevUUID + uuid.UUIDSize
)

func (s Stripe) DevID() uint64      { return SliceUint64LE(s[stripeDevID:]) }
func (s Stripe) Offset() uint64     { return SliceUint64LE(s[stripeOffset:]) }
func (s Stripe) DevUUID() uuid.UUID { return SliceUUID(s[stripeDevUUID:]) }

type Chunk []byte

const (
	chunkLength     = 0
	chunkOwner      = chunkLength + 8
	chunkStripeLen  = chunkOwner + 8
	chunkType       = chunkStripeLen + 8
	chunkIOAlign    = chunkType + 8
	chunkIOWidth    = chunkIOAlign + 4
	chunkSectorSize = chunkIOWidth + 4
	chunkNumStripes = chunkSectorSize + 4
	chunkSubStripes = chunkNumStripes + 2
	chunkStripes    = chunkSubStripes + 2
)

// Length returns the size of this chunk in bytes
func (c Chunk) Length() uint64 { return SliceUint64LE(c[chunkLength:]) }

// Owner returns the ObjectID of the root referencing this chunk
func (c Chunk) Owner() uint64 { return SliceUint64LE(c[chunkOwner:]) }

func (c Chunk) StripeLen() uint64 { return SliceUint64LE(c[chunkStripeLen:]) }

// Type returns the type of this chunk. Reuses BlockGroupItem's Flags
func (c Chunk) Type() uint64 { return SliceUint64LE(c[chunkType:]) }

// IOAlign returns the optimal IO alignment for this chunk
func (c Chunk) IOAlign() uint32 { return SliceUint32LE(c[chunkIOAlign:]) }

// IOWidth returns the optimal IO width for this chunk
func (c Chunk) IOWidth() uint32 { return SliceUint32LE(c[chunkIOWidth:]) }

// SectorSize returns the minimal IO size for this chunk
func (c Chunk) SectorSize() uint32 { return SliceUint32LE(c[chunkSectorSize:]) }

// NumStripes returns the number of stripes in this chunk.
// 2^16 stripes is quite a lot, a second limit is the size of a single
// item in the btree
func (c Chunk) NumStripes() uint16 { return SliceUint16LE(c[chunkNumStripes:]) }

// Sub stripes only matter for raid10
func (c Chunk) SubStripes() uint16 { return SliceUint16LE(c[chunkSubStripes:]) }

// Stripe returns the ith stripe of this chunk.
func (c Chunk) Stripe(i uint16) Stripe {
	return Stripe(c[chunkStripes+i*stripeEnd:])
}

type InodeItem []byte

// InodeItem offsets for parsing from byte slice
const (
	// NFS style generation number
	inodeItemGeneration = 0
	// Transid that last touched this inode
	inodeItemTransID    = inodeItemGeneration + 8
	inodeItemSize       = inodeItemTransID + 8
	inodeItemNbytes     = inodeItemSize + 8
	inodeItemBlockGroup = inodeItemNbytes + 8
	inodeItemNlink      = inodeItemBlockGroup + 8
	inodeItemUID        = inodeItemNlink + 4
	inodeItemGID        = inodeItemUID + 4
	inodeItemMode       = inodeItemGID + 4
	inodeItemRdev       = inodeItemMode + 4
	inodeItemFlags      = inodeItemRdev + 8
	// Modification sequence number for NFS
	inodeItemSequence = inodeItemFlags + 8
	// Room for future expansion, for more than this we can just grow the
	// inode item and version it.
	inodeItemReserved = inodeItemSequence + 8
	inodeItemAtime    = inodeItemReserved + 32
	inodeItemCtime    = inodeItemAtime + 12
	inodeItemMtime    = inodeItemCtime + 12
	inodeItemOtime    = inodeItemMtime + 12
	InodeItemLen      = inodeItemOtime + 12
)

func (i InodeItem) Generation() uint64 { return SliceUint64LE(i[inodeItemGeneration:]) }
func (i InodeItem) TransID() uint64    { return SliceUint64LE(i[inodeItemTransID:]) }
func (i InodeItem) Size() uint64       { return SliceUint64LE(i[inodeItemSize:]) }
func (i InodeItem) BlockGroup() uint64 { return SliceUint64LE(i[inodeItemBlockGroup:]) }
func (i InodeItem) Nlink() uint32      { return SliceUint32LE(i[inodeItemNlink:]) }
func (i InodeItem) UID() uint32        { return SliceUint32LE(i[inodeItemUID:]) }
func (i InodeItem) GID() uint32        { return SliceUint32LE(i[inodeItemGID:]) }
func (i InodeItem) Mode() uint32       { return SliceUint32LE(i[inodeItemMode:]) }
func (i InodeItem) Rdev() uint64       { return SliceUint64LE(i[inodeItemRdev:]) }
func (i InodeItem) Flags() uint64      { return SliceUint64LE(i[inodeItemFlags:]) }
func (i InodeItem) Sequence() uint64   { return SliceUint64LE(i[inodeItemSequence:]) }
func (i InodeItem) Reserved() [4]uint64 {
	return [4]uint64{SliceUint64LE(i[inodeItemReserved:]),
		SliceUint64LE(i[inodeItemReserved+8:]),
		SliceUint64LE(i[inodeItemReserved+16:]),
		SliceUint64LE(i[inodeItemReserved+24:]),
	}
}
func (i InodeItem) Atime() time.Time { return SliceTimeLE(i[inodeItemAtime:]).UTC() }
func (i InodeItem) Ctime() time.Time { return SliceTimeLE(i[inodeItemCtime:]).UTC() }
func (i InodeItem) Mtime() time.Time { return SliceTimeLE(i[inodeItemMtime:]).UTC() }
func (i InodeItem) Rtime() time.Time { return SliceTimeLE(i[inodeItemOtime:]).UTC() }

type InodeRefItem []byte

// InodeRefItem offsets for parsing from byte slice
const (
	inodeRefItemIndex   = 0
	inodeRefItemNameLen = inodeRefItemIndex + 8
	inodeRefItemName    = inodeRefItemNameLen + 2
)

func (i InodeRefItem) Index() uint64   { return SliceUint64LE(i[inodeRefItemIndex:]) }
func (i InodeRefItem) NameLen() uint16 { return SliceUint16LE(i[inodeRefItemNameLen:]) }
func (i InodeRefItem) Name() string {
	l := int(i.NameLen())
	if l > 255 {
		l = 255
	}
	return string(i[inodeRefItemName : inodeRefItemName+l])
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

type DirItem []byte

// DirItem offsets for parsing from byte slice
const (
	dirItemLocation = 0
	dirItemTransID  = dirItemLocation + 8 + 1 + 8
	dirItemDataLen  = dirItemTransID + 8
	dirItemNameLen  = dirItemDataLen + 2
	dirItemType     = dirItemNameLen + 2
	dirItemName     = dirItemType + 1
)

func (d DirItem) Location() Key   { return SliceKey(d[dirItemLocation:]) }
func (d DirItem) TransID() uint64 { return SliceUint64LE(d[dirItemTransID:]) }
func (d DirItem) DataLen() uint16 { return SliceUint16LE(d[dirItemDataLen:]) }
func (d DirItem) NameLen() uint16 { return SliceUint16LE(d[dirItemNameLen:]) }
func (d DirItem) Type() uint8     { return uint8(d[dirItemType]) }

func (d DirItem) Name() string {
	l := int(d.NameLen())
	if l > 255 {
		l = 255
	}
	return string(d[dirItemName : dirItemName+l])
}

func (d DirItem) Data() string {
	o := dirItemName + d.NameLen()
	return string(d[o : o+d.DataLen()])
}

func (d DirItem) IsDir() bool       { return d.Type() == FtDir }
func (d DirItem) IsSubvolume() bool { return d.Location().Type == RootItemKey }

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

type BlockGroupItem []byte

// BlockGroupItem offsets for parsing from byte slice
const (
	blockGroupItemUsed          = 0
	blockGroupItemChunkObjectID = blockGroupItemUsed + 8
	blockGroupItemFlags         = blockGroupItemChunkObjectID + 8
	blockGroupItemEnd           = blockGroupItemFlags + 8
)

func (i BlockGroupItem) Used() uint64          { return SliceUint64LE(i[blockGroupItemUsed:]) }
func (i BlockGroupItem) ChunkObjectID() uint64 { return SliceUint64LE(i[blockGroupItemChunkObjectID:]) }
func (i BlockGroupItem) Flags() uint64         { return SliceUint64LE(i[blockGroupItemFlags:]) }

// File extent type
const (
	FileExtentInline = iota
	FileExtentReg
	FileExtentPreAlloc
)

type FileExtentItem []byte

// FileExtentItem offsets for parsing from byte slice
const (
	fileExtentItemGeneration    = 0
	fileExtentItemRAMBytes      = fileExtentItemGeneration + 8
	fileExtentItemCompression   = fileExtentItemRAMBytes + 8
	fileExtentItemEncryption    = fileExtentItemCompression + 1
	fileExtentItemOtherEncoding = fileExtentItemEncryption + 1
	fileExtentItemType          = fileExtentItemOtherEncoding + 2
	// At this offset in the structure, the inline extent data starts.
	// The following fields are valid only if Type != FileExtentInline:
	fileExtentItemDiskByteNr   = fileExtentItemType + 1
	fileExtentItemDiskNumBytes = fileExtentItemDiskByteNr + 8
	fileExtentItemOffset       = fileExtentItemDiskNumBytes + 8
	fileExtentItemNumBytes     = fileExtentItemOffset + 8
	FileExtentItemEnd          = fileExtentItemNumBytes + 8
)

// Transaction id that created this extent
func (i FileExtentItem) Generation() uint64 { return SliceUint64LE(i[fileExtentItemGeneration:]) }

// Max number of bytes to hold this extent in ram when we split a
// compressed extent we can't know how big each of the resulting pieces
// will be. So, this is an upper limit on the size of the extent in ram
// instead of an exact limit.
func (i FileExtentItem) RAMBytes() uint64 { return SliceUint64LE(i[fileExtentItemRAMBytes:]) }

// 32 bits for the various ways we might encode the data, including
// compression and encryption. If any of these are set to something a
// given disk format doesn't understand it is treated like an incompat
// flag for reading and writing, but not for stat.
func (i FileExtentItem) Compression() uint8 { return i[fileExtentItemCompression] }
func (i FileExtentItem) Encryption() uint8  { return i[fileExtentItemEncryption] }

// For later use
func (i FileExtentItem) OtherEncoding() uint16 { return SliceUint16LE(i[fileExtentItemOtherEncoding:]) }

// Are we inline data or a real extent?
func (i FileExtentItem) Type() uint8 { return i[fileExtentItemType] }

// Disk space consumed by the extent, checksum blocks are included in
// these numbers.

func (i FileExtentItem) DiskByteNr() uint64   { return SliceUint64LE(i[fileExtentItemDiskByteNr:]) }
func (i FileExtentItem) DiskNumBytes() uint64 { return SliceUint64LE(i[fileExtentItemDiskNumBytes:]) }

// The logical offset in file blocks (no csums) this extent record is
// for. This allows a file extent to point into the middle of an existing
// extent on disk, sharing it between two snapshots (useful if some bytes
// in the middle of the extent have changed.
func (i FileExtentItem) Offset() uint64 { return SliceUint64LE(i[fileExtentItemOffset:]) }

// The logical number of file blocks (no csums included). This always
// reflects the size uncompressed and without encoding.
func (i FileExtentItem) NumBytes() uint64 { return SliceUint64LE(i[fileExtentItemNumBytes:]) }

func (i FileExtentItem) IsInline() bool { return i.Type() == FileExtentInline }

// The data returned is only valid if Type == FileExtentInline:
func (i FileExtentItem) Data() string {
	l := int(i.RAMBytes())
	if l > DefaultBlockSize {
		l = DefaultBlockSize
	}
	return string(i[fileExtentItemDiskByteNr : fileExtentItemDiskByteNr+l])
}

type CSumItem []byte

// CSumItem offsets for parsing from byte slice
const (
	cSumItemCSum = 0
)

func (i CSumItem) CSum() CSum {
	// TODO(cblichmann): Have recon.go figure out checksum sizes, use 4
	//                   (CRC32) as default. Heuristic: Scan CSumItems and
	//                   check how many padding bytes there are.
	c := CSum{}
	copy(c[:], i[cSumItemCSum:])
	return c
}

type RootItem []byte

// RootItem offsets for parsing from byte slice
const (
	rootItemInode        = 0
	rootItemGeneration   = rootItemInode + InodeItemLen
	rootItemRootDirID    = rootItemGeneration + 8
	rootItemByteNr       = rootItemRootDirID + 8
	rootItemByteLimit    = rootItemByteNr + 8
	rootItemBytesUsed    = rootItemByteLimit + 8
	rootItemLastSnapshot = rootItemBytesUsed + 8
	rootItemFlags        = rootItemLastSnapshot + 8
	rootItemRefs         = rootItemFlags + 8
	rootItemDropProgress = rootItemRefs + 4
	rootItemDropLevel    = rootItemDropProgress + 17
	rootItemLevel        = rootItemDropLevel + 1
	// The following fields appear after subvol_uuids+subvol_times were
	// introduced.
	rootItemGenerationV2 = rootItemLevel + 1
	rootItemUUID         = rootItemGenerationV2 + 8
	rootItemParentUUID   = rootItemUUID + uuid.UUIDSize
	rootItemReceivedUUID = rootItemParentUUID + uuid.UUIDSize
	rootItemCTransID     = rootItemReceivedUUID + uuid.UUIDSize
	rootItemOTransID     = rootItemCTransID + 8
	rootItemSTransID     = rootItemOTransID + 8
	rootItemRTransID     = rootItemSTransID + 8
	rootItemCtime        = rootItemRTransID + 8
	rootItemOtime        = rootItemCtime + 12
	rootItemStime        = rootItemOtime + 12
	rootItemRtime        = rootItemStime + 12
	rootItemReserved     = rootItemRtime + 12
	RootItemLen          = rootItemReserved + 8*8
)

func (i RootItem) Inode() InodeItem     { return InodeItem(i[rootItemInode:]) }
func (i RootItem) Generation() uint64   { return SliceUint64LE(i[rootItemGeneration:]) }
func (i RootItem) RootDirID() uint64    { return SliceUint64LE(i[rootItemRootDirID:]) }
func (i RootItem) ByteNr() uint64       { return SliceUint64LE(i[rootItemByteNr:]) }
func (i RootItem) ByteLimit() uint64    { return SliceUint64LE(i[rootItemByteLimit:]) }
func (i RootItem) BytesUsed() uint64    { return SliceUint64LE(i[rootItemBytesUsed:]) }
func (i RootItem) LastSnapshot() uint64 { return SliceUint64LE(i[rootItemLastSnapshot:]) }
func (i RootItem) Flags() uint64        { return SliceUint64LE(i[rootItemFlags:]) }
func (i RootItem) Refs() uint32         { return SliceUint32LE(i[rootItemRefs:]) }
func (i RootItem) DropProgress() Key    { return SliceKey(i[rootItemDropProgress:]) }
func (i RootItem) DropLevel() uint8     { return i[rootItemDropLevel] }
func (i RootItem) Level() uint8         { return i[rootItemLevel] }

// This generation number is used to test if the new fields are valid
// and up to date while reading the root item. Everytime the root item
// is written out, the "generation" field is copied into this field. If
// anyone ever mounted the fs with an older kernel, we will have
// mismatching generation values here and thus must invalidate the
// new fields.
func (i RootItem) GenerationV2() uint64    { return SliceUint64LE(i[rootItemGenerationV2:]) }
func (i RootItem) UUID() uuid.UUID         { return SliceUUID(i[rootItemUUID:]) }
func (i RootItem) ParentUUID() uuid.UUID   { return SliceUUID(i[rootItemParentUUID:]) }
func (i RootItem) ReceivedUUID() uuid.UUID { return SliceUUID(i[rootItemReceivedUUID:]) }

// Updated when an inode changes
func (i RootItem) CTransID() uint64 { return SliceUint64LE(i[rootItemCTransID:]) }

// Trans when created
func (i RootItem) OTransID() uint64 { return SliceUint64LE(i[rootItemOTransID:]) }

// Trans when sent. Non-zero for received subvol
func (i RootItem) STransID() uint64 { return SliceUint64LE(i[rootItemSTransID:]) }

// Trans when received. Non-zero for received subvol
func (i RootItem) RTransID() uint64 { return SliceUint64LE(i[rootItemRTransID:]) }
func (i RootItem) Ctime() time.Time { return SliceTimeLE(i[rootItemCtime:]).UTC() }
func (i RootItem) Otime() time.Time { return SliceTimeLE(i[rootItemOtime:]).UTC() }
func (i RootItem) Stime() time.Time { return SliceTimeLE(i[rootItemStime:]).UTC() }
func (i RootItem) Rtime() time.Time { return SliceTimeLE(i[rootItemRtime:]).UTC() }
func (i RootItem) Reserved() [8]uint64 {
	return [8]uint64{SliceUint64LE(i[rootItemReserved:]),
		SliceUint64LE(i[rootItemReserved+8:]),
		SliceUint64LE(i[rootItemReserved+16:]),
		SliceUint64LE(i[rootItemReserved+24:]),
		SliceUint64LE(i[rootItemReserved+32:]),
		SliceUint64LE(i[rootItemReserved+40:]),
		SliceUint64LE(i[rootItemReserved+48:]),
		SliceUint64LE(i[rootItemReserved+56:]),
	}
}

func (i RootItem) IsGenerationV2() bool { return i.Generation() == i.GenerationV2() }

// This is used for both forward and backward root refs
type RootRef []byte

// RootRef offsets for parsing from byte slice
const (
	rootRefDirID    = 0
	rootRefSequence = rootRefDirID + 8
	rootRefNameLen  = rootRefSequence + 8
	rootRefName     = rootRefNameLen + 2
)

func (r RootRef) DirID() uint64    { return SliceUint64LE(r[rootRefDirID:]) }
func (r RootRef) Sequence() uint64 { return SliceUint64LE(r[rootRefSequence:]) }
func (r RootRef) NameLen() uint16  { return SliceUint16LE(r[rootRefNameLen:]) }

func (r RootRef) Name() string {
	l := int(r.NameLen())
	if l > 255 {
		l = 255
	}
	return string(r[rootRefName : rootRefName+l])
}

// Extent item flags
const (
	ExtentFlagData = 1 << iota
	ExtentFlagTreeBlock
	ExtentFlagFullBackref = 0x80
)

// Items in the extent btree are used to record the objectid of the
// owner of the block and the number of references.
type ExtentItem []byte

// ExtentItem offsets for parsing from byte slice
const (
	extentItemRefs       = 0
	extentItemGeneration = extentItemRefs + 8
	extentItemFlags      = extentItemGeneration + 8
	extentItemEnd        = extentItemFlags + 8
)

func (i ExtentItem) Refs() uint64       { return SliceUint64LE(i[extentItemRefs:]) }
func (i ExtentItem) Generation() uint64 { return SliceUint64LE(i[extentItemGeneration:]) }
func (i ExtentItem) Flags() uint64      { return SliceUint64LE(i[extentItemFlags:]) }

// TODO(cblichmann): btrfs_tree_block_info when Flags == ExtentFlagFullBackref

func (i ExtentItem) IsCompatV0() bool { return len(i) < 8+8+8 }
