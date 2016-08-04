/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Ordered sets/multisets backed by arrays
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
	"math"
	"sort"
	"strings"
)

type Set interface {
	// Inserts a new element into the set. Returns the inserted element and a
	// bool that indicates whether the element was inserted.
	// Complexity: O(log(n)) for searching plus a linear cost for relocating
	// elements if insertion was not at the end of the backing array.
	Insert(value interface{}) (maybeInserted interface{}, didInsert bool)

	Find(value interface{}) int

	At(int) interface{}

	// Removes the element at position i. Erasing elements in positions other
	// than Len()-1 causes the underlying container to relocate all elements in
	// positions > i.
	// Complexity: Linear on the number of elements erased.
	EraseAt(i int)

	// Convenience accessors
	IntAt(int) int
	Float64At(i int) float64
	StringAt(i int) string

	// Direct access to the backing array.
	Data() []interface{}
	Cap() int

	sort.Interface
}

type MultiSet interface {
	Set
}

func IntCompare(a, b interface{}) int { return a.(int) - b.(int) }
func Float64Compare(a, b interface{}) int {
	r := math.Float64bits(a.(float64) - b.(float64))
	return int(int32(r | r>>32 /* Always keep sign bit */))
}
func StringCompare(a, b interface{}) int {
	return strings.Compare(a.(string), b.(string))
}

func NewSet(compare func(a, b interface{}) int,
	values ...interface{}) Set {
	s := &container{noDuplicates, compare, values}
	sort.Sort(s)
	return s
}

func NewMultiSet(compare func(a, b interface{}) int,
	values ...interface{}) MultiSet {
	s := &container{stableDuplicates, compare, values}
	sort.Stable(s)
	return s
}

const (
	noDuplicates = iota
	stableDuplicates
)

type container struct {
	insertionPolicy uint8
	compare         func(a, b interface{}) int
	array           []interface{}
}

func (s container) At(i int) interface{} {
	return s.array[i]
}

func (s *container) EraseAt(i int) {
	copy(s.array[i:], s.array[i+1:])
	s.array[s.Len()-1] = nil
	s.array = s.array[:s.Len()-1]
}

func (s *container) IntAt(i int) int { return s.array[i].(int) }
func (s *container) Float64At(i int) float64 {
	return s.array[i].(float64)
}
func (s *container) StringAt(i int) string { return s.array[i].(string) }

func (s container) Data() []interface{} { return s.array }

func (s container) Cap() int {
	return cap(s.array)
}

func (s container) Len() int {
	return len(s.array)
}

func (s container) Less(i, j int) bool {
	return s.compare(s.array[i], s.array[j]) < 0
}

func (s container) Swap(i, j int) {
	s.array[i], s.array[j] = s.array[j], s.array[i]
}

func (s *container) findInsertionPos(v interface{}) int {
	return sort.Search(s.Len(), func(i int) bool {
		return s.compare(s.array[i], v) >= 0
	})
}

func (s *container) Insert(v interface{}) (interface{}, bool) {
	i := s.findInsertionPos(v)
	swap := false
	if i < s.Len() && s.compare(s.array[i], v) == 0 {
		if s.insertionPolicy == noDuplicates {
			return s.array[i], false
		}
		swap = true
	}
	// Grow the array in amortized constant time. This relies on the builtin
	// append() function to reallocate efficiently.
	s.array = append(s.array, nil)
	copy(s.array[i+1:], s.array[i:])
	if !swap {
		s.array[i] = v
	} else {
		s.array[i], s.array[i+1] = s.array[i+1], v
	}
	return s.array[i], true
}

func (s *container) Find(v interface{}) int {
	p := s.findInsertionPos(v)
	if p < s.Len() && s.compare(s.array[p], v) == 0 {
		return p
	}
	return -1
}
