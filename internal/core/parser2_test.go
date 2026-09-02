package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

/*
[0] a, +, 1, |, 1
[1] b, +, 2
[2] c, +, 3, :, 7, |, 2
[3] 10
[4] $, bass
[5] 20
[6] @, chord, (, e2, 7, )
[7] 30
[8] a, :, 2
[9] $, lead
[10] b, :, 8, +, [, 0, 4, :, 2, 7, +, 2, 11, ], +, [, 0, 1, 3, 5, ]
[11] @, chord, {, tone, e2, vol, 0.4, }
[12] c, :, 3

// most recent test
[0] a, +, b
[1] 10
[2] $, bass
[3] 20
[4] @, cho
[5] (, e2, 7, ), :, 2
[6] 30
[7] a5
[8] 20
[9] :

[0] a, +, b
[1] 10, $
[2] bass
[3] 20, @
[4] cho, (, e2, 7, ), :, 2
[5] 30
[6] a5
[7] 20
[8] :
*/

// kinda correct one
// [[10] [$ bass] [20] [@ cho] [( e2 7 ) : 2] [30] [a5] [20] [:]]

func TestSplit2(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "All characters used",
			given: `
			#[a b c] * [1 2 3]
			[a b c] + 2 [1 2 3]+x

			a + b & $lead c

			10 $bass 20 @cho(e2 7):2 30 a5

			20

			 a+1|1 b+2 c+3:7|2
			# a:2 $lead() b:8 + [0 4:2 7 + 2 11] + [0 1 3 5]+2 @chord{tone e2 vol 0.4} c:3

				   :

			`,
			expected: [][]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := split(tc.given)
			fmt.Printf("ACT\n%v\n", actual)
			for i, act := range actual {
				fmt.Printf("[%d] %s\n", i, strings.Join(act, ", "))
			}
		})
	}
}

func TestSplit3(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "-All characters used",
			given: `
			e2 2 + [a b c] + [0 1] d3
			# [a b c] + 2 [1 2 3]+x
			# [a b c] [1 2 3]+x
			# a + b & $lead() c
			# 10 $bass() 20 @cho(e2 7):2 30 a5
			`,
			expected: [][]string{},
		},
		{
			name: "Value + operator + value (one expression)",
			given: `
			[a b] + 2
			[a b] * 3 + 5
			c + [0 1]
			`,
		},
		{
			name: "Value + value (should split)",
			given: `
			[a b] 2
			[a b] [c d]
			c [a b] d
			`,
		},
		{
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
		},
		{
			name: "With functions",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
		},
		{
			name: "Edge cases",
			given: `
			[a b]
			[a b] + [c d] + [e f]
			10 + [1 2] * 3 [4 5]
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual, sub := normalize(tc.given)
			fmt.Printf("INPUT\n%v\n", tc.given)
			fmt.Printf("ACT\n%v\n", actual)
			fmt.Printf("SUB\n%v\n", sub)
			// fmt.Println("List")
			// for i, act := range actual {
			// 	fmt.Printf("[%d] %s\n", i, strings.Join(act, ", "))
			// }
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "All characters used",
			given: `
			b:8 + [0 4:2 7 + 2 11] + [0 1 3 5]

			a+1 b+2 c+3
			d+4 [5 6 7] + 2
			10 $bass 20 @chord 30

			b:8 + [0:8 4 7+2 11] :2

			b:8 + [0:8 4 7 + 2 11]
			[x y] * [a & b] + 2
			3 * [a & b] + [1 2 3]
			[0 2 7 2] + [e2 & b2]
			2 + [e2 & b2] + [5 8 0 3]
			2 + [a & b] + [5 7 8 3]
			a b c 2 + [a & b] + [1 2&3] * [1 1 0 1]
			___ a b +2

			-2+10 &
			10+-2
			2+a*3 b/8 c+b a & b c |
			20 30 40 @chord(a b) a b
			7 + 8 @song {n 60 vol 1.2}
			a b c 2 + [a&b] * [1 2 | | 3]
			`,
			expected: [][]string{},
		},
	}

	// a b c 2 + [a&b] * [1 2 3]

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := split(tc.given)

			// expanded := expand(actual[0])
			// fmt.Printf("INPUT %v\n", actual[0])
			// fmt.Printf("EXPANDED %v\n", expanded)
			// for _, exp := range expanded {
			// 	fmt.Printf("%s\n", strings.Join(exp, ""))
			// }

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
			// 488.9 ns/op	     624 B/op	      10 allocs/op
			// 517.4 ns/op	     720 B/op	      11 allocs/op
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
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
				// split(tc.given)
				normalize(tc.given)
			}
		})
	}
}

func BenchmarkExpand(b *testing.B) {
	tests := []struct{ given string }{
		{
			// 81.29 ns/op	     240 B/op	       2 allocs/op
			given: "a + 2",
		},
		{
			// 127.4 ns/op	     352 B/op	       4 allocs/op
			given: "[a b] + c",
		},
		{
			// 182.0 ns/op	     496 B/op	       5 allocs/op
			given: "2 + [a & b] + c",
		},
		{
			// 608.3 ns/op	    1664 B/op	      16 allocs/op
			given: "2 + [e2 & b2] + [5 8 0 3]",
		},
		{
			// 2434 ns/op	    8832 B/op	      55 allocs/op
			given: "4 + [a b c] + [1 2 3 4] * [1 1 0 1]",
		},
		{
			// 4291 ns/op	   16688 B/op	      90 allocs/op
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
