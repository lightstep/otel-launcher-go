package fprint

import (
	"math"
	"time"

	// Our use of farmhash is sort of arbitrary: we want a fast,
	// fingerprint function and farmhash is familiar (to
	// Lightstep).  We do not require a fixed (never-changing)
	// hash function here althought this is.
	farm "github.com/dgryski/go-farm"
)

// Mix combines multiple fingerprints together.
func Mix(is ...uint64) uint64 {
	if len(is) == 0 {
		return 0
	}
	accumulator := is[0]
	for _, i := range is[1:] {
		accumulator = mix(accumulator, i)
	}
	return accumulator
}

// Borrowed from farmhash.
func mix(x uint64, y uint64) uint64 {
	const mul uint64 = 0x9ddfea08eb382d69
	a := (x ^ y) * mul
	a ^= a >> 47
	b := (y ^ a) * mul
	b ^= b >> 47
	b *= mul
	return b
}

func Fingerprint64(s []byte) uint64 {
	return farm.Fingerprint64(s)
}

func FingerprintUint64(i uint64) uint64 {
	return i
}

func FingerprintInt64(i int64) uint64 {
	return uint64(i)
}

func FingerprintInt(i int) uint64 {
	return uint64(i)
}

func FingerprintBool(b bool) uint64 {
	if b {
		return uint64(1)
	}
	return uint64(0)
}

func FingerprintFloat64(f float64) uint64 {
	return math.Float64bits(f)
}

func unsafeFingerprintString(s string) uint64 {
	bs, err := unsafeStringToBytes(s)
	if err != nil {
		// Gnarly! There should be an attribute size limit before this happens.
		bs = []byte(s)
	}
	return Fingerprint64(bs)
}

func FingerprintString(s string) uint64 {
	// We know that the go-farm implementation
	// we use does not modify the []byte it is passed,
	// so we use an unsafe conversion here from string to
	// []byte to avoid a copy.
	return unsafeFingerprintString(s)
}

func FingerprintTime64(t time.Time) uint64 {
	return uint64(t.UnixNano())
}
