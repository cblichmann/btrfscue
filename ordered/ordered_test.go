/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Ordered sets/multisets tests
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

package ordered // import "blichmann.eu/code/btrfscue/ordered"

import (
	"math/rand"
	"sort"
	"testing"
)

type pair struct {
	First  int
	Second int
}

func TestIntCompare(t *testing.T) {
	if IntCompare(10, 10) != 0 {
		t.FailNow()
	}
	if IntCompare(-10, -10) != 0 {
		t.FailNow()
	}
	if IntCompare(-10, 10) >= 0 {
		t.FailNow()
	}
	if IntCompare(10, -10) <= 0 {
		t.FailNow()
	}
}

func TestFloat64Compare(t *testing.T) {
	if Float64Compare(549755813888.0, 549755813888.0) != 0 {
		t.FailNow()
	}
	if Float64Compare(-549755813888.0, -549755813888.0) != 0 {
		t.FailNow()
	}
	if Float64Compare(-549755813888.0, 549755813888.0) >= 0 {
		t.FailNow()
	}
	if Float64Compare(549755813888.0, -549755813888.0) <= 0 {
		t.FailNow()
	}
}

func TestSetConstruction(t *testing.T) {
	s := NewSet(IntCompare, 3, 2, 1, 0)
	if s.Len() != 4 {
		t.FailNow()
	}
	if !sort.IsSorted(s) {
		t.FailNow()
	}
}

func TestInsertion(t *testing.T) {
	s := NewSet(IntCompare, 3, 2, 1)
	for i := 8; i > 3; i-- {
		s.Insert(i)
	}
	if !sort.IsSorted(s) {
		t.FailNow()
	}
	for i, _ := range s.Data() {
		if s.IntAt(i) != i+1 {
			t.Fatalf("%d vs %d", s.IntAt(i), i+1)
		}
	}
	if pos, didInsert := s.Insert(2); pos != 1 || didInsert {
		t.Fatalf("%d vs %d, %t", pos, 1, didInsert)
	}
}

func TestRandomInsertion(t *testing.T) {
	s := NewSet(IntCompare)
	const n = 1000
	for i := 0; i < n; i++ {
		s.Insert(rand.Intn(n * 10))
	}
	last := 0
	for i, _ := range s.Data() {
		cur := s.IntAt(i)
		if cur < last {
			t.Fatalf("%d vs %d", cur, last)
		}
		last = cur
	}
}

func TestInsertionWithPairs(t *testing.T) {
	s := NewSet(func(a, b interface{}) int {
		if r := a.(pair).First - b.(pair).First; r != 0 {
			return r
		}
		return a.(pair).Second - b.(pair).Second
	})
	s.Insert(pair{1, 80})
	s.Insert(pair{4, 10})
	s.Insert(pair{3, 20})
	s.Insert(pair{2, 30})
	s.Insert(pair{1, 20})
	s.Insert(pair{1, 40})
	for i, v := range s.Data() {
		p := v.(pair)
		t.Logf("%d: %d %d", i, p.First, p.Second)
	}
}

func TestFind(t *testing.T) {
	s := NewSet(IntCompare, 0, 1, 2, 3)
	for i, v := range s.Data() {
		if s.Find(v) != i {
			t.FailNow()
		}
	}
}

func TestErase(t *testing.T) {
	s := NewSet(IntCompare, 0, 1, 2, 3)
	s.EraseAt(2)
	if s.Len() != 3 {
		t.FailNow()
	}
}

func TestMultiSetConstruction(t *testing.T) {
	s := NewMultiSet(IntCompare, 3, 2, 2, 1, 0)
	if s.Len() != 5 {
		t.FailNow()
	}
	if !sort.IsSorted(s) {
		t.FailNow()
	}
}

func TestMultiInsertion(t *testing.T) {
	s := NewMultiSet(func(a, b interface{}) int {
		return a.(pair).First - b.(pair).First
	})
	s.Insert(pair{1, 80})
	s.Insert(pair{4, 10})
	s.Insert(pair{3, 20})
	s.Insert(pair{2, 30})
	s.Insert(pair{1, 20})
	s.Insert(pair{1, 40})
	if !sort.IsSorted(s) {
		t.FailNow()
	}
	// Check stable order of secondary value
	if s.At(0).(pair).Second != 80 ||
		s.At(1).(pair).Second != 40 ||
		s.At(2).(pair).Second != 20 {
		t.FailNow()
	}
}

func BenchmarkInsertion(b *testing.B) {
	s := NewSet(IntCompare).(*container)
	s.array = make([]interface{}, 0, b.N)
	for i := 0; i < b.N/2; i++ {
		s.Insert(i * 2)
	}
	for i := b.N / 2; i > 0; i-- {
		s.Insert(i * 4)
	}
}
