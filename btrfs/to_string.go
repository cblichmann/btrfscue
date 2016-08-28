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

func ObjectIDString(id uint64) string {
	switch id {
	case RootTreeObjectId:
		return "ROOT_TREE"
	case ExtentTreeObjectId:
		return "EXTENT_TREE"
	case ChunkTreeObjectId:
		return "CHUNK_TREE"
	case DevTreeObjectId:
		return "DEV_TREE"
	case FSTreeObjectId:
		return "FS_TREE"
	case RootTreeDirObjectId:
		return "ROOT_TREE_DIR"
	case CSumTreeObjectId:
		return "CSUM_TREE"
	case QuotaTreeObjectId:
		return "QUOTA_TREE"
	case UuidTreeObjectId:
		return "UUID"
	case FreeSpaceTreeObjectId:
		return "FREES_SPACE_TREE"
	case DevStatsObjectId:
		return "DEV_STATS"
	case BalanceObjectId:
		return "BALANCE"
	case OrphanObjectId:
		return "ORPHAN"
	case TreeLogObjectId:
		return "TREE_LOG"
	case TreeLogFixupObjectId:
		return "TREE_LOG_FIXUP"
	case TreeRelocObjectId:
		return "TREE_RELOC"
	case DataRelocTreeObjectId:
		return "DATA_RELOC_TREE"
	case ExtentCSumObjectId:
		return "EXTENT_CSUM"
	case FreeSpaceObjectId:
		return "FREE_SPACE"
	case FreeInoObjectId:
		return "FREE_INO"
	case MultipleObjectIds:
		return "MULTIPLES"
	case FirstFreeObjectId:
		return "FIRST_FREE"
	case LastFreeObjectId:
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
	return fmt.Sprintf("key (%s %s %d=%#[3]x)", ObjectIDString(k.ObjectID),
		KeyTypeString(k.Type), k.Offset)
}
