package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"testing"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name   string
		given  string
		frames [][]string
		groups [][]string
	}{
		{
			name: "Simple sequence",
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
			`,
			frames: [][]string{
				{"2", "+", "a", "*", "3"}, {"b", "/", "8"}, {"c", "+", "b"},
				{"a"}, {"b"}, {"c"}, {"2", "+", "D", "*", "2"},
			},
		},
		{
			name: "Simple sequence with comments",
			given: `
			# eeee aaaa ff ff
			e2:4 a2:4 f2:2|2
			# run bass intro
			$bass(0.77 12)
			@intro{vol 0.5 tone g3}
			`,
			frames: [][]string{
				{"e2", ":", "4"}, {"a2", ":", "4"}, {"f2", ":", "2", "|", "2"},
				{"$", "bass", "(", "0.77", "12", ")"},
				{"@", "intro", "{", "vol", "0.5", "tone", "g3", "}"},
			},
		},
		{
			name: "Value + operator + value",
			given: `
			[a b] + 2
			[x y] * 3 + 5
			c + [0 1 2]
			`,
			frames: [][]string{
				{"§0", "+", "2"},
				{"§1", "*", "3", "+", "5"},
				{"c", "+", "§2"},
			},
			groups: [][]string{
				{"a", "b"},
				{"x", "y"},
				{"0", "1", "2"},
			},
		},
		{
			name: "Value value",
			given: `
			[a b] 2
			[a b] [c d]
			c [x y z] d
			`,
			frames: [][]string{
				{"§0"}, {"2"},
				{"§1"}, {"§2"},
				{"c"}, {"§3"}, {"d"},
			},
			groups: [][]string{
				{"a", "b"},
				{"a", "b"}, {"c", "d"},
				{"x", "y", "z"},
			},
		},
		{
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
			frames: [][]string{
				{"§0", "+", "§1"},
				{"e2"}, {"2", "+", "§2", "+", "§3"}, {"d3"},
				{"§4", "*", "2"}, {"a3"},
			},
			groups: [][]string{
				{"a", "b"}, {"c", "d"},
				{"a", "b", "c"}, {"0", "1"},
				{"x", "y"},
			},
		},
		{
			name: "With functions",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
			frames: [][]string{
				{"@", "chord", "(", "e2", "7", ")"}, {"§0"},
				{"§1"}, {"@", "chord", "(", "e2", ")"},
				{"$", "bass", "(", ")"}, {"20"}, {"§2"},
			},
			groups: [][]string{
				{"a", "b"},
				{"a", "b"},
				{"0", "4", "7"},
			},
		},
		{
			name: "Edge cases",
			given: `
			[a b]
			[a b] + [c d] % [e f]
			10 + [1 2] * 3 [4 5]
			`,
			frames: [][]string{
				{"§0"},
				{"§1", "+", "§2", "%", "§3"},
				{"10", "+", "§4", "*", "3"}, {"§5"},
			},
			groups: [][]string{
				{"a", "b"},
				{"a", "b"}, {"c", "d"}, {"e", "f"},
				{"1", "2"}, {"4", "5"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			frames, groups := scan(tc.given)

			testutils.AssertInts(t, len(tc.frames), len(frames))
			for i, act := range frames {
				exp := tc.frames[i]
				testutils.AssertSlices(t, exp, act)
			}

			testutils.AssertInts(t, len(tc.groups), len(groups))
			for i, act := range groups {
				exp := tc.groups[i]
				testutils.AssertSlices(t, exp, act)
			}
		})
	}
}

func BenchmarkScan(b *testing.B) {
	tests := []struct {
		name  string
		given string
	}{
		{
			// 527.0 ns/op	     720 B/op	      11 allocs/op
			name: "Simple sequence",
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
			`,
		},
		{
			// 752.4 ns/op	     976 B/op	       9 allocs/op
			name: "Simple sequence with comments",
			given: `
			# eeee aaaa ff ff
			e2:4 a2:4 f2:2|2
			# run bass intro
			$bass(0.77 12)
			@intro{vol 0.5 tone g3}
			`,
		},
		{
			// 688.9 ns/op	     745 B/op	      13 allocs/op
			name: "Value + operator + value",
			given: `
			[a b] + 2
			[x y] * 3 + 5
			c + [0 1 2]
			`,
		},
		{
			// 811.9 ns/op	     685 B/op	      19 allocs/op
			name: "Value value",
			given: `
			[a b] 2
			[a b] [c d]
			c [x y z] d
			`,
		},
		{
			// 1064 ns/op	    1088 B/op	      21 allocs/op
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
		},
		{
			// 916.4 ns/op	     937 B/op	      17 allocs/op
			name: "With functions",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
		},
		{
			// 1091 ns/op	    1075 B/op	      21 allocs/op
			name: "Edge cases",
			given: `
			[a b]
			[a b] + [c d] % [e f]
			10 + [1 2] * 3 [4 5]
			`,
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
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
