// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Shanon entropy tests

package coding

import (
	"math"
	"testing"
)

// round rounds a float64 to a number of digits after the comma.
func round(f float64, p int) float64 {
	scale := math.Pow(10, float64(p))
	return float64(int(f*scale+math.Copysign(0.5, f))) / scale
}

func TestShannonEntropy(t *testing.T) {
	values := []struct {
		b []byte
		e float64
		r int
	}{
		// Rounded sample values taken from: Kozlowski, L. Shannon entropy
		// calculator. www.shannonentropy.netmark.pl
		{[]byte("1100101"), 0.98523, 5},
		{[]byte("Lorem ipsum"), 3.27761, 5},
		// http://planetcalc.com/2476/
		{[]byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit, " +
			"sed do eiusmod tempor incididunt ut labore et dolore magna " +
			"aliqua."), 4.02, 2},
	}
	for _, v := range values {
		if a := round(ShannonEntropy(v.b), v.r); a != v.e {
			t.Fatalf("expected %f, got: %f", v.e, a)
		}
	}
}
