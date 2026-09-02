package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
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
			actual, sub := scan(tc.given)
			fmt.Printf("INPUT\n%v\n", tc.given)
			fmt.Printf("ACT\n%v\n", actual)
			fmt.Printf("SUB\n%v\n", sub)
			fmt.Println("List")
			for i, act := range actual {
				fmt.Printf("[%d] %s\n", i, strings.Join(act, ", "))
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
				scan(tc.given)
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
		splitted, _ := scan(tc.given)
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				expand(splitted[0])
			}
		})
	}
}
