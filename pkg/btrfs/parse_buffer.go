/*
 * btrfscue version 0.6
 * Copyright (c)2011-2022 Christian Blichmann
 *
 * Read-only and restartable byte buffer
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
