/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Utility functions for dealing with byte slices
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
	"encoding/binary"
	"time"

	"blichmann.eu/code/btrfscue/uuid"
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
