/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Identify filesystems sub-command
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
 * ARE DISCLAIMED. IN NO EVENT SHALL CHRISTIAN BLICHMANN BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 * THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"blichmann.eu/code/btrfscue/btrfs"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

const (
	FromStart = iota
	FromCurrent
	FromEnd
)

func CheckedDeviceSize(rs io.ReadSeeker, blockSize uint64) (uint64, error) {
	size, err := rs.Seek(0, FromEnd)
	if err == nil && uint64(size) < blockSize {
		err = fmt.Errorf("device smaller than block size: %d < %d", size,
			blockSize)
	}
	return uint64(size), err
}

func ReadBlockAt(r io.ReaderAt, block []byte, offset, blockSize uint64) error {
	read, err := r.ReadAt(block, int64(offset))
	if uint64(read) != blockSize {
		err = fmt.Errorf("read %d bytes, expected: %d", read, blockSize)
	}
	return err
}

func ShannonEntropy(b []byte) float64 {
	hist := make(map[byte]uint)
	for _, v := range b {
		hist[v]++
	}
	p := make(map[byte]float64)
	l := float64(len(b))
	for v, c := range hist {
		p[v] = float64(c) / l
	}
	e := float64(0)
	for _, v := range p {
		e += v * math.Log2(v)
	}
	return -e
}

type Uint64Slice []uint64

func (a Uint64Slice) Len() int           { return len(a) }
func (a Uint64Slice) Less(i, j int) bool { return a[i] < a[j] }
func (a Uint64Slice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Uint64Slice) Sort()              { sort.Sort(a) }

type UuidOccurrence struct {
	Uuid    btrfs.Uuid
	Count   uint
	Entropy float64
}

type UuidOccurrenceList []UuidOccurrence

func (a UuidOccurrenceList) Len() int { return len(a) }
func (a UuidOccurrenceList) Less(i, j int) bool {
	ci, cj := a[i].Count, a[j].Count
	if ci != cj {
		return ci < cj
	}
	ei, ej := a[i].Entropy, a[j].Entropy
	if ei != ej {
		return ei < ej
	}
	return bytes.Compare(a[i].Uuid[:], a[j].Uuid[:]) < 0
}
func (a UuidOccurrenceList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a UuidOccurrenceList) Sort()         { sort.Sort(a) }

type identifyCommand struct {
	blockSize      *uint
	sampleFraction *float64
	minBlocks      *uint
	maxBlocks      *uint
	minOccurrence  *uint
}

func (ic *identifyCommand) DefineFlags(fs *flag.FlagSet) {
	ic.blockSize = fs.Uint("block-size", btrfs.DefaultBlockSize,
		"file system block size, usually the memory page size of the system "+
			"that created the it")
	ic.sampleFraction = fs.Float64("sample-fraction", 0.0001,
		"fraction of blocks to sample for filesystem ids")
	ic.minBlocks = fs.Uint("min-blocks", 1000, "minimum number of blocks to "+
		"scan")
	ic.maxBlocks = fs.Uint("max-blocks", 500000, "maximum number of blocks "+
		"to scan")
	ic.minOccurrence = fs.Uint("min-occurrence", 4, "minimum number of "+
		"occurrences of an id for a file system to be reported")
}

func (ic *identifyCommand) Run(args []string) {
	if len(args) == 0 {
		fatalf("missing device file\n")
	} else if len(args) > 1 {
		fatalf("extra operand '%s'\n", args[1])
	}

	filename := args[0]
	f, err := os.Open(filename)
	reportError(err)
	defer f.Close()

	blockSize := uint64(*ic.blockSize)

	// Get total file/device size
	fileSize, err := CheckedDeviceSize(f, blockSize)
	reportError(err)

	// Sanity check: BTRFS minimum filesystem size is 64MiB plus a few blocks
	if uint64(fileSize) < btrfs.SuperInfoOffset2+blockSize*100 {
		fatalf("'%s' too small, must be > 64MiB\n", filename)
	}

	// Parse sampleFraction * 100% of all blocks (minimum minBlocks, up to
	// maxBlocks) like this:
	// 1. Read 100 blocks in the vicinity of all superblock copies and collect
	//    FSIDs.
	// 2. Read the rest of the blocks distributed randomly and collect FSIDs
	// Return FSIDs that are most common.
	numBlocks := int64(fileSize / blockSize)
	numSamples := uint(*ic.sampleFraction * float64(numBlocks))
	if numSamples < *ic.minBlocks {
		numSamples = *ic.minBlocks
	} else if numSamples > *ic.maxBlocks {
		numSamples = *ic.maxBlocks
	}

	verbosef("sampling %d blocks...\n", numSamples)

	rand.Seed(time.Now().UnixNano())
	sampleSet := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		sampleSet[btrfs.SuperInfoOffset+uint64(i)*blockSize] = true
		sampleSet[btrfs.SuperInfoOffset2+uint64(i+100)*blockSize] = true
	}
	if fileSize >= btrfs.SuperInfoOffset3 {
		for i := 0; i < 100; i++ {
			sampleSet[btrfs.SuperInfoOffset3+uint64(i+200)*blockSize] = true
		}
		if fileSize >= btrfs.SuperInfoOffset4 {
			for i := 0; i < 100; i++ {
				sampleSet[btrfs.SuperInfoOffset3+uint64(i+300)*blockSize] = true
			}
		}
	}
	for uint(len(sampleSet)) < numSamples {
		sampleSet[uint64(rand.Int63n(numBlocks))*blockSize] = true
	}

	// Sort samples vector to access device in one direction only
	samples := make(Uint64Slice, 0, len(sampleSet))
	for offset, _ := range sampleSet {
		samples = append(samples, offset)
	}
	samples.Sort()

	buf := make([]byte, blockSize)
	h := btrfs.Header{}
	zeroFsId := btrfs.Uuid{}
	ffffFsId := btrfs.Uuid{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	hist := make(map[btrfs.Uuid]uint)
	for _, offset := range samples {
		reportError(ReadBlockAt(f, buf, offset, blockSize))
		h.Parse(buf)
		// Only gather blocks that look like leaves
		if !h.IsLeaf() {
			continue
		}
		// Skip blocks with zero FSID or with an FSID that consists only of
		// 0xFF bytes
		if h.FsId == zeroFsId || h.FsId == ffffFsId {
			continue
		}
		hist[h.FsId]++
	}

	// Sort FSIDs by number of times found, skip items that occur less than
	// four times.
	occ := make(UuidOccurrenceList, 0, len(hist))
	for uuid, count := range hist {
		if count > *ic.minOccurrence {
			occ = append(occ, UuidOccurrence{uuid, count,
				ShannonEntropy(uuid[:])})
		}
	}
	sort.Sort(sort.Reverse(occ))

	w := tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	fmt.Fprintln(w, "fsid\tcount\tentropy")
	for _, entry := range occ {
		fmt.Fprintf(w, "%s\t%d\t%.6f\n", entry.Uuid, entry.Count,
			entry.Entropy)
	}
	w.Flush()
}
