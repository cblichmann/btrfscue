// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Compute Shanon entropy of byte slices

package coding

import (
	"math"
)

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
