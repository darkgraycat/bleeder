package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "All characters used",
			given: `
			2 + [a | | b] |
			2 + [a & b] + [5 7 8 3]
			# a b c 2 + [a & b] + [1 2&3] * [1 1 0 1]
			# ___ a b +2

			# -2+10 &
			# 10+-2
			# 2+a*3 b/8 c+b a & b c |
			# 20 30 40 @chord(a b) a b
			# 7 + 8 @song {n 60 vol 1.2}
			# a b c 2 + [a&b] * [1 2 | | 3]
			`,
			expected: [][]string{},
		},
	}

	// a b c 2 + [a&b] * [1 2 3]

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := tokenize(tc.given)

			expand(actual[0])

			// testutils.AssertInts(t, len(tc.expected), len(actual))
			for i, act := range actual {
				fmt.Printf("[%d] %s\n", i, strings.Join(act, ", "))
				// testutils.AssertSlices(t, tc.expected[i], act)
			}
		})
	}
}

func BenchmarkTokenize(b *testing.B) {
	tests := []struct{ given string }{
		{
			// 634.3 ns/op	     784 B/op	      10 allocs/op
			given: `
			2+a*3 b/8 c+b
			a b c 2 + [a&b] * [1 2 3]
			`,
		},
	}

	/*
		[a b c] +[0 0 7 0] &
		@voices(e2 .8) & @voices{tone:e3 gain:.6}
		[2 2*2 5 7] |&|+7 ___ |+4
	*/

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				tokenize(tc.given)
			}
		})
	}
}
