// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package histogram // import "go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"

import (
	"fmt"
	"math/bits"
)

type (
	// buckets stores counts for measurement values in the range
	// (0, +Inf).
	buckets struct {
		// backing is a slice of nil, []uint8, []uint16, []uint32, or []uint64
		backing bucketsBacking

		// indexBase is index of the 0th position in the
		// backing array, i.e., backing[0] is the count associated with
		// indexBase which is in [indexStart, indexEnd]
		indexBase int32

		// indexStart is the smallest index value represented
		// in the backing array.
		indexStart int32

		// indexEnd is the largest index value represented in
		// the backing array.
		indexEnd int32
	}

	// bucketsCount are the possible backing array widths.
	bucketsCount interface {
		uint8 | uint16 | uint32 | uint64
	}

	// bucketsVarwidth is a variable-width slice of unsigned int counters.
	bucketsVarwidth[N bucketsCount] struct {
		counts []N
	}

	// bucketsBacking is implemented by bucektsVarwidth[N].
	bucketsBacking interface {
		// size returns the physical size of the backing
		// array, which is >= buckets.Len() the number allocated.
		size() int32
		// growTo grows a backing array and copies old enties
		// into their correct new positions.
		growTo(newSize, oldPositiveLimit, newPositiveLimit int32)
		// reverse reverse the items in a backing array in the
		// range [from, limit).
		reverse(from, limit int32)
		// moveCount empies the count from a bucket, for
		// moving into another.
		moveCount(src int32) uint64
		// tryIncrement increments a bucket by `incr`, returns
		// false if the result would overflow the current
		// backing width.
		tryIncrement(bucketIndex int32, incr uint64) bool
		// countAt returns the count in a specific bucket.
		countAt(pos uint32) uint64
		// reset resets all buckets to zero count.
		reset()
	}

	// highLow is used to establish the maximum range of bucket
	// indices needed, in order to establish the best value of the
	// scale parameter.
	highLow struct {
		low  int32
		high int32
	}
)

// Move atomically copies and resets `s`.  The modification to `dest`
// is not synchronized.
func (s *State[N, Traits]) Move(dest *State[N, Traits]) {
	if dest != nil {
		// Swap case: This is the ordinary case for a
		// synchronous instrument, where the SDK allocates two
		// Aggregators and lock contention is anticipated.
		// Reset the target state before swapping it under the
		// lock below.
		dest.clearState()
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if dest != nil {
		dest.data, s.data = s.data, dest.data
	} else {
		// No swap case: This is the ordinary case for an
		// asynchronous instrument, where the SDK allocates a
		// single Aggregator and there is no anticipated lock
		// contention.
		s.clearState()
	}
}

// Update adds the recorded measurement to the current data set.
func (s *State[N, Traits]) Update(number N) {
	s.UpdateByIncr(number, 1)
}

// UpdateByIncr supports updating a histogram with a non-negative
// increment.
func (s *State[N, Traits]) UpdateByIncr(number N, incr uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	value := float64(number)

	// Note: Not checking for overflow here. TODO.
	s.count += incr

	if value == 0 {
		s.zeroCount++
		return
	}

	// Sum maintains the original type, otherwise we use the floating point value.
	s.sum += number * N(incr)

	var b *buckets
	if value > 0 {
		b = &s.positive
	} else {
		value = -value
		b = &s.negative
	}

	s.update(b, value, incr)
}

// downscale subtracts `change` from the current mapping scale.
func (s *State[N, Traits]) downscale(change int32) {
	if change < 0 {
		panic(fmt.Sprint("impossible change of scale", change))
	}
	newScale := s.mapping.Scale() - change

	s.positive.downscale(change)
	s.negative.downscale(change)
	var err error
	s.mapping, err = newMapping(newScale)
	if err != nil {
		panic(fmt.Sprint("impossible scale", newScale))
	}
}

// changeScale computes how much downscaling is needed by shifting the
// high and low values until they are separated by no more than size.
func changeScale(hl highLow, size int32) int32 {
	var change int32
	for hl.high-hl.low >= size {
		hl.high >>= 1
		hl.low >>= 1
		change++
	}
	return change
}

// update increments the appropriate buckets for a given absolute
// value by the provided increment.
func (s *State[N, Traits]) update(b *buckets, value float64, incr uint64) {
	index := s.mapping.MapToIndex(value)

	hl, success := s.incrementIndexBy(b, index, incr)
	if success {
		return
	}

	s.downscale(changeScale(hl, s.maxSize))

	index = s.mapping.MapToIndex(value)
	if _, success := s.incrementIndexBy(b, index, incr); !success {
		panic("downscale logic error")
	}
}

// increment determines if the index lies inside the current range
// [indexStart, indexEnd] and, if not, returns the minimum size (up to
// maxSize) will satisfy the new value.
func (s *State[N, Traits]) incrementIndexBy(b *buckets, index int32, incr uint64) (highLow, bool) {
	if incr == 0 {
		// Skipping a bunch of work for 0 increment.  This
		// happens when merging sparse data, for example.
		// This also happens UpdateByIncr is used with a 0
		// increment, means it can be safely skipped.
		return highLow{}, true
	}
	if b.Len() == 0 {
		if b.backing == nil {
			b.backing = &bucketsVarwidth[uint8]{
				counts: []uint8{0},
			}
		}
		b.indexStart = index
		b.indexEnd = b.indexStart
		b.indexBase = b.indexStart
	} else if index < b.indexStart {
		if span := b.indexEnd - index; span >= s.maxSize {
			// rescale needed: mapped value to the right
			return highLow{
				low:  index,
				high: b.indexEnd,
			}, false
		} else if span >= b.backing.size() {
			s.grow(b, span+1)
		}
		b.indexStart = index
	} else if index > b.indexEnd {
		if span := index - b.indexStart; span >= s.maxSize {
			// rescale needed: mapped value to the left
			return highLow{
				low:  b.indexStart,
				high: index,
			}, false
		} else if span >= b.backing.size() {
			s.grow(b, span+1)
		}
		b.indexEnd = index
	}

	bucketIndex := index - b.indexBase
	if bucketIndex < 0 {
		bucketIndex += b.backing.size()
	}
	b.incrementBucket(bucketIndex, incr)
	return highLow{}, true
}

// grow resizes the backing array by doubling in size up to maxSize.
// this extends the array with a bunch of zeros and copies the
// existing counts to the same position.
func (s *State[N, Traits]) grow(b *buckets, needed int32) {
	size := b.backing.size()
	bias := b.indexBase - b.indexStart
	oldPositiveLimit := size - bias
	newSize := int32(1) << (32 - bits.LeadingZeros32(uint32(needed)))
	if newSize > s.maxSize {
		newSize = s.maxSize
	}
	newPositiveLimit := newSize - bias
	b.backing.growTo(newSize, oldPositiveLimit, newPositiveLimit)
}

// downscale first rotates, then collapses 2**`by`-to-1 buckets
func (b *buckets) downscale(by int32) {
	b.rotate()

	size := 1 + b.indexEnd - b.indexStart
	each := int64(1) << by
	inpos := int32(0)
	outpos := int32(0)

	for pos := b.indexStart; pos <= b.indexEnd; {
		mod := int64(pos) % each
		if mod < 0 {
			mod += each
		}
		for i := mod; i < each && inpos < size; i++ {
			b.relocateBucket(outpos, inpos)
			inpos++
			pos++
		}
		outpos++
	}

	b.indexStart >>= by
	b.indexEnd >>= by
	b.indexBase = b.indexStart
}

// rotate shifts the backing array contents so that indexStart ==
// indexBase to simplify the downscale logic.
func (b *buckets) rotate() {
	bias := b.indexBase - b.indexStart

	if bias == 0 {
		return
	}

	// Rotate the array so that indexBase == indexStart
	b.indexBase = b.indexStart

	b.backing.reverse(0, b.backing.size())
	b.backing.reverse(0, bias)
	b.backing.reverse(bias, b.backing.size())
}

// relocateBucket adds the count in counts[src] to counts[dest] and
// resets count[src] to zero.
func (b *buckets) relocateBucket(dest, src int32) {
	if dest == src {
		return
	}

	b.incrementBucket(dest, b.backing.moveCount(src))
}

// incrementBucket increments the backing array index by `incr`.
func (b *buckets) incrementBucket(bucketIndex int32, incr uint64) {
	for {
		if b.backing.tryIncrement(bucketIndex, incr) {
			return
		}

		switch bt := b.backing.(type) {
		case *bucketsVarwidth[uint8]:
			b.backing = widenBuckets[uint8, uint16](bt)
		case *bucketsVarwidth[uint16]:
			b.backing = widenBuckets[uint16, uint32](bt)
		case *bucketsVarwidth[uint32]:
			b.backing = widenBuckets[uint32, uint64](bt)
		case *bucketsVarwidth[uint64]:
			// Problem. The exponential histogram has overflowed a uint64.
			// However, this shouldn't happen because the total count would
			// overflow first.
			panic("bucket overflow must be avoided")
		}
	}
}

// Merge combines data from `o` into `s`.  The modification to `s` is synchronized.
func (s *State[N, Traits]) Merge(o *State[N, Traits]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Note: Not checking for overflow here. TODO.
	s.sum += o.sum
	s.count += o.count
	s.zeroCount += o.zeroCount

	minScale := int32min(s.Scale(), o.Scale())

	hlp := s.highLowAtScale(&s.positive, minScale)
	hlp = hlp.with(o.highLowAtScale(&o.positive, minScale))

	hln := s.highLowAtScale(&s.negative, minScale)
	hln = hln.with(o.highLowAtScale(&o.negative, minScale))

	minScale = int32min(
		minScale-changeScale(hlp, s.maxSize),
		minScale-changeScale(hln, s.maxSize),
	)

	s.downscale(s.Scale() - minScale)

	s.mergeBuckets(&s.positive, o, &o.positive, minScale)
	s.mergeBuckets(&s.negative, o, &o.negative, minScale)
}

// mergeBuckets translates index values from another histogram into
// the corresponding buckets of this histogram.
func (s *State[N, Traits]) mergeBuckets(mine *buckets, other *State[N, Traits], theirs *buckets, scale int32) {
	theirOffset := theirs.Offset()
	theirChange := other.Scale() - scale

	for i := uint32(0); i < theirs.Len(); i++ {
		_, success := s.incrementIndexBy(
			mine,
			(theirOffset+int32(i))>>theirChange,
			theirs.At(i),
		)
		if !success {
			panic("incorrect merge scale")
		}
	}
}

// highLowAtScale is an accessory for Merge() to calculate ideal combined scale.
func (s *State[N, Traits]) highLowAtScale(b *buckets, scale int32) highLow {
	if b.Len() == 0 {
		return highLow{
			low:  0,
			high: -1,
		}
	}
	shift := s.Scale() - scale
	return highLow{
		low:  b.indexStart >> shift,
		high: b.indexEnd >> shift,
	}
}

// with is an accessory for Merge() to calculate ideal combined scale.
func (h *highLow) with(o highLow) highLow {
	if o.empty() {
		return *h
	}
	if h.empty() {
		return o
	}
	return highLow{
		low:  int32min(h.low, o.low),
		high: int32max(h.high, o.high),
	}
}

// empty indicates whether there are any values in a highLow
func (h *highLow) empty() bool {
	return h.low > h.high
}

func int32min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func int32max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// bucketsVarwidth[]
//
// Each of the methods below is generic with respect to the underlying
// backing array.  See the interface-level comments.

func (b *bucketsVarwidth[N]) countAt(pos uint32) uint64 {
	return uint64(b.counts[pos])
}

func (b *bucketsVarwidth[N]) reset() {
	for i := range b.counts {
		b.counts[i] = 0
	}
}

func (b *bucketsVarwidth[N]) size() int32 {
	return int32(len(b.counts))
}

func (b *bucketsVarwidth[N]) growTo(newSize, oldPositiveLimit, newPositiveLimit int32) {
	tmp := make([]N, newSize)
	copy(tmp[newPositiveLimit:], b.counts[oldPositiveLimit:])
	copy(tmp[0:oldPositiveLimit], b.counts[0:oldPositiveLimit])
	b.counts = tmp
}

func (b *bucketsVarwidth[N]) reverse(from, limit int32) {
	num := ((from + limit) / 2) - from
	for i := int32(0); i < num; i++ {
		b.counts[from+i], b.counts[limit-i-1] = b.counts[limit-i-1], b.counts[from+i]
	}
}

func (b *bucketsVarwidth[N]) moveCount(src int32) uint64 {
	tmp := b.counts[src]
	b.counts[src] = 0
	return uint64(tmp)
}

func (b *bucketsVarwidth[N]) tryIncrement(bucketIndex int32, incr uint64) bool {
	var limit = uint64(N(0) - 1)
	if uint64(b.counts[bucketIndex])+incr <= limit {
		b.counts[bucketIndex] += N(incr)
		return true
	}
	return false
}

func widenBuckets[From, To bucketsCount](in *bucketsVarwidth[From]) *bucketsVarwidth[To] {
	tmp := make([]To, len(in.counts))
	for i := range in.counts {
		tmp[i] = To(in.counts[i])
	}
	return &bucketsVarwidth[To]{counts: tmp}
}
