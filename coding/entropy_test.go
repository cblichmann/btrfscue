/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Shanon entropy tests
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

package coding // import "blichmann.eu/code/btrfscue/coding"

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
			t.Fatalf("%f vs. %f", a, v.e)
		}
	}
}
