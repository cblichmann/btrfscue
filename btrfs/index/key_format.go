package index // import "blichmann.eu/code/btrfscue/btrfs/index"

import (
	"encoding/binary"

	"blichmann.eu/code/btrfscue/btrfs"
)

// keyV1 holds the BTRFS key of an owned FS object as well as its
// generation. The tuple (owner, Key, generation) is encoded in big endian for
// lexicographical comparison.
// This is the key format used up until metadata version 20161109.
type keyV1 []byte

// Offsets for parsing from byte slice
const (
	keyV1Owner      = 0
	keyV1Type       = keyV1Owner + 8
	keyV1ObjectID   = keyV1Type + 1
	keyV1Offset     = keyV1ObjectID + 8
	keyV1Generation = keyV1Offset + 8
	keyV1End        = keyV1Generation + 8
)

func (ik keyV1) Owner() uint64      { return binary.BigEndian.Uint64(ik[keyV1Owner:]) }
func (ik keyV1) Type() uint8        { return uint8(ik[keyV1Type]) }
func (ik keyV1) ObjectID() uint64   { return binary.BigEndian.Uint64(ik[keyV1ObjectID:]) }
func (ik keyV1) Offset() uint64     { return binary.BigEndian.Uint64(ik[keyV1Offset:]) }
func (ik keyV1) Generation() uint64 { return binary.BigEndian.Uint64(ik[keyV1Generation:]) }
func (ik keyV1) Key() btrfs.Key {
	return btrfs.Key{ObjectID: ik.ObjectID(), Type: ik.Type(), Offset: ik.Offset()}
}

// keyV2 holds the BTRFS key of an owned FS object as well as its
// generation. The tuple (generation, owner, Key) is encoded in big endian for
// lexicographical comparison.
// This is the key format in use since metadata version 20190809.
type keyV2 []byte

// Offsets for parsing from byte slice
const (
	keyV2Owner      = 0
	keyV2Type       = keyV2Owner + 8
	keyV2ObjectID   = keyV2Type + 1
	keyV2Offset     = keyV2ObjectID + 8
	keyV2Generation = keyV2Offset + 8
	keyV2End        = keyV2Generation + 8
)

func newIndexKey(owner uint64, k btrfs.Key, generation uint64) keyV2 {
	ik := [keyV2End]byte{}
	binary.BigEndian.PutUint64(ik[keyV2Owner:], owner)
	ik[keyV2Type] = k.Type
	binary.BigEndian.PutUint64(ik[keyV2ObjectID:], k.ObjectID)
	binary.BigEndian.PutUint64(ik[keyV2Offset:], k.Offset)
	binary.BigEndian.PutUint64(ik[keyV2Generation:], generation)
	return ik[:]
}

func NewIndexKeyFromV1(ik keyV1) keyV2 {
	return newIndexKey(ik.Owner(), ik.Key(), ik.Generation())
}

func (ik keyV2) Owner() uint64      { return binary.BigEndian.Uint64(ik[keyV2Owner:]) }
func (ik keyV2) Type() uint8        { return uint8(ik[keyV2Type]) }
func (ik keyV2) ObjectID() uint64   { return binary.BigEndian.Uint64(ik[keyV2ObjectID:]) }
func (ik keyV2) Offset() uint64     { return binary.BigEndian.Uint64(ik[keyV2Offset:]) }
func (ik keyV2) Generation() uint64 { return binary.BigEndian.Uint64(ik[keyV2Generation:]) }
func (ik keyV2) Key() btrfs.Key {
	return btrfs.Key{ObjectID: ik.ObjectID(), Type: ik.Type(), Offset: ik.Offset()}
}

type KeyV1 keyV1
type KeyV2 keyV2

var MetadataKey = metadataKey
