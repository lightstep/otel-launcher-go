package fprint

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func mkHeapString(s string) string {
	out, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	var ret string
	err = json.Unmarshal(out, &ret)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestConvert(t *testing.T) {
	t.Run("const", func(t *testing.T) {
		const s = "stryyyyyyyyyyyy"
		l := len(s)

		sl, err := unsafeStringToBytes(s)
		require.NoError(t, err)
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, s, string(sl))
		require.Equal(t, "stryyyyyyyyyyyy", s)
	})
	t.Run("var", func(t *testing.T) {
		s := "stryyyyyyyyyyyy"
		l := len(s)

		sl, err := unsafeStringToBytes(s)
		require.NoError(t, err)
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, s, string(sl))
		require.Equal(t, "stryyyyyyyyyyyy", s)
	})
	t.Run("anonymous", func(t *testing.T) {
		sl, err := unsafeStringToBytes("gizmo")
		require.NoError(t, err)
		l := 5
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "gizmo", string(sl))
	})
	t.Run("ptr", func(t *testing.T) {
		ptr := new(string)
		*ptr = "hello"

		l := 5

		sl, err := unsafeStringToBytes(*ptr)
		require.NoError(t, err)
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "hello", string(sl))
		require.Equal(t, "hello", *ptr)
	})
	t.Run("empty", func(t *testing.T) {
		sl, err := unsafeStringToBytes("")
		require.NoError(t, err)
		l := 0
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "", string(sl))
	})
	t.Run("string on the heap", func(t *testing.T) {
		s := mkHeapString("foo")
		l := 3

		sl, err := unsafeStringToBytes(s)
		require.NoError(t, err)
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "foo", string(sl))
		require.Equal(t, "foo", s)
	})
	t.Run("evil mutation", func(t *testing.T) {
		// If this test ever fails, just remove it.
		// We're deliberately breaking the rules
		// to show what you should not do.

		s := mkHeapString("foo")
		l := 3

		sl, err := unsafeStringToBytes(s)
		require.NoError(t, err)
		runtime.GC()
		sl[0] = 'x'
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "xoo", string(sl))
		require.Equal(t, "xoo", s)
	})
	t.Run("s is not live later on", func(t *testing.T) {
		s := mkHeapString("foo")
		l := 3

		sl, err := unsafeStringToBytes(s)
		require.NoError(t, err)
		// s is not live after the previous line, so the backing array
		// could be garbage collected if we got our
		// pointers wrong.
		runtime.GC()
		require.Equal(t, l, len(sl))
		require.Equal(t, l, cap(sl))
		require.Equal(t, "foo", string(sl))
	})
}
