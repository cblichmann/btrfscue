// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Utility functions for dealing with byte slices

package btrfs

import (
	"encoding/binary"
	"time"

	"blichmann.eu/code/btrfscue/pkg/uuid"
)

func SliceUint16LE(b []byte) uint16 { return binary.LittleEndian.Uint16(b[:2]) }
func SliceUint32LE(b []byte) uint32 { return binary.LittleEndian.Uint32(b[:4]) }
func SliceUint64LE(b []byte) uint64 { return binary.LittleEndian.Uint64(b[:8]) }

func SliceUUID(b []byte) uuid.UUID {
	u := uuid.UUID{}
	copy(u[:], b[:uuid.UUIDSize])
	return u
}

func SliceKey(b []byte) Key {
	return Key{SliceUint64LE(b), uint8(b[8]), SliceUint64LE(b[9:])}
}

func SliceTimeLE(b []byte) time.Time {
	return time.Unix(int64(SliceUint64LE(b)), int64(SliceUint32LE(b[8:])))
}
