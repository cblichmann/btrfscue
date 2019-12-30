/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * BTRFS filesystem index data structure
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

package index // import "blichmann.eu/code/btrfscue/btrfs/index"

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"sort"
	"strings"

	"fmt"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/uuid"
	"github.com/etcd-io/bbolt"
)

func init() { fmt.Printf("") } //DBG!!!

// Fake FS key used for metadata. Also acts as a sentinel value for
// non-existing FS objects. Since it always sorts lexicographically last,
// Seek() on the "index" bucket will never return nil.
var metadataKey = newIndexKey(^uint64(0), KL(), ^uint64(0))

const (
	// Index metadata version. Set to ISO date (decimal) whenever there are
	// incompatible changes.
	MetadataVersion           = 20190809 // V2: Put generation first (decending) in key
	MetadataVersionUpgradable = 20161109 // V1: Orignal format using Boltdb
)

type stripeMapEntry struct {
	logical uint64
	devID   uint64
	offset  uint64
}

// Index encapsulates metadata of a BTRFS to be recovered/analyzed. It uses
// a memory-mapped key-value store to quickly access FS objects.
// When opened read/write, concurrent access to this object must be guarded.
type Index struct {
	db     *bbolt.DB
	tx     *bbolt.Tx
	bucket *bbolt.Bucket

	// Number of inserts that are in flight in the current transaction
	txNum      int
	Generation uint64

	stripeMap []stripeMapEntry
}

// Options sets options for opening a metadata index.
type Options struct {
	ReadOnly        bool
	BlockSize       uint
	FSID            uuid.UUID
	Generation      uint64
	AllowOldVersion bool
}

type indexMetadata []byte

// Offsets for parsing from byte slice
const (
	indexMetadataVersion    = 0
	indexMetadataBlockSize  = indexMetadataVersion + 8
	indexMetadataFSID       = indexMetadataBlockSize + 4
	indexMetadataGeneration = indexMetadataFSID + uuid.UUIDSize
	indexMetadataEnd        = indexMetadataGeneration + 8
)

func newIndexMetadata(o *Options) indexMetadata {
	m := [indexMetadataEnd]byte{}
	binary.LittleEndian.PutUint64(m[indexMetadataVersion:], MetadataVersion)
	binary.LittleEndian.PutUint32(m[indexMetadataBlockSize:], uint32(
		o.BlockSize))
	copy(m[indexMetadataFSID:], o.FSID[:])
	binary.LittleEndian.PutUint64(m[indexMetadataGeneration:], o.Generation)
	return m[:]
}

func (m indexMetadata) Version() uint64    { return btrfs.SliceUint64LE(m[indexMetadataVersion:]) }
func (m indexMetadata) BlockSize() uint32  { return btrfs.SliceUint32LE(m[indexMetadataBlockSize:]) }
func (m indexMetadata) FSID() uuid.UUID    { return btrfs.SliceUUID(m[indexMetadataFSID:]) }
func (m indexMetadata) Generation() uint64 { return btrfs.SliceUint64LE(m[indexMetadataGeneration:]) }

// Open opens a metadata index with the specified options.
func Open(path string, m os.FileMode, o *Options) (*Index, error) {
	return openIndex(path, m, o)
}

// OpenReadOnly opens a metadata index for reading/querying.
func OpenReadOnly(path string) (*Index, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return openIndex(path, 0644 /* Mode */, nil)
}

func openIndex(path string, m os.FileMode, o *Options) (*Index, error) {
	if o == nil {
		o = &Options{ReadOnly: true, Generation: ^uint64(0)}
	}
	ix := &Index{Generation: o.Generation}
	var err error
	if ix.db, err = bbolt.Open(path, m, &bbolt.Options{
		ReadOnly: o.ReadOnly}); err != nil {
		return nil, err
	}
	if err = ix.checkUpdateMetadata(o); err != nil {
		return nil, err
	}
	if err = ix.ensureTx(!o.ReadOnly); err != nil {
		return nil, err
	}
	return ix, nil
}

func (ix *Index) checkUpdateMetadata(o *Options) error {
	fn := ix.db.Update
	if o.ReadOnly {
		fn = ix.db.View
	}
	return fn(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("index"))
		if err != nil {
			return err
		}
		m := indexMetadata(bucket.Get(metadataKey))
		if m == nil {
			if o.ReadOnly {
				return errors.New("no index in metadata")
			}
			// No index yet, set current
			if err = bucket.Put(metadataKey, newIndexMetadata(o)); err != nil {
				return err
			}
			return nil
		}

		if !o.AllowOldVersion && m.Version() < MetadataVersion {
			return fmt.Errorf("metadata version v%d too old, upgrade using "+
				"upgrade-index", m.Version())
		}
		if m.Version() > MetadataVersion {
			return fmt.Errorf("incompatible metadata, expected v%d got: v%d",
				MetadataVersion, m.Version())
		}
		if o.AllowOldVersion {
			// Skip other checks if we're upgrading
			return nil
		}
		if m.BlockSize() != uint32(o.BlockSize) {
			return fmt.Errorf("block size mismatch, expected %d got: %d",
				o.BlockSize, m.BlockSize())
		}
		if m.FSID() != o.FSID {
			return fmt.Errorf("filesystem id mismatch, expected %s", o.FSID)
		}
		return nil
	})
}

// Close closes the index and its underlying bbolt database.
func (ix *Index) Close() {
	// Make sure writable and read-only transactions are completed.
	if !ix.db.IsReadOnly() {
		ix.Commit()
	} else {
		ix.tx.Rollback()
	}
	ix.db.Close()
}

func (ix *Index) ensureTx(writable bool) error {
	var err error
	if ix.tx != nil {
		if ix.tx.Writable() == writable {
			return nil
		}
		if err = ix.Commit(); err != nil {
			return err
		}
	}
	if ix.tx, err = ix.db.Begin(writable); err != nil {
		return err
	}
	ix.bucket = ix.tx.Bucket([]byte("index"))
	return err
}

// InsertItem inserts a filesystem item into the index, referenceable by its
// BTRFS key. Also stores the item's owner and generation number, as well as
// its inline data.
func (ix *Index) InsertItem(k btrfs.Key, h btrfs.Header, item,
	data []byte) error {
	if err := ix.ensureTx(true); err != nil {
		return err
	}
	l := btrfs.ItemLen + len(data)
	tc := make([]byte, l)
	copy(tc, item)
	copy(tc[btrfs.ItemLen:], data)
	if err := ix.bucket.Put(newIndexKey(h.Owner(), k, h.Generation()),
		tc); err != nil {
		return err
	}
	ix.txNum++
	if ix.txNum > 10000 {
		return ix.Commit()
	}
	return nil
}

// Commit commits any pending transaction.
func (ix *Index) Commit() error {
	if ix.tx == nil {
		return nil
	}
	err := ix.tx.Commit()
	ix.tx = nil
	ix.txNum = 0
	ix.bucket = nil
	return err
}

// RawTx executes a transaction function on underlying BoltDB database. This
// is an advanced feature that should not be called in regular operation. It
// is used by the upgrade-index sub-command, for example, to upgrade legacy
// metadata indices.
// This function requires that the index was opened read-write. It behaves
// similar to bbolt.Update() otherwise.
func (ix *Index) RawTx(fn func(db *bbolt.DB, tx *bbolt.Tx,
	bucket *bbolt.Bucket) error) error {
	if err := ix.ensureTx(true); err != nil {
		return err
	}
	if err := fn(ix.db, ix.tx, ix.bucket); err != nil {
		return err
	}
	return ix.Commit()
}

// lowerBound finds an FS key under a given owner and only up to a prefix
// length. It find the key with the highest generation number smaller than or
// equal to the index generation.
// For example, to search for (256 DIR_INDEX ?) owned by the FS tree object:
//   lowerBound(FSTreeObjectID, KF(DIR_INDEX, 256), keyV2Offset)
func lowerBound(c *bbolt.Cursor, owner uint64, k btrfs.Key, gen uint64,
	prefix int) btrfs.Key {
	var cur, next keyV2
	search := newIndexKey(owner, k, 0)[:prefix]
	for cur, _ = c.Seek(search); bytes.Equal(cur[:prefix], search); cur = next {
		next, _ = c.Next()
		if bytes.Equal(next[:prefix], search) && next.Generation() >= gen {
			return cur.Key()
		}
	}
	return KL()
}

func find(c *bbolt.Cursor, owner uint64, k btrfs.Key, generation uint64) (
	keyV2, btrfs.Item) {
	search := newIndexKey(owner, k, generation)
	var found keyV2
	found, v := c.Seek(search)
	if bytes.Compare(found[:keyV2Generation],
		search[:keyV2Generation]) <= 0 {
		return found, v
	}
	if found, v = c.Prev(); len(found) > 0 && bytes.Equal(
		found[:keyV2Generation], search[:keyV2Generation]) {
		return found, v
	}
	return nil, nil
}

func findNext(c *bbolt.Cursor, ik, end keyV2) (keyV2, btrfs.Item) {
	search := newIndexKey(ik.Owner(), ik.Key(), ^uint64(0))
	var found keyV2
	if found, _ = c.Seek(search); found != nil {
		search = newIndexKey(ik.Owner(), found.Key(), ik.Generation())
		var v btrfs.Item
		if found, v = c.Seek(search); bytes.Compare(found, search) <= 0 {
			return found, v
		} else if end != nil && bytes.Compare(found, end) <= 0 {
			return found, v
		}
	}
	return nil, nil
}

// Range encapsulates a generic index range. Internally, it holds a cursor of
// the underlying key-value store as well as the current position and the end
// of range marker.
type Range struct {
	ix       *Index
	cursor   *bbolt.Cursor
	key, end keyV2
	value    btrfs.Item
}

// Index returns the Index that this Range refers to.
func (r *Range) Index() *Index { return r.ix }

func (r *Range) HasNext() bool {
	return r.key != nil && bytes.Compare(r.key, r.end) <= 0
}

func (r *Range) Next() []byte {
	if r.key, r.value = findNext(r.cursor, r.key, nil); r.key != nil {
		return r.value.Data()
	}
	return nil
}

func (r *Range) Owner() uint64      { return r.key.Owner() }
func (r *Range) Key() btrfs.Key     { return r.key.Key() }
func (r *Range) Generation() uint64 { return r.key.Generation() }
func (r *Range) Item() btrfs.Item   { return r.value }

// Range returns an index range [first, last) for the given keys.
func (ix *Index) Range(owner uint64, first, last btrfs.Key) (Range, []byte) {
	r := Range{
		ix:     ix,
		cursor: ix.bucket.Cursor(),
		end:    newIndexKey(owner, last, ix.Generation),
	}
	lowerFirst := lowerBound(r.cursor, owner, first, ix.Generation,
		keyV2Offset)
	r.key, r.value = find(r.cursor, owner, lowerFirst, ix.Generation)
	if r.key != nil {
		return r, r.value.Data()
	}
	return r, nil
}

// RangeAll returns an index range for all items related to the key (t id ?).
func (ix *Index) RangeAll(owner uint64, t uint8, id uint64) (Range, []byte) {
	return ix.Range(owner, KF(uint64(t), id), KL(uint64(t), id))
}

type FullRange struct {
	Range
}

func (r *FullRange) Next() []byte {
	if r.key, r.value = r.cursor.Next(); r.key != nil && !bytes.Equal(r.key,
		r.end) {
		return r.value.Data()
	}
	return nil
}

// FullRange returns an index range for all items in the index regardless of
// generation. This is mainly useful for debugging.
func (ix *Index) FullRange() (FullRange, []byte) {
	r := FullRange{Range{
		ix:     ix,
		cursor: ix.bucket.Cursor(),
		end:    newIndexKey(^uint64(0), KL(), ix.Generation),
	}}
	// No check, since openIndex should fail if there's no data
	r.key, r.value = r.cursor.First()
	return r, r.value.Data()
}

// Subvolumes returns an index range containing all of the subvolumes. Note that
// this excludes the FS tree root itself, which is not considered to be a
// subvolume. To access the FS tree root, use
//   Find(KF(RootItemKey, FSTreeObjectID))
func (ix *Index) Subvolumes() (Range, btrfs.RootItem) {
	return ix.Range(btrfs.RootTreeObjectID,
		KF(btrfs.RootItemKey, btrfs.FirstFreeObjectID),
		KL(btrfs.RootItemKey, btrfs.LastFreeObjectID))
}

func (ix *Index) DirItems(owner, id uint64) (Range, btrfs.DirItem) {
	return ix.RangeAll(owner, btrfs.DirItemKey, id)
}

func (ix *Index) XAttrItems(owner, id uint64) (Range, btrfs.DirItem) {
	return ix.RangeAll(owner, btrfs.XAttrItemKey, id)
}

func (ix *Index) FileExtentItems(owner, id uint64) (Range,
	btrfs.FileExtentItem) {
	return ix.RangeAll(owner, btrfs.ExtentDataKey, id)
}

// Chunks returns an index range containing all of the chunk items.
func (ix *Index) Chunks() (Range, btrfs.Chunk) {
	return ix.RangeAll(btrfs.ChunkTreeObjectID, btrfs.ChunkItemKey,
		btrfs.FirstFreeObjectID)
}

// DevItems returns an index range containing all of the device items.
func (ix *Index) DevItems() (Range, btrfs.DevItem) {
	return ix.RangeAll(btrfs.ChunkTreeObjectID, btrfs.DevItemKey,
		btrfs.DevItemsObjectID)
}

// FindItem searches for an FS key at the latest generation smaller or equal
// to the current index generation. If the index generation is smaller than
// any existing generation, the data at the earliest generatation is returned.
func (ix *Index) FindItem(owner uint64, k btrfs.Key) btrfs.Item {
	_, i := find(ix.bucket.Cursor(), owner, k, ix.Generation)
	return i
}

func (ix *Index) FindInodeItem(owner uint64, inode uint64) btrfs.InodeItem {
	if i := ix.FindItem(owner, KF(btrfs.InodeItemKey, inode)); i != nil {
		return i.Data()
	}
	return nil
}

func (ix *Index) FindFileExtentItem(owner uint64, id uint64) btrfs.FileExtentItem {
	if i := ix.FindItem(owner, KF(btrfs.ExtentDataKey, id)); i != nil {
		return i.Data()
	}
	return nil
}

func (ix *Index) FindExtentItem(diskByteNr, diskNumBytes uint64) btrfs.ExtentItem {
	if i := ix.FindItem(btrfs.ExtentTreeObjectID,
		KF(btrfs.ExtentItemKey, diskByteNr, diskNumBytes)); i != nil {
		return i.Data()
	}
	return nil
}

// Physical maps a filesystem logical address to a physical, on-disk address.
// Note: No attempt is made to deal with "holes", missing chunk entries, or
// logical addresses not mapping to any chunk.
func (ix *Index) Physical(logical uint64) (devID uint64, offset uint64) {
	if ix.stripeMap == nil {
		ix.stripeMap = make([]stripeMapEntry, 0, 5 /* Initial capacity */)
		for r, c := ix.Chunks(); r.HasNext(); c = r.Next() {
			if c.NumStripes() == 0 {
				// Invalid chunk, should always have at least one stripe
				continue
			}
			s := c.Stripe(0)
			ix.stripeMap = append(ix.stripeMap, stripeMapEntry{
				logical: r.Key().Offset,
				devID:   s.DevID(),
				offset:  s.Offset()})
			fmt.Printf("stripe: %d => %d @ %d\n", r.Key().Offset, s.Offset(), r.Generation())
		}
	}

	stripeMapLen := len(ix.stripeMap)
	// Index structure entries are sorted
	i := sort.Search(stripeMapLen, func(i int) bool {
		return ix.stripeMap[i].logical >= logical
	})
	if i == stripeMapLen || (i > 0 && ix.stripeMap[i].logical != logical) {
		// Not found, adjust index value, as it's an insertion point (see
		// documentation for sort.Search()). Unless i was 0, in which case the
		// position is fine.
		i--
	}
	s := ix.stripeMap[i]
	// TODO(cblichmann): We can check for the failure case if we store the
	//                   stripe length in the cache and do a range check here.
	return s.devID, s.offset + logical - s.logical
}

// FindDirItemForPath finds the DirItem for a given FS path. It assumes that
// the path is clean, as returned by path.Clean().
func (ix *Index) FindDirItemForPath(owner uint64, path string) btrfs.DirItem {
	result := btrfs.DirItem(nil)
	dirID := uint64(btrfs.FirstFreeObjectID)
	pathComps := strings.Split(strings.TrimLeft(path, "/"), "/")
	for i, comp := range pathComps {
		found := false
		last := i == len(pathComps)-1
		var d btrfs.DirItem
		// Fast path: Try to find item by name hash search first:
		// (dirID, BTRFS_DIR_ITEM, CRC32(comp))
		if v := ix.FindItem(owner, KF(btrfs.DirItemKey, dirID,
			uint64(NameHash(comp)))); v != nil {
			d = v.Data()
			if found = d.Name() == comp; found {
				result = d
			}
		}
		if !found {
			// Slow path: Lookup items in directory (dirID, BTRFS_DIR_INDEX, ?)
			var r Range
			for r, d = ix.DirItems(owner, dirID); r.HasNext(); d = r.Next() {
				if found = d.Name() == comp && (d.IsDir() || last); found {
					result = d
					break
				}
			}
		}

		if !found {
			return nil
		}
		if !d.IsSubvolume() {
			dirID = d.Location().ObjectID
		} else {
			owner = d.Location().ObjectID
			dirID = btrfs.FirstFreeObjectID
		}
	}
	return result
}

func keyFirst(k *btrfs.Key, args []uint64) {
	switch len(args) {
	case 3:
		k.Offset = args[2]
		fallthrough
	case 2:
		k.ObjectID = args[1]
		fallthrough
	case 1:
		k.Type = uint8(args[0])
	case 0:
	default:
		panic("invalid number of arguments")
	}
}

// KF returns a new key useful for searching an index. To enable more natural
// queries, the order of the arguments is different from the one in the Key
// struct: Type, ObjectID, Offset.
func KF(args ...uint64) (k btrfs.Key) {
	keyFirst(&k, args)
	return
}

func KL(args ...uint64) btrfs.Key {
	k := btrfs.Key{ObjectID: btrfs.LastFreeObjectID,
		Type: ^uint8(0), Offset: ^uint64(0)}
	keyFirst(&k, args)
	return k
}

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// NameHash computes the CRC32C value of an FS object name. The initial CRC
// in BTRFS is ^uint32(1) and the end result is _not_ inverted.
func NameHash(name string) uint32 {
	return ^crc32.Update(^uint32(0xFFFFFFFE), castagnoliTable, []byte(name))
}
