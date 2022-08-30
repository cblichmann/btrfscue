/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
 *
 * Human readable output for BTRFS filesystem structures
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
	"fmt"
)

func ObjectIDString(id uint64) string {
	switch id {
	case RootTreeObjectID:
		return "ROOT_TREE"
	case ExtentTreeObjectID:
		return "EXTENT_TREE"
	case ChunkTreeObjectID:
		return "CHUNK_TREE"
	case DevTreeObjectID:
		return "DEV_TREE"
	case FSTreeObjectID:
		return "FS_TREE"
	case RootTreeDirObjectID:
		return "ROOT_TREE_DIR"
	case CSumTreeObjectID:
		return "CSUM_TREE"
	case QuotaTreeObjectID:
		return "QUOTA_TREE"
	case UuidTreeObjectID:
		return "UUID"
	case FreeSpaceTreeObjectID:
		return "FREES_SPACE_TREE"
	case DevStatsObjectID:
		return "DEV_STATS"
	case BalanceObjectID:
		return "BALANCE"
	case OrphanObjectID:
		return "ORPHAN"
	case TreeLogObjectID:
		return "TREE_LOG"
	case TreeLogFixupObjectID:
		return "TREE_LOG_FIXUP"
	case TreeRelocObjectID:
		return "TREE_RELOC"
	case DataRelocTreeObjectID:
		return "DATA_RELOC_TREE"
	case ExtentCSumObjectID:
		return "EXTENT_CSUM"
	case FreeSpaceObjectID:
		return "FREE_SPACE"
	case FreeInoObjectID:
		return "FREE_INO"
	case MultipleObjectIDs:
		return "MULTIPLE"
	case FirstFreeObjectID:
		return "FIRST_FREE"
	case LastFreeObjectID:
		return "LAST_FREE"
	default:
		return fmt.Sprint(id)
	}
}

func KeyTypeString(t uint8) string {
	switch t {
	case InodeItemKey:
		return "INODE_ITEM"
	case InodeRefKey:
		return "INODE_REF"
	case InodeExtrefKey:
		return "INODE_EXTREF"
	case XAttrItemKey:
		return "XATTR_ITEM"
	case OrphanItemKey:
		return "ORPHAN_ITEM"
	case DirLogItemKey:
		return "DIR_LOG_ITEM"
	case DirLogIndexKey:
		return "DIR_LOG_INDEX"
	case DirItemKey:
		return "DIR_ITEM"
	case DirIndexKey:
		return "DIR_INDEX"
	case ExtentDataKey:
		return "EXTENT_DATA"
	case ExtentCSumKey:
		return "EXTENT_CSUM"
	case RootItemKey:
		return "ROOT_ITEM"
	case RootBackRefKey:
		return "ROOT_BACKREF"
	case RootRefKey:
		return "ROOT_REF"
	case ExtentItemKey:
		return "EXTENT_ITEM"
	case MetadataItemKey:
		return "METADATA_ITEM"
	case TreeBlockRefKey:
		return "TREE_BLOCK_REF"
	case ExtentDataRefKey:
		return "EXTENT_DATA_REF"
	case ExtentRefV0Key:
		return "EXTENT_REF_V0"
	case SharedBlockRefKey:
		return "SHARED_BLOCK_REF"
	case SharedDataRefKey:
		return "SHARED_DATA_REF"
	case BlockGroupItemKey:
		return "BLOCK_GROUP_ITEM"
	case FreeSpaceInfoKey:
		return "FREE_SPACE_INFO"
	case FreeSpaceExtentKey:
		return "FREE_SPACE_EXTENT"
	case FreeSpaceBitmapKey:
		return "FREE_SPACE_BITMAP"
	case DevExtentKey:
		return "DEV_EXTENT"
	case DevItemKey:
		return "DEV_ITEM"
	case ChunkItemKey:
		return "CHUNK_ITEM"
	case QgroupStatusKey:
		return "QGROUP_STATUS"
	case QgroupInfoKey:
		return "QGROUP_INFO"
	case QgroupRelationKey:
		return "QGROUP_RELATION"
	case TemporaryItemKey:
		return "TEMPORARY_ITEM"
	case PersistentItemKey:
		return "PERSISTENT_ITEM"
	case DevReplaceKey:
		return "DEV_REPLACE"
	case UUIDKeySubvol:
		return "UUID_SUBVOL"
	case UUIDKeyReceivedSubvol:
		return "UUID_RECEIVED_SUBVOL"
	case StringItemKey:
		return "STRING_ITEM"
	default:
		return fmt.Sprint(t)
	}
}

func (k Key) String() string {
	// %d=%#[3]x
	return fmt.Sprintf("key (%s %s %d)", ObjectIDString(k.ObjectID),
		KeyTypeString(k.Type), int64(k.Offset))
}
