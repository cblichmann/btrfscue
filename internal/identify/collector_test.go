// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Collect filesystem IDs from block I/O

package identify

import (
	"blichmann.eu/code/btrfscue/pkg/btrfs"
	"blichmann.eu/code/btrfscue/pkg/uuid"
	"encoding/binary"
	"testing"
)

func makeHeader(block []byte, fsid uuid.UUID, nrItems uint32) {
	o := btrfs.CSumSize
	copy(block[o:], fsid[:])
	o += uuid.UUIDSize + 8 /* ByteNr */ + 8 /* Flags */ +
		uuid.UUIDSize /* ChunkTreeUUID */ + 8 /* Generation */ +
		8 /* Owner */
	binary.LittleEndian.PutUint32(block[o:], nrItems)
}

func makeFirstItem(block []byte, offset uint32) {
	o := btrfs.HeaderLen + btrfs.KeyLen /* Key */
	binary.LittleEndian.PutUint32(block[o:], offset)
}

func TestCollecter(t *testing.T) {
	const (
		shouldIdentifyId  = "a0dbfe80-3a38-11ea-b510-2ff108252d04"
		numFsIdFound      = 15
		fsidForBlockSize  = "d39dcd77-1133-4e69-b69e-197a9976f7f1"
		expectedEntries   = 4
		expectedBlockSize = 16384
	)
	var headers = []struct {
		Fsid            string
		Times           int
		NrItems         int
		FirstItemOffset int
	}{
		// NrItems and FirstItemOffset from a real FS, as collected by "recon"
		{"a2ecf93e-3a35-11ea-a363-4fdb514b33aa", 4, 1, 16191},
		{fsidForBlockSize, 1, 2, 16235},
		{fsidForBlockSize, 1, 3, 16235},
		{fsidForBlockSize, 1, 5, 16235},
		{fsidForBlockSize, 1, 6, 16185},
		{fsidForBlockSize, 1, 6, 16243},
		{fsidForBlockSize, 1, 8, 15844},
		{fsidForBlockSize, 1, 10, 15844},
		{shouldIdentifyId, numFsIdFound, 10, 16250},
		{fsidForBlockSize, 1, 10, 16259},
		{fsidForBlockSize, 1, 11, 16259},
		{fsidForBlockSize, 1, 12, 16230},
		// The next two should be ignored
		{"00000000-0000-0000-0000-000000000000", 100, 0, 0},
		{"ffffffff-ffff-ffff-ffff-ffffffffffff", 1, 0, 0},
		{fsidForBlockSize, 1, 12, 16259},
		{fsidForBlockSize, 1, 13, 16230},
		{"65cab3bc-3a39-11ea-80ab-cbca08b47b3b", 7, 28, 16123},
	}

	c := FSIDCollecter{}
	for _, header := range headers {
		fsid, _ := uuid.New(header.Fsid)
		for i := 0; i < header.Times; i++ {
			block := make(btrfs.Header, btrfs.X86RegularPageSize)
			makeHeader(block, fsid, uint32(header.NrItems))
			makeFirstItem(block, uint32(header.FirstItemOffset))
			c.CollectBlock(block)
		}
	}

	entries := c.Entries(4)
	if len(entries) != expectedEntries {
		t.Errorf("expected %d entries, actual %d", expectedEntries,
			len(entries))
	}
	e := entries[0]
	if e.FSID.String() != shouldIdentifyId {
		t.Errorf("expected FSID %s, actual %s", shouldIdentifyId, e.FSID)
	}
	if e.Count != numFsIdFound {
		t.Errorf("expected %d FSIDs found, actual %d", numFsIdFound, e.Count)
	}
	if e.BlockSize != expectedBlockSize {
		t.Errorf("expected block size %d, actual %d", expectedBlockSize,
			e.BlockSize)
	}
}
