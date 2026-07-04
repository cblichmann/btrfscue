// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Collect filesystem IDs from block I/O

package identify

import (
	"bytes"
	"sort"

	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/coding"
	"blichmann.eu/code/btrfscue/pkg/uuid"
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
	} else if h.NrItems() > 0 {
		item := btrfs.Item(block[btrfs.HeaderLen:])
		// Since item headers and their data grow towards each other, the
		// first item's offset will be the largest. In order to guess the
		// actual block size, sum offsets for all entries belonging to an
		// FSID to compute the average later.
		entry.BlockSizeSum += uint64(item.Offset())
	}
	entry.Count++
}

func (c *FSIDCollecter) Entries(minOccurrence uint) []FSEntry {
	// Sort FSIDs by number of times found, skip items that occur less than
	// minOccurrence times.
	occ := make(fsEntries, 0, len(c.hist))
	for uuid, entry := range c.hist {
		if entry.Count >= minOccurrence {
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
