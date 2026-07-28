package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "Simple multiline sequence",
			given: `
			c4&eb4 60:2 |+8
			@chord:as2 & a2 gs2
			`,
			expected: [][]string{
				{"c4", "&", "eb4", "60:2", "|+8"},
				{"@chord:as2", "&", "a2", "gs2"},
			},
		},
		{
			name: "Complex multiline sequence",
			given: `
			0-12:8 & - 4 - 11
			0-10:8 & 2 | | |+2
			0:8:.5 0 # 1 2 3
			`,
			expected: [][]string{
				{"0-12:8", "&", "-", "4", "-", "11"},
				{"0-10:8", "&", "2", "|", "|", "|+2"},
				{"0:8:.5", "0"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := tokenize(tc.given)
			testutils.AssertInts(t, len(tc.expected), len(actual))
			for i, act := range actual {
				testutils.AssertSlices(t, tc.expected[i], act)
			}
		})
	}
}

func BenchmarkTokenize(b *testing.B) {
	tests := []struct{ given string }{
		{
			// old
			// 355.3 ns/op	     320 B/op	       4 allocs/op
			given: `
			0-12:8 & - 4 - 11
			0-10:8 & 2 | | |+2
			0:8:.5 0 # 1 2 3
			`,
		},
		{
			// 257.7 ns/op	     368 B/op	       5 allocs/op
			given: `
			c4&eb4 60:2 |+8
			@chord:as2 & a2 gs2
			`,
		},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				tokenize(tc.given)
			}
		})
	}
}
