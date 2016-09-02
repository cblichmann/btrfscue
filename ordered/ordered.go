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

	"fmt" //DBG!!!
)

// Set backed by a sorted array. Except where noted, all operations preserve
// the invariant that the underlying array is sorted.
type Set interface {
	// Insert inserts a new element into the set. Returns the position of the
	// element and a bool that indicates whether the element was inserted.
	// Complexity: O(log(n)) for searching plus a linear cost for relocating
	// elements if insertion was not at the end of the backing array.
	Insert(value interface{}) (maybeInsertedAt int, didInsert bool)

	// BatchInsert inserts new elements into the set. After calling
	// BatchInsert, the backing array may no longer be sorted. Call Fix to
	// restore the invariant. Note that the default implementations keep the
	// set sorted.
	// Returns whether the all of the elements in values have been inserted.
	BatchInsert(values ...interface{}) bool

	// Fix restores the sort order in the underlying array.
	Fix()

	// Find returns the index of the first item that is equal to value. If
	// there is no such item, Find returns Len.
	Find(value interface{}) int

	At(int) interface{}

	// EraseAt removes the element at position i. Erasing elements in
	// positions other than Len()-1 causes the set to relocate all elements
	// in positions > i.
	// Complexity: Linear on the number of elements erased.
	EraseAt(i int)

	// Convenience accessors. These may panic if the item at the specified
	// index is not convertible into the requested type.
	IntAt(int) int
	Float64At(int) float64
	StringAt(int) string

	// LowerBound finds the index in Data() where v would be inserted,
	// starting at index from.
	LowerBound(from, to int, v interface{}) int

	// Data provides direct access into the underlying array. This can be used
	// to conveniently range over the data in place of using At and Len.
	Data() []interface{}

	// Cap returns the capacity of the backing array.
	Cap() int

	sort.Interface
}

// MultiSet is a Set that allows duplicates.
type MultiSet interface {
	Set

	// TODO(cblichmann): These two are basically untested.
	UpperBound(from, to int, v interface{}) int
	EqualRange(from, to int, v interface{}) (low, high int)
}

type Compare func(a, b interface{}) int

// IntCompare is a convenience function that compares two integers
// lexicographically. It typecasts its arguments to int.
func IntCompare(a, b interface{}) int { return a.(int) - b.(int) }

// Float64Compare compares two float64 values by typecasting its arguments.
// This function is provided for convenience. See IntCompare().
func Float64Compare(a, b interface{}) int {
	r := math.Float64bits(a.(float64) - b.(float64))
	return int(int32(r | r>>32 /* Always keep sign bit */))
}

// StringCompare compares two strings by typecasting its arguments. Provided
// for convenience, see IntCompare().
func StringCompare(a, b interface{}) int {
	return strings.Compare(a.(string), b.(string))
}

type DefaultSet struct {
	Compare func(a, b interface{}) int
	Array   []interface{}
}

func NewDefaultSet(compare Compare) DefaultSet {
	return DefaultSet{Compare: compare,
		Array: make([]interface{}, 0)}
}

func NewSet(compare Compare, values ...interface{}) Set {
	s := NewDefaultSet(compare)
	s.BatchInsert(values...)
	s.Fix()
	return &s
}

func (s DefaultSet) At(i int) interface{} { return s.Array[i] }

func (s *DefaultSet) EraseAt(i int) {
	copy(s.Array[i:], s.Array[i+1:])
	s.Array[s.Len()-1] = nil
	s.Array = s.Array[:s.Len()-1]
}

func (s *DefaultSet) IntAt(i int) int { return s.Array[i].(int) }
func (s *DefaultSet) Float64At(i int) float64 {
	return s.Array[i].(float64)
}
func (s *DefaultSet) StringAt(i int) string { return s.Array[i].(string) }

func (s DefaultSet) Data() []interface{} { return s.Array }

func (s DefaultSet) Cap() int { return cap(s.Array) }

func (s DefaultSet) Len() int { return len(s.Array) }

func (s DefaultSet) Less(i, j int) bool {
	return s.Compare(s.Array[i], s.Array[j]) < 0
}

func (s DefaultSet) Swap(i, j int) {
	s.Array[i], s.Array[j] = s.Array[j], s.Array[i]
}

func (s *DefaultSet) LowerBound(from, to int, v interface{}) int {
	return sort.Search(to-from, func(i int) bool {
		return s.Compare(s.Array[from+i], v) >= 0
	})
}

func (s *DefaultSet) Find(v interface{}) int {
	if p := s.LowerBound(0, s.Len(), v); p < s.Len() &&
		s.Compare(s.Array[p], v) == 0 {
		return p
	}
	return s.Len()
}

func (s *DefaultSet) Insert(v interface{}) (int, bool) {
	i := s.LowerBound(0, s.Len(), v)
	if i < s.Len() && s.Compare(s.Array[i], v) == 0 {
		return i, false
	}
	// Grow the array in amortized constant time. This relies on the builtin
	// append() function to reallocate efficiently.
	s.Array = append(s.Array, nil)
	copy(s.Array[i+1:], s.Array[i:])
	s.Array[i] = v
	return i, true
}

func (s *DefaultSet) BatchInsert(values ...interface{}) bool {
	for _, v := range values {
		if _, ok := s.Insert(v); !ok {
			return false
		}
	}
	//for _, v := range values {
	//	if s.hash != nil {
	//		h := s.hash(v)
	//		if _, ok := s.batch[h]; ok {
	//			return false
	//		}
	//		s.batch[h] = true
	//	}
	//	s.Array = append(s.Array, v)
	//}
	return true
}

func (s *DefaultSet) Fix() {
	sort.Sort(s)
}

type KeyExtract func(v interface{}) string

type HashSet struct {
	DefaultSet
	keyExtract KeyExtract
	batch      map[string]bool
}

func NewHashSet(compare Compare, keyExtract KeyExtract) Set {
	return &HashSet{
		DefaultSet: NewDefaultSet(compare),
		keyExtract: keyExtract,
		batch:      make(map[string]bool),
	}
}

func (s *HashSet) BatchInsert(values ...interface{}) bool {
	for _, v := range values {
		k := s.keyExtract(v)
		if _, ok := s.batch[k]; ok {
			return false
		}
		s.batch[k] = true
		s.Array = append(s.Array, v)
	}
	return true
}

func (s *HashSet) Fix() {
	sort.Sort(s)
	s.batch = make(map[string]bool)
}

type DefaultMultiSet struct {
	DefaultSet
}

func NewMultiSet(compare Compare, values ...interface{}) MultiSet {
	s := &DefaultMultiSet{}
	s.DefaultSet = NewDefaultSet(compare)
	s.BatchInsert(values...)
	s.Fix()
	return s
}

func (s *DefaultMultiSet) BatchInsert(values ...interface{}) bool {
	s.Array = append(s.Array, values...)
	return true
}

func (s *DefaultMultiSet) Fix() {
	sort.Stable(s)
}

func (s *DefaultMultiSet) Insert(v interface{}) (int, bool) {
	i := s.UpperBound(0, s.Len(), v)
	// Grow the array in amortized constant time. This relies on the builtin
	// append() function to reallocate efficiently.
	s.Array = append(s.Array, nil)
	copy(s.Array[i+1:], s.Array[i:])
	s.Array[i] = v
	return i, true
}

func (s *DefaultMultiSet) UpperBound(from, to int, v interface{}) int {
	return from + sort.Search(to-from, func(i int) bool {
		return s.Compare(s.Array[from+i], v) > 0
	})
}

func (s *DefaultMultiSet) EqualRange(from, to int, v interface{}) (
	low, high int) {
	low = s.LowerBound(from, to, v)
	high = s.UpperBound(low, to, v)
	return
}

func init() {
	fmt.Printf("") //DBG!!!
}
