/*
 * btrfscue version 0.5
 * Copyright (c)2011-2019 Christian Blichmann
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
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"gopkg.in/cheggaaa/pb.v1"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/btrfscue"
	"blichmann.eu/code/btrfscue/cliutil"
	"blichmann.eu/code/btrfscue/ioutil"
	"blichmann.eu/code/btrfscue/subcommand"
)

type Uint64Array []uint64

func (a Uint64Array) Len() int           { return len(a) }
func (a Uint64Array) Less(i, j int) bool { return a[i] < a[j] }
func (a Uint64Array) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Uint64Array) Sort()              { sort.Sort(a) }

func MakeSampleOffsets(devSize, blockSize, numSamples uint64) []uint64 {
	sampleSet := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		sampleSet[btrfs.SuperInfoOffset+uint64(i)*blockSize] = true
		sampleSet[btrfs.SuperInfoOffset2+uint64(i+100)*blockSize] = true
	}
	if devSize >= btrfs.SuperInfoOffset3 {
		for i := 0; i < 100; i++ {
			sampleSet[btrfs.SuperInfoOffset3+uint64(i+200)*blockSize] = true
		}
		if devSize >= btrfs.SuperInfoOffset4 {
			// For completeness, handle huge devices
			for i := 0; i < 100; i++ {
				sampleSet[btrfs.SuperInfoOffset3+uint64(i+300)*blockSize] = true
			}
		}
	}
	numBlocks := int64(devSize / blockSize)
	for uint64(len(sampleSet)) < numSamples {
		sampleSet[uint64(rand.Int63n(numBlocks))*blockSize] = true
	}
	// Sort samples vector to access device in one direction only
	samples := make(Uint64Array, numSamples)
	i := 0
	for o, _ := range sampleSet {
		samples[i] = o
		i++
	}
	samples.Sort()
	return samples
}

type identifyCommand struct {
	sampleFraction *float64
	minBlocks      *uint
	maxBlocks      *uint
	minOccurrence  *uint
}

func (ic *identifyCommand) DefineFlags(fs *flag.FlagSet) {
	ic.sampleFraction = fs.Float64("sample-fraction", 0.0001,
		"fraction of blocks to sample for filesystem ids")
	ic.minBlocks = fs.Uint("min-blocks", 1000, "minimum number of blocks to "+
		"scan")
	ic.maxBlocks = fs.Uint("max-blocks", 1000000, "maximum number of blocks "+
		"to scan")
	ic.minOccurrence = fs.Uint("min-occurrence", 4, "minimum number of "+
		"occurrences of an id for a file system to be reported")
}

func (ic *identifyCommand) Run(args []string) {
	if len(args) == 0 {
		cliutil.Fatalf("missing device file\n")
	} else if len(args) > 1 {
		cliutil.Fatalf("extra operand '%s'\n", args[1])
	}

	filename := args[0]
	f, err := os.Open(filename)
	cliutil.ReportError(err)
	defer f.Close()

	bs := uint64(*btrfscue.BlockSize)

	// Get total file/device size
	devSize, err := btrfs.CheckDeviceSize(f, bs)
	cliutil.ReportError(err)

	// Parse sampleFraction * 100% of all blocks (minimum minBlocks, up to
	// maxBlocks) like this:
	// 1. Read 100 blocks in the vicinity of all superblock copies and collect
	//    FSIDs.
	// 2. Read the rest of the blocks distributed randomly and collect FSIDs
	// Return FSIDs that are most common.
	numSamples := uint(*ic.sampleFraction * float64(devSize/bs))
	if numSamples < *ic.minBlocks {
		numSamples = *ic.minBlocks
	} else if numSamples > *ic.maxBlocks {
		numSamples = *ic.maxBlocks
	}

	cliutil.Verbosef("sampling %d blocks...\n", numSamples)

	rand.Seed(time.Now().UnixNano())
	samples := MakeSampleOffsets(devSize, bs, uint64(numSamples))

	bar := pb.New(len(samples))
	bar.Start()

	buf := make([]byte, bs)
	coll := FSIDCollecter{}
	for i, offset := range samples {
		bar.Set(i + 1)
		cliutil.ReportError(ioutil.ReadBlockAt(f, buf, offset, bs))
		coll.CollectBlock(buf)
	}
	bar.Finish()

	occ := coll.Entries(*ic.minOccurrence)
	if len(occ) == 0 {
		cliutil.Warnf("no filesystem id occured more than %d times, check "+
			"--min-occurrence\n", *ic.minOccurrence)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 4, 1, ' ', 0)
	fmt.Fprintln(w, "fsid\tcount\tentropy\tblock size")
	for _, entry := range occ {
		fmt.Fprintf(w, "%s\t%d\t%.6f\t%d\n", entry.FSID, entry.Count,
			entry.Entropy, entry.BlockSize)
	}
	w.Flush()
}

func init() {
	subcommand.Register("identify",
		"identify BTRFS filesystems on a device", &identifyCommand{})
}
