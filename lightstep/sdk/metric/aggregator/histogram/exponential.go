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

// SynchronizedMove implements export.Aggregator.
func (s *State[N, Traits]) SynchronizedMove(dest *State[N, Traits]) error {
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

	return nil
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

	if value == 0 {
		s.count += incr
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

	s.count += incr

	s.update(b, value, incr)
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

// changeScale computes the required change of scale.
//
// sizeReq = (high-low+1) is the minimum size needed to fit the new
// index at the current scale, i.e., the distance to the more-distant
// extreme inclusive bucket.  We have that:
//
//   sizeReq >= maxSize
//
// Compute the shift equal to the number of times sizeReq must be
// divided by two before sizeReq < maxSize.
//
// Note: this can be computed in a conservative way w/o use of a loop,
// e.g.,
//
//   shift := 64-bits.LeadingZeros64((high-low+1)/int64(a.maxSize))
//
// however this under-counts by 1 some of the time depending on
// alignment.
func changeScale(hl highLow, size int32) int32 {
	var change int32
	for hl.high-hl.low >= size {
		hl.high >>= 1
		hl.low >>= 1
		change++
	}
	return change
}

// size() reflects the allocated size of the array, not to be confused
// with Len() which is the range of non-zero values.
func (b *buckets) size() int32 {
	switch counts := b.backing.(type) {
	case []uint8:
		return int32(len(counts))
	case []uint16:
		return int32(len(counts))
	case []uint32:
		return int32(len(counts))
	case []uint64:
		return int32(len(counts))
	}
	return 0
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
	if b.Len() == 0 {
		if b.backing == nil {
			b.backing = []uint8{0}
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
		} else if span >= b.size() {
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
		} else if span >= b.size() {
			s.grow(b, span+1)
		}
		b.indexEnd = index
	}

	bucketIndex := index - b.indexBase
	if bucketIndex < 0 {
		bucketIndex += b.size()
	}
	b.incrementBucket(bucketIndex, incr)
	return highLow{}, true
}

// grow resizes the backing array by doubling in size up to maxSize.
// this extends the array with a bunch of zeros and copies the
// existing counts to the same position.
func (s *State[N, Traits]) grow(b *buckets, needed int32) {
	size := b.size()
	bias := b.indexBase - b.indexStart
	diff := size - bias
	growTo := int32(1) << (32 - bits.LeadingZeros32(uint32(needed)))
	if growTo > s.maxSize {
		growTo = s.maxSize
	}
	part := growTo - bias
	switch counts := b.backing.(type) {
	case []uint8:
		tmp := make([]uint8, growTo)
		copy(tmp[part:], counts[diff:])
		copy(tmp[0:diff], counts[0:diff])
		b.backing = tmp
	case []uint16:
		tmp := make([]uint16, growTo)
		copy(tmp[part:], counts[diff:])
		copy(tmp[0:diff], counts[0:diff])
		b.backing = tmp
	case []uint32:
		tmp := make([]uint32, growTo)
		copy(tmp[part:], counts[diff:])
		copy(tmp[0:diff], counts[0:diff])
		b.backing = tmp
	case []uint64:
		tmp := make([]uint64, growTo)
		copy(tmp[part:], counts[diff:])
		copy(tmp[0:diff], counts[0:diff])
		b.backing = tmp
	default:
		panic("grow() with size() == 0")
	}
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

	b.reverse(0, b.size())
	b.reverse(0, bias)
	b.reverse(bias, b.size())
}

func (b *buckets) reverse(from, limit int32) {
	num := ((from + limit) / 2) - from
	switch counts := b.backing.(type) {
	case []uint8:
		for i := int32(0); i < num; i++ {
			counts[from+i], counts[limit-i-1] = counts[limit-i-1], counts[from+i]
		}
	case []uint16:
		for i := int32(0); i < num; i++ {
			counts[from+i], counts[limit-i-1] = counts[limit-i-1], counts[from+i]
		}
	case []uint32:
		for i := int32(0); i < num; i++ {
			counts[from+i], counts[limit-i-1] = counts[limit-i-1], counts[from+i]
		}
	case []uint64:
		for i := int32(0); i < num; i++ {
			counts[from+i], counts[limit-i-1] = counts[limit-i-1], counts[from+i]
		}
	}
}

// relocateBucket adds the count in counts[src] to counts[dest] and
// resets count[src] to zero.
func (b *buckets) relocateBucket(dest, src int32) {
	if dest == src {
		return
	}
	switch counts := b.backing.(type) {
	case []uint8:
		tmp := counts[src]
		counts[src] = 0
		b.incrementBucket(dest, uint64(tmp))
	case []uint16:
		tmp := counts[src]
		counts[src] = 0
		b.incrementBucket(dest, uint64(tmp))
	case []uint32:
		tmp := counts[src]
		counts[src] = 0
		b.incrementBucket(dest, uint64(tmp))
	case []uint64:
		tmp := counts[src]
		counts[src] = 0
		b.incrementBucket(dest, uint64(tmp))
	}
}

// incrementBucket increments the backing array index by `incr`.
func (b *buckets) incrementBucket(bucketIndex int32, incr uint64) {
	for {
		switch counts := b.backing.(type) {
		case []uint8:
			if uint64(counts[bucketIndex])+incr < 0x100 {
				counts[bucketIndex] += uint8(incr)
				return
			}
			tmp := make([]uint16, len(counts))
			for i := range counts {
				tmp[i] = uint16(counts[i])
			}
			b.backing = tmp
			continue
		case []uint16:
			if uint64(counts[bucketIndex])+incr < 0x10000 {
				counts[bucketIndex] += uint16(incr)
				return
			}
			tmp := make([]uint32, len(counts))
			for i := range counts {
				tmp[i] = uint32(counts[i])
			}
			b.backing = tmp
			continue
		case []uint32:
			if uint64(counts[bucketIndex])+incr < 0x100000000 {
				counts[bucketIndex] += uint32(incr)
				return
			}
			tmp := make([]uint64, len(counts))
			for i := range counts {
				tmp[i] = uint64(counts[i])
			}
			b.backing = tmp
			continue
		case []uint64:
			counts[bucketIndex] += incr
			return
		default:
			panic("increment with nil slice")
		}
	}
}

// Merge combines two histograms that have the same buckets into a single one.
func (s *State[N, Traits]) Merge(o *State[N, Traits]) error {

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

	return nil
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

func (h *highLow) empty() bool {
	return h.low > h.high
}
