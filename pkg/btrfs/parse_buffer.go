// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Read-only and restartable byte buffer

package btrfs

import (
	"encoding/binary"
	"time"
)

// ParseBuffer is similar to bytes.Buffer, except that it is read-only and can
// be restarted.
type ParseBuffer struct {
	binary.ByteOrder
	buf    []byte // Read from the bytes buf[offset : len(buf)]
	offset int    // At &buf[offset]
}

func NewParseBuffer(buf []byte) *ParseBuffer {
	return &ParseBuffer{binary.LittleEndian, buf, 0}
}

func (b *ParseBuffer) Rewind() {
	b.SetOffset(0)
}

func (b *ParseBuffer) Len() int {
	return len(b.buf)
}

func (b *ParseBuffer) Cap() int {
	return cap(b.buf)
}

func (b *ParseBuffer) Offset() int {
	return b.offset
}

func (b *ParseBuffer) SetOffset(offset int) {
	b.offset = offset
}

func (b *ParseBuffer) Unread() int {
	return len(b.buf) - b.offset
}

// Next returns a slice containing the next n bytes from the buffer and
// advancing it. Panics if there are fewer than n bytes in the buffer.
// The slice is only valid until the next call to one of the next methods.
func (b *ParseBuffer) Next(n int) []byte {
	data := b.buf[b.offset : b.offset+n]
	b.offset += n
	return data
}

func (b *ParseBuffer) NextUint8() uint8 {
	return b.Next(1)[0]
}

func (b *ParseBuffer) NextUint16() uint16 {
	return b.ByteOrder.Uint16(b.Next(2))
}

func (b *ParseBuffer) NextUint32() uint32 {
	return b.ByteOrder.Uint32(b.Next(4))
}

func (b *ParseBuffer) NextUint64() uint64 {
	return b.ByteOrder.Uint64(b.Next(8))
}

func (b *ParseBuffer) NextTime() time.Time {
	return time.Unix(int64(b.NextUint64()), int64(b.NextUint32()))
}
