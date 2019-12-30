/*
 * btrfscue version 0.6
 * Copyright (c)2011-2020 Christian Blichmann
 *
 * Collect filesystem IDs from block I/O
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

package identify // import "blichmann.eu/code/btrfscue/identify"

import (
	"bytes"
	"sort"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/coding"
	"blichmann.eu/code/btrfscue/uuid"
)

type FSEntry struct {
	FSID      uuid.UUID
	Count     uint
	Entropy   float64
	BlockSize uint32
}

type fsEntries []FSEntry

func (a fsEntries) Len() int { return len(a) }
func (a fsEntries) Less(i, j int) bool {
	ci, cj := a[i].Count, a[j].Count
	if ci != cj {
		return ci < cj
	}
	ei, ej := a[i].Entropy, a[j].Entropy
	if ei != ej {
		return ei < ej
	}
	return bytes.Compare(a[i].FSID[:], a[j].FSID[:]) < 0
}
func (a fsEntries) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a fsEntries) Sort()         { sort.Sort(a) }

type histEntry struct {
	Count        uint
	BlockSizeSum uint64
}

type FSIDCollecter struct {
	hist map[uuid.UUID]*histEntry
}

func (c *FSIDCollecter) CollectBlock(block []byte) {
	h := btrfs.Header(block)
	// Only gather blocks that look like leaves
	if !h.IsLeaf() {
		return
	}
	fsid := h.FSID()
	// Skip blocks with zero FSID or with an FSID that consists only of 0xFF
	// bytes
	if fsid.IsZero() || fsid.IsAllFs() {
		return
	}
	if c.hist == nil {
		c.hist = make(map[uuid.UUID]*histEntry)
	}
	entry, ok := c.hist[fsid]
	if !ok {
		entry = &histEntry{}
		c.hist[fsid] = entry
	} else {
		if h.NrItems() > 0 {
			item := btrfs.Item(block[btrfs.HeaderLen:])
			// Since item headers and their data grow towards each other, the
			// first item's offset will be the largest. In order to guess the
			// actual block size, sum offsets for all entries belonging to an
			// FSID to compute the average later.
			entry.BlockSizeSum += uint64(item.Offset())
		}
	}
	entry.Count++
}

func (c *FSIDCollecter) Entries(minOccurrence uint) []FSEntry {
	// Sort FSIDs by number of times found, skip items that occur less than
	// minOccurrence times.
	occ := make(fsEntries, 0, len(c.hist))
	for uuid, entry := range c.hist {
		if entry.Count > minOccurrence {
			// Compute average and round to nearest 4KiB.
			guess := uint32(float64(entry.BlockSizeSum)/float64(entry.Count)+
				btrfs.X86RegularPageSize) / btrfs.X86RegularPageSize *
				btrfs.X86RegularPageSize
			occ = append(occ, FSEntry{uuid, entry.Count,
				coding.ShannonEntropy(uuid[:]), guess})
		}
	}
	sort.Sort(sort.Reverse(occ))
	return occ
}
