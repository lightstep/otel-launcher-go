package fprint

import (
	"fmt"
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
					FingerprintString(s),
					unsafeFingerprintString(s),
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
