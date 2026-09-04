package core

import (
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

func TestExperiments(t *testing.T) {
	given := `
	#2 + [a b]
	# gb7
	2 + [e3 a3] + [0 1 2]
	2 + [g2 & f2] + [0 1 2]
	e2:[3 4 3]
	[c2 d2]:3

	# T+ [e2 c3 d4] + [0 0 1 2]
	# c2 c4 @chord(e4)
	`
	t.Run("dev", func(t *testing.T) {
		frames, groups := scan(given)
		fmt.Printf("FRAMES: %v\n", frames)
		fmt.Printf("GROUPS: %v\n", groups)

		expressions := flat(frames, groups)
		fmt.Printf("EXPRESSIONS: %v\n", expressions)
		for _, expr := range expressions {
			fmt.Printf("%v\n", expr)
		}
	})
}

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
				{"§", "+", "2"},
				{"§", "*", "3", "+", "5"},
				{"c", "+", "§"},
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
			[a & b] 2
			[a b] [c:2 d|3]
			c [x y z] d
			`,
			frames: [][]string{
				{"§"}, {"2"},
				{"§"}, {"§"},
				{"c"}, {"§"}, {"d"},
			},
			groups: [][]string{
				{"a", "&", "b"},
				{"a", "b"}, {"c", ":", "2", "d", "|", "3"},
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
				{"§", "+", "§"},
				{"e2"}, {"2", "+", "§", "+", "§"}, {"d3"},
				{"§", "*", "2"}, {"a3"},
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
				{"@", "chord", "(", "e2", "7", ")"}, {"§"},
				{"§"}, {"@", "chord", "(", "e2", ")"},
				{"$", "bass", "(", ")"}, {"20"}, {"§"},
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
				{"§"},
				{"§", "+", "§", "%", "§"},
				{"10", "+", "§", "*", "3"}, {"§"},
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
			testutils.UseFlags(t)
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
			// 515.1 ns/op	     720 B/op	      11 allocs/op
			name: "Simple sequence",
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
			`,
		},
		{
			// 751.5 ns/op	     976 B/op	       9 allocs/op
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
			// 535.3 ns/op	     736 B/op	      10 allocs/op
			name: "Value + operator + value",
			given: `
			[a b] + 2
			[x y] * 3 + 5
			c + [0 1 2]
			`,
		},
		{
			// 685.2 ns/op	     912 B/op	      16 allocs/op
			name: "Value value",
			given: `
			[a & b] 2
			[a b] [c:2 d|3]
			c [x y z] d
			`,
		},
		{
			// 817.3 ns/op	    1072 B/op	      16 allocs/op
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
		},
		{
			// 775.2 ns/op	     928 B/op	      14 allocs/op
			name: "With functions",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
		},
		{
			// 781.8 ns/op	    1056 B/op	      15 allocs/op
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

func TestFlat(t *testing.T) {
	f := strings.Fields
	tests := []struct {
		name     string
		frames   [][]string
		groups   [][]string
		expected [][]string
	}{
		{
			name: "Dev",
			frames: [][]string{
				{"2", "+", "§", "+", "§"},
				{"§", "+", "§", "*", ".5"},
				{"e2", ":", "§"},
				{"§", ":", "3"},
			},
			groups: [][]string{
				{"e3", "a3"}, {"0", "1", "2"},
				{"g2", "&", "f2"}, {"3", "4", "5"},
				{"3", "4", "3"},
				{"c2", "d2"},
			},
			expected: [][]string{
				f("2 + e3 + 0"), f("2 + a3 + 0"),
				f("2 + e3 + 1"), f("2 + a3 + 1"),
				f("2 + e3 + 2"), f("2 + a3 + 2"),

				f("g2 + 3 * .5"), {"&"}, f("f2 + 3 * .5"),
				f("g2 + 4 * .5"), {"&"}, f("f2 + 4 * .5"),
				f("g2 + 5 * .5"), {"&"}, f("f2 + 5 * .5"),

				f("e2 : 3"), f("e2 : 4"), f("e2 : 3"),
				f("c2 : 3"), f("d2 : 3"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			actual := flat(tc.frames, tc.groups)

			testutils.AssertInts(t, len(tc.expected), len(actual))
			for i, act := range actual {
				exp := tc.expected[i]
				testutils.AssertSlices(t, exp, act)
			}
		})
	}
}

func BenchmarkFlat(b *testing.B) {
	tests := []struct {
		name   string
		frames [][]string
		groups [][]string
	}{
		{
			// 555.3 ns/op	    1632 B/op	      15 allocs/op
			name: "For comparison with old expand",
			frames: [][]string{
				{"2", "+", "§", "+", "§"},
			},
			groups: [][]string{
				{"e2", "&", "b2"}, {"5", "8", "0", "3"},
			},
		},
		{
			// 853.7 ns/op	    2720 B/op	      22 allocs/op
			name: "Dev",
			frames: [][]string{
				{"2", "+", "§", "+", "§"},
				{"§", "+", "§", "*", ".5"},
				{"e2", ":", "§"},
				{"§", ":", "3"},
			},
			groups: [][]string{
				{"e3", "a3"}, {"0", "1", "2"},
				{"g2", "&", "f2"}, {"3", "4", "5"},
				{"3", "4", "3"},
				{"c2", "d2"},
			},
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(tc.frames, tc.groups)
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
