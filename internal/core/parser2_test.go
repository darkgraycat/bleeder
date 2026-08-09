package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "All characters used",
			given: `
			2 + [e2 & b2] + [5 8 0 3]
			# 2 + [a & b] + [5 7 8 3]
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
			actual := split(tc.given)

			expanded := expand(actual[0])
			fmt.Printf("INPUT %v\n", actual[0])
			fmt.Printf("EXPANDED %v\n", expanded)
			for _, exp := range expanded {
				fmt.Printf("%s\n", strings.Join(exp, ""))
			}

			// testutils.AssertInts(t, len(tc.expected), len(actual))
			for i, act := range actual {
				fmt.Printf("[%d] %s\n", i, strings.Join(act, ", "))
				// testutils.AssertSlices(t, tc.expected[i], act)
			}
		})
	}
}

func BenchmarkSplit(b *testing.B) {
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
				split(tc.given)
			}
		})
	}
}

func BenchmarkExpand(b *testing.B) {
	tests := []struct{ given string }{
		{
			// 83.64 ns/op	     240 B/op	       2 allocs/op
			given: "a + 2",
		},
		{
			// 158.0 ns/op	     355 B/op	       5 allocs/op
			given: "[a b] + c",
		},
		{
			// 205.9 ns/op	     499 B/op	       6 allocs/op
			given: "2 + [a & b] + c",
		},
		{
			// 654.6 ns/op	    1670 B/op	      18 allocs/op
			given: "2 + [e2 & b2] + [5 8 0 3]",
		},
		{
			// 2486 ns/op	    8841 B/op	      58 allocs/op
			given: "4 + [a b c] + [1 2 3 4] * [1 1 0 1]",
		},
		{
			// 4348 ns/op	   16700 B/op	      94 allocs/op
			given: "[a b c] + [1 2 3] * [x y z] / [.7 .8 .9]",
		},
	}

	for i, tc := range tests {
		splitted := split(tc.given)
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				expand(splitted[0])
			}
		})
	}
}
