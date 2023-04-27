package aggregation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTemporality(t *testing.T) {
	for _, test := range []struct {
		input string
		tempo Temporality
		ok    bool
	}{
		{"cumulative", CumulativeTemporality, true},
		{"Cumulative", CumulativeTemporality, true},
		{"delta", DeltaTemporality, true},
		{"DELTA", DeltaTemporality, true},
		{"other", UndefinedTemporality, false},
		{"", UndefinedTemporality, false},
	} {
		tempo, ok := ParseTemporality(test.input)
		require.Equal(t, test.tempo, tempo)
		require.Equal(t, test.ok, ok)
	}
}
