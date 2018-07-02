/*
 * btrfscue version 0.5
 * Copyright (c)2011-2018 Christian Blichmann
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
	"hash/crc32"
	"os"
	"strings"

	"fmt" //DBG!!!

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/uuid"
	"github.com/coreos/bbolt"
)

func init() { fmt.Printf("") } //DBG!!!

// indexKey holds the BTRFS key of an owned FS object as well as its
// generation. The tuple (owner, Key, generation) is encoded in big endian for
// lexicographical comparison.
type indexKey []byte

// Offsets for parsing from byte slice
const (
	indexKeyOwner      = 0
	indexKeyType       = indexKeyOwner + 8
	indexKeyObjectID   = indexKeyType + 1
	indexKeyOffset     = indexKeyObjectID + 8
	indexKeyGeneration = indexKeyOffset + 8
	indexKeyEnd        = indexKeyGeneration + 8
)

func newIndexKey(owner uint64, k btrfs.Key, generation uint64) indexKey {
	ik := [indexKeyEnd]byte{}
	binary.BigEndian.PutUint64(ik[indexKeyOwner:], owner)
	ik[indexKeyType] = k.Type
	binary.BigEndian.PutUint64(ik[indexKeyObjectID:], k.ObjectID)
	binary.BigEndian.PutUint64(ik[indexKeyOffset:], k.Offset)
	binary.BigEndian.PutUint64(ik[indexKeyGeneration:], generation)
	return ik[:]
}

func (ik indexKey) Owner() uint64      { return binary.BigEndian.Uint64(ik[indexKeyOwner:]) }
func (ik indexKey) Type() uint8        { return uint8(ik[indexKeyType]) }
func (ik indexKey) ObjectID() uint64   { return binary.BigEndian.Uint64(ik[indexKeyObjectID:]) }
func (ik indexKey) Offset() uint64     { return binary.BigEndian.Uint64(ik[indexKeyOffset:]) }
func (ik indexKey) Generation() uint64 { return binary.BigEndian.Uint64(ik[indexKeyGeneration:]) }
func (ik indexKey) Key() btrfs.Key {
	return btrfs.Key{ObjectID: ik.ObjectID(), Type: ik.Type(), Offset: ik.Offset()}
}

// Fake FS key used for metadata. Also acts as a sentinel value for
// non-existing FS objects. Since it always sorts lexicographically last,
// Seek() on the "index" bucket will never return nil.
var metadataKey = newIndexKey(^uint64(0), KL(), ^uint64(0))

// Index metadata version. Set to ISO date (decimal) whenever there are
// incompatible changes.
const metadataVersion = 20161109

// Index encapsulates metadata of a BTRFS to be recovered/analyzed. It uses
// a memory-mapped key-value store to quickly access FS objects.
// When opened read/write, concurrent access to this object must be guarded.
type Index struct {
	db     *bolt.DB
	tx     *bolt.Tx
	bucket *bolt.Bucket

	// Number of inserts that are in flight in the current transaction
	txNum      int
	Generation uint64
}

// IndexOptions set options for opening a metdata index.
type IndexOptions struct {
	ReadOnly   bool
	BlockSize  uint
	FSID       uuid.UUID
	Generation uint64
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

func newIndexMetadata(o *IndexOptions) indexMetadata {
	m := [indexMetadataEnd]byte{}
	binary.LittleEndian.PutUint64(m[indexMetadataVersion:], metadataVersion)
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
func Open(path string, m os.FileMode, o *IndexOptions) (*Index, error) {
	return open(path, m, o)
}

// OpenIndexReadOnly opens a metadata index for reading/querying.
func OpenReadOnly(path string) (*Index, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return open(path, 0644 /* Mode */, nil)
}

//
func open(path string, m os.FileMode, o *IndexOptions) (*Index, error) {
	if o == nil {
		o = &IndexOptions{ReadOnly: true, Generation: ^uint64(0)}
	}
	ix := &Index{Generation: o.Generation}
	var err error
	if ix.db, err = bolt.Open(path, m, &bolt.Options{
		ReadOnly: o.ReadOnly}); err != nil {
		return nil, err
	}
	if !o.ReadOnly {
		if err = ix.checkUpdateMetadata(o); err != nil {
			return nil, err
		}
	}
	if err = ix.ensureTx(!o.ReadOnly); err != nil {
		return nil, err
	}
	return ix, nil
}

func (ix *Index) checkUpdateMetadata(o *IndexOptions) error {
	return ix.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("index"))
		if err != nil {
			return err
		}
		m := indexMetadata(bucket.Get(metadataKey))
		if m == nil {
			// No metadata yet, set current
			if err = bucket.Put(metadataKey, newIndexMetadata(o)); err != nil {
				return err
			}
			return nil
		}

		if m.Version() > metadataVersion {
			return fmt.Errorf("incompatible metadata, expected v%d got: v%d",
				metadataVersion, m.Version())
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

func (ix *Index) Close() { ix.db.Close() }

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

// lowerBound finds an FS key under a given owner and only up to a prefix
// length. This ignores the generation number.
// For example, to search for (256 DIR_INDEX ?) owned by the FS tree object:
//   lowerBound(FSTreeObjectID, KF(DIR_INDEX, 256), indexKeyOffset)
func lowerBound(c *bolt.Cursor, owner uint64, k btrfs.Key,
	prefix int) btrfs.Key {
	var found indexKey
	search := newIndexKey(owner, k, 0)
	if found, _ = c.Seek(search[:prefix]); bytes.Equal(found[:prefix],
		search[:prefix]) {
		return found.Key()
	}
	return KL()
}

func find(c *bolt.Cursor, owner uint64, k btrfs.Key, generation uint64) (
	indexKey, btrfs.Item) {
	search := newIndexKey(owner, k, generation)
	var found indexKey
	found, v := c.Seek(search)
	if bytes.Compare(found[:indexKeyGeneration],
		search[:indexKeyGeneration]) <= 0 {
		return found, v
	}
	if found, v = c.Prev(); len(found) > 0 && bytes.Equal(
		found[:indexKeyGeneration], search[:indexKeyGeneration]) {
		return found, v
	}
	return nil, nil
}

func findNext(c *bolt.Cursor, ik, end indexKey) (indexKey, btrfs.Item) {
	search := newIndexKey(ik.Owner(), ik.Key(), ^uint64(0))
	var found indexKey
	if found, _ = c.Seek(search); found != nil {
		search = newIndexKey(ik.Owner(), found.Key(), ik.Generation())
		var v btrfs.Item
		if found, v = c.Seek(search); bytes.Compare(found, search) <= 0 {
			return found, v
		} else if end != nil && bytes.Compare(found, end) <= 0 {
			//fmt.Println("!", found.Owner(), found.Key(), found.Generation(), "-", search.Owner(), search.Key(), search.Generation())
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
	cursor   *bolt.Cursor
	key, end indexKey
	value    btrfs.Item
}

func (r *Range) Index() *Index { return r.ix }

func (r *Range) HasNext() bool {
	return r.key != nil && bytes.Compare(r.key, r.end) <= 0
}

func (r *Range) Next() []byte {
	if r.key, r.value = findNext(r.cursor, r.key, nil); r.key != nil {
		return btrfs.Item(r.value).Data()
	}
	return nil
}

func (r *Range) Owner() uint64      { return r.key.Owner() }
func (r *Range) Key() btrfs.Key     { return r.key.Key() }
func (r *Range) Generation() uint64 { return r.key.Generation() }
func (r *Range) Item() btrfs.Item   { return r.value }

// Range returns an index range [low, high) for the given keys.
func (ix *Index) Range(owner uint64, first, last btrfs.Key) (Range, []byte) {
	r := Range{
		ix:     ix,
		cursor: ix.bucket.Cursor(),
		end:    newIndexKey(owner, last, ix.Generation),
	}
	lowerFirst := lowerBound(r.cursor, owner, first, indexKeyOffset)
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
	if r.key, r.value = findNext(r.cursor, r.key, r.end); r.key != nil {
		return btrfs.Item(r.value).Data()
	}
	return nil
}

// FullRange returns an index range for all items in the index. This is
// mainly useful for debugging.
func (ix *Index) FullRange() (FullRange, []byte) {
	r := FullRange{Range{
		ix:     ix,
		cursor: ix.bucket.Cursor(),
		end:    newIndexKey(^uint64(0), KL(), ix.Generation),
	}}
	// No check, since open should fail if there's no data
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

func (ix *Index) FindExtentItem(owner uint64, id uint64) btrfs.FileExtentItem {
	if i := ix.FindItem(owner, KF(btrfs.ExtentDataKey, id)); i != nil {
		return i.Data()
	}
	return nil
}

//func (ix *Index) FindInodeRefItem(inode uint64) int {
//	if i, end := ix.Range(KF(btrfs.InodeRefKey, inode),
//		KL(btrfs.InodeRefKey, inode)); i < end {
//		return i
//	}
//	return -1
//}

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

func (ix *Index) Experimental() {
	var ik indexKey
	var di btrfs.DirItem
	var v []byte
	_ = ik
	_ = di

	for r, v := ix.Subvolumes(); r.HasNext(); v = r.Next() {
		fmt.Printf("ID %d gen %d top level %d path \n", r.Key().ObjectID, v.Generation(), v.Level())
	}

	fmt.Println("-----------------------")
	dirID := uint64(256)
	for r, v := ix.DirItems(btrfs.FSTreeObjectID, dirID); r.HasNext(); v = r.Next() {
		fmt.Printf("%d %s %d %s ", r.Owner(), r.Key(), r.Generation(), v.Name())
		if v.IsSubvolume() {
			fmt.Printf("subvol %s", v.Location())
		} else {
			fmt.Printf("%s", v.Location())
			if ii := ix.FindInodeItem(r.Owner(), v.Location().ObjectID); ii != nil {
				fmt.Printf("%d", ii.Size())
			}
		}
		fmt.Println()
	}
	fmt.Println("-----------------------")
	di = ix.FindDirItemForPath(btrfs.FSTreeObjectID, "z0-subvolume-03-aaaaaaaa/z0-subvolume-04-bbbbbbbb/demo/jpda/src.zip")
	if di != nil {
		fmt.Println(di.Name())
	}

	fmt.Println("-----------------------")

	return
	c := ix.bucket.Cursor()
	fmt.Println("-----------------------")
	ik, v = find(c, btrfs.RootTreeObjectID, KF(btrfs.RootItemKey, btrfs.FirstFreeObjectID), ix.Generation)
	if ik != nil {
		fmt.Println("  ", ik.Key(), ik.Generation(), btrfs.Item(v).Key(), btrfs.RootItem(v[btrfs.ItemLen:]).RootDirID())
	}
	fmt.Println("-----------------------")
	ik, v = find(c, btrfs.RootTreeObjectID, KF(btrfs.RootItemKey, btrfs.FirstFreeObjectID), ^uint64(0))
	if ik != nil {
		fmt.Println("  ", ik.Key(), ik.Generation(), btrfs.Item(v).Key(), btrfs.RootItem(v[btrfs.ItemLen:]).RootDirID())
	}
	fmt.Println("-----------------------")
	ik, v = find(c, btrfs.RootTreeObjectID, KF(btrfs.RootItemKey, btrfs.FirstFreeObjectID), 0)
	if ik != nil {
		fmt.Println("  ", ik.Key(), ik.Generation(), btrfs.Item(v).Key(), btrfs.RootItem(v[btrfs.ItemLen:]).RootDirID())
	}
	fmt.Println("-----------------------")
	ik, v = find(c, btrfs.RootTreeObjectID, KF(btrfs.RootItemKey, btrfs.FirstFreeObjectID+1), 0)
	if ik != nil {
		fmt.Println("  ", ik.Key(), ik.Generation(), btrfs.Item(v).Key(), btrfs.RootItem(v[btrfs.ItemLen:]).RootDirID())
	}
	fmt.Println("-----------------------")
	ik, v = find(c, btrfs.RootTreeObjectID, KF(btrfs.RootItemKey, 320), 5)
	if ik != nil {
		fmt.Println("  ", ik.Key(), ik.Generation(), btrfs.Item(v).Key(), btrfs.RootItem(v[btrfs.ItemLen:]).RootDirID())
	}
}

func (ix *Index) Physical(logical uint64) (devID uint64, offset uint64) {
	//if low, high := ix.Range(KF(btrfs.ChunkItemKey, btrfs.FirstChunkTreeObjectID),
	//	KL(btrfs.ChunkItemKey, btrfs.FirstChunkTreeObjectID)); low < high {
	//	if i := low + sort.Search(high-low, func(i int) bool {
	//		i += low
	//		return int64(logical-ix.Key(i).Offset) < int64(ix.Chunk(i).Length)
	//	}); i < high {
	//		stripe := ix.Chunk(i).Stripes[0]
	//		return stripe.DevID, stripe.Offset + logical - ix.Key(i).Offset
	//	}
	//}
	return
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
