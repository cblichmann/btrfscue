/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
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

package btrfs // import "blichmann.eu/code/btrfscue/btrfs"

import (
	"fmt"
)

func ObjectIdString(id uint64) string {
	switch id {
	case RootTreeObjectId:
		return "BTRFS_ROOT_TREE_OBJECTID"
	case ExtentTreeObjectId:
		return "BTRFS_EXTENT_TREE_OBJECTID"
	case ChunkTreeObjectId:
		return "BTRFS_CHUNK_TREE_OBJECTID"
	case DevTreeObjectId:
		return "BTRFS_DEV_TREE_OBJECTID"
	case FSTreeObjectId:
		return "BTRFS_FS_TREE_OBJECTID"
	case RootTreeDirObjectId:
		return "BTRFS_ROOT_TREE_DIR_OBJECTID"
	case CSumTreeObjectId:
		return "BTRFS_CSUM_TREE_OBJECTID"
	case OrphanObjectId:
		return "BTRFS_ORPHAN_OBJECTID"
	case TreeLogObjectId:
		return "BTRFS_TREE_LOG_OBJECTID"
	case TreeLogFixupObjectId:
		return "BTRFS_TREE_LOG_FIXUP_OBJECTID"
	case TreeRelocObjectId:
		return "BTRFS_TREE_RELOC_OBJECTID"
	case DataRelocTreeObjectId:
		return "BTRFS_DATA_RELOC_TREE_OBJECTID"
	case ExtentCSumObjectId:
		return "BTRFS_EXTENT_CSUM_OBJECTID"
	case FreeSpaceObjectId:
		return "BTRFS_FREE_SPACE_OBJECTID"
	case MultipleObjectIdS:
		return "BTRFS_MULTIPLE_OBJECTIDS"
	case FirstFreeObjectId:
		return "BTRFS_FIRST_FREE_OBJECTID"
	case LastFreeObjectId:
		return "BTRFS_LAST_FREE_OBJECTID"
	default:
		return fmt.Sprint(id)
	}
}

func KeyTypeString(t uint8) string {
	switch t {
	case InodeItemKey:
		return "BTRFS_INODE_ITEM_KEY"
	case InodeRefKey:
		return "BTRFS_INODE_REF_KEY"
	case InodeExtrefKey:
		return "BTRFS_INODE_EXTREF_KEY"
	case XAttrItemKey:
		return "BTRFS_XATTR_ITEM_KEY"
	case OrphanItemKey:
		return "BTRFS_ORPHAN_ITEM_KEY"
	case DirLogItemKey:
		return "BTRFS_DIR_LOG_ITEM_KEY"
	case DirLogIndexKey:
		return "BTRFS_DIR_LOG_INDEX_KEY"
	case DirItemKey:
		return "BTRFS_DIR_ITEM_KEY"
	case DirIndexKey:
		return "BTRFS_DIR_INDEX_KEY"
	case ExtentDataKey:
		return "BTRFS_EXTENT_DATA_KEY"
	case ExtentCSumKey:
		return "BTRFS_EXTENT_CSUM_KEY"
	case RootItemKey:
		return "BTRFS_ROOT_ITEM_KEY"
	case RootBackRefKey:
		return "BTRFS_ROOT_BACKREF_KEY"
	case RootRefKey:
		return "BTRFS_ROOT_REF_KEY"
	case ExtentItemKey:
		return "BTRFS_EXTENT_ITEM_KEY"
	case MetadataItemKey:
		return "BTRFS_METADATA_ITEM_KEY"
	case TreeBlockRefKey:
		return "BTRFS_TREE_BLOCK_REF_KEY"
	case ExtentDataRefKey:
		return "BTRFS_EXTENT_DATA_REF_KEY"
	case ExtentRefV0Key:
		return "BTRFS_EXTENT_REF_V0_KEY"
	case SharedBlockRefKey:
		return "BTRFS_SHARED_BLOCK_REF_KEY"
	case SharedDataRefKey:
		return "BTRFS_SHARED_DATA_REF_KEY"
	case BlockGroupItemKey:
		return "BTRFS_BLOCK_GROUP_ITEM_KEY"
	case FreeSpaceInfoKey:
		return "BTRFS_FREE_SPACE_INFO_KEY"
	case FreeSpaceExtentKey:
		return "BTRFS_FREE_SPACE_EXTENT_KEY"
	case FreeSpaceBitmapKey:
		return "BTRFS_FREE_SPACE_BITMAP_KEY"
	case DevExtentKey:
		return "BTRFS_DEV_EXTENT_KEY"
	case DevItemKey:
		return "BTRFS_DEV_ITEM_KEY"
	case ChunkItemKey:
		return "BTRFS_CHUNK_ITEM_KEY"
	case QgroupStatusKey:
		return "BTRFS_QGROUP_STATUS_KEY"
	case QgroupInfoKey:
		return "BTRFS_QGROUP_INFO_KEY"
	case QgroupRelationKey:
		return "BTRFS_QGROUP_RELATION_KEY"
	case BalanceItemKey:
		return "BTRFS_BALANCE_ITEM_KEY"
	case DevStatsKey:
		return "BTRFS_DEV_STATS_KEY"
	case DevReplaceKey:
		return "BTRFS_DEV_REPLACE_KEY"
	case StringItemKey:
		return "BTRFS_STRING_ITEM_KEY"
	default:
		return fmt.Sprint(t)
	}
}

func (k Key) String() string {
	return fmt.Sprintf("key (%s %s %d=%#[3]x)", ObjectIdString(k.ObjectId),
		KeyTypeString(k.Type), k.Offset)
}
