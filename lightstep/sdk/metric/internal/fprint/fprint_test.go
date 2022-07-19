package fprint

import (
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestBasic(t *testing.T) {
	var fp uint64 = 0xc43fb29ab5effcfe // for "foobar"

	f := func() error {
		for i := 0; i < 100; i++ {
			ss := []string{
				"foobar",
				mkHeapString("foobar"),
			}
			runtime.GC()
			for _, s := range ss {
				runtime.GC()
				// Perform a bunch of fingerprints and check that they match the golden value.
				for i, res := range []uint64{
					FingerprintString64(s),
					fingerprintString64(s),
					unsafeFingerprintString64(s),
				} {
					if res != fp {
						return fmt.Errorf("%d: %v != %v", i, res, fp)
					}
				}
			}
		}
		return nil
	}

	var eg errgroup.Group
	for i := 0; i < 10; i++ {
		eg.Go(f)
	}

	require.NoError(t, eg.Wait())
}

func BenchmarkScrambleBitsAvalancheEffect(b *testing.B) {
	// An ideal implementation of ScrambleBits has good adherence to the
	// "avalanche effect" (if a single bit is flipped in an input, the
	// output would probabilistically have half its bits flipped).
	// This benchmark tests the "Strict Avalanche Criterion" of a
	// number of possible implementations of ScrambleBits.

	kases := []struct {
		name string
		f    func(uint64) uint64
	}{
		{
			name: "ScrambleBits",
			f:    ScrambleBits,
		},
		{
			name: "mix(u, u)",
			f:    func(u uint64) uint64 { return mix(u, u) },
		},
		{
			name: "mix(u, 0xff51afd7ed558ccd) ",
			f:    func(u uint64) uint64 { return mix(u, 0xff51afd7ed558ccd) },
		},
		{
			name: "MurmurHash3/fmix",
			f: func(key uint64) uint64 {
				// https://github.com/aappleby/smhasher/wiki/MurmurHash3/34a5fb13c94754a997b8484aaaf4184a6bce8cbb
				key ^= (key >> 33)
				key *= 0xff51afd7ed558ccd
				key ^= (key >> 33)
				key *= 0xc4ceb9fe1a85ec53
				key ^= (key >> 33)

				return key
			},
		},
		{
			name: "identity (sanity check)",
			f:    func(u uint64) uint64 { return u },
		},
	}

	for _, kase := range kases {
		mean := 0
		variance := 0
		numTrials := 10000000 // 1e7

		for i := 0; i < numTrials; i++ {
			x := rand.Uint64()

			// hash of the input sample 'x'
			h := kase.f(x)

			// walk all single bit flips of 'x'
			for j := 0; j < 64; j++ {
				pop := bits.OnesCount64(h ^ kase.f(x^(1<<j)))
				mean += pop
				variance += (pop - 32) * (pop - 32) // expected mean = 32
			}
		}
		b.Logf("name: %v, mean flips: %v, std dev: %v", kase.name, float64(mean)/float64(numTrials*64), math.Sqrt(float64(variance)/float64(numTrials*64)))
	}
}
