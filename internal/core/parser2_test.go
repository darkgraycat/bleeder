package core

import (
	"bleeder/internal/shared"
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

func TestExperiments(t *testing.T) {
	tests := []string{
		"2 + [a b] * [c d e] + 3",
		// "2 + [a b] 7 [c d] + 3",
		// "e2 g3 a + 7 b * [a b] d3 d6",
		// "2 + [a b] 3 + [c d]",
		// "2 + [a b] * [8 9 0] : T",
		// "[a & b] + [2 4]",
		// "[a b+3] + [2 4] : 8",
		// "e2|3 b3:2|2 & d2",
		// "@song1(e2 2):8+T",
		// "[@chord(g3)]:4",
		// "2 + [e2 2]",
		// "[e2 d2 c3]:3",
		// "[a b] & [c d]",
		// "[e2+T a2:d] * 2",
		// "[a b] + [0 1] * [I J]",
		// "@song2(e2 2) + [0 0]",
		// "[@chord{note g3 dur 7} @mysong(g2 a4)] : 8 + [a b] * 3",
	}

	for _, tc := range tests {
		t.Run("experiments", func(t *testing.T) {
			out1 := split(tc)
			out2 := plain2(out1)
			fmt.Printf("INPUT: %v\n", tc)
			fmt.Printf("SPLIT: %v\n", out1)
			fmt.Printf("PLAIN: %v\n", out2)
			// for i, expr := range out {
			// 	fmt.Printf("[%d] ", i)
			// 	for _, tok := range expr {
			// 		fmt.Printf("%s", tok)
			// 	}
			// 	fmt.Println()
			// }

			// frames, groups := scan(tc)
			// fmt.Printf("FRAMES: %v\n", frames)
			// fmt.Printf("GROUPS: %v\n", groups)
			// expressions := flat(frames, groups)
			// fmt.Printf("EXPRESSIONS: %v\n", expressions)
			// for i, expr := range expressions {
			// 	fmt.Printf("[%d] ", i)
			// 	for _, tok := range expr {
			// 		fmt.Printf("%s", tok)
			// 	}
			// 	fmt.Println()
			// }
		})
		fmt.Println()
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected []string
	}{
		{
			name:  "Sequential frames",
			given: "a2 b2 c3+x d4|2:8",
			expected: []string{
				"a2", "b2", "c3 + x", "d4 | 2 : 8",
			},
		},
		{
			name:  "Grouped frames",
			given: "e2 3+[a b]*[1 2+x]:T f2",
			expected: []string{
				"e2", "3 + [", "a", "b", "] * [", "1", "2 + x", "] : T", "f2",
			},
		},
		{
			name:  "Links to sequence",
			given: "@chord(a2) @chord(f2):T",
			expected: []string{
				// want to
				"@ chord (", "a2", ")", " @chord (", "f2", ") : T",
				// current
				// "@", "chord", "(", "a2", ") @", "chord", "(", "f2", ") : T",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			frames := split(tc.given)
			expFrames := shared.Map(tc.expected, strings.Fields)
			testutils.AssertNested(t, expFrames, frames)
		})
	}
}

func BenchmarkSplit(b *testing.B) {
	tests := []struct {
		name  string
		given string
	}{

		// 264.1 ns/op	     376 B/op	       6 allocs/op
		{name: "Sequential frames", given: "a2 b2 c3+x d4|2:8"},
		// 510.6 ns/op	     944 B/op	      13 allocs/op
		{name: "Grouped frames", given: "e2 3+[a b]*[1 2+x]:T f2"},
		// 486.9 ns/op	     864 B/op	      14 allocs/op
		{name: "Links to sequence", given: "@chord(a2) @chord(f2):T"},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				split(tc.given)
			}
		})
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		name   string
		given  string
		frames []string
		groups []string
	}{
		{
			name: "Simple sequence",
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
			`,
			frames: []string{
				"2 + a * 3", "b / 8", "c + b",
				"a", "b", "c", "2 + D * 2",
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
			frames: []string{
				"e2 : 4", "a2 : 4", "f2 : 2 | 2",
				"$ bass ( 0.77 12 )",
				"@ intro { vol 0.5 tone g3 }",
			},
		},
		{
			name: "Value + operator + value",
			given: `
			[a b] + 2
			[x y] * 3 + 5
			c + [0 1 2]
			`,
			frames: []string{
				"§ + 2",
				"§ * 3 + 5",
				"c + §",
			},
			groups: []string{
				"a b",
				"x y",
				"0 1 2",
			},
		},
		{
			name: "Value value",
			given: `
			[a & b] 2
			[a b] [c:2 d|3]
			c [x y z] d
			`,
			frames: []string{
				"§", "2",
				"§", "§",
				"c", "§", "d",
			},
			groups: []string{
				"a & b",
				"a b", "c : 2 d | 3",
				"x y z",
			},
		},
		{
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
			frames: []string{
				"§ + §",
				"e2", "2 + § + §", "d3",
				"§ * 2", "a3",
			},
			groups: []string{
				"a b", "c d",
				"a b c", "0 1",
				"x y",
			},
		},
		{
			name: "With sequences",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
			frames: []string{
				"@ chord ( e2 7 )", "§",
				"§", "@ chord ( e2 )",
				"$ bass ( )", "20", "§",
			},
			groups: []string{
				"a b",
				"a b",
				"0 4 7",
			},
		},
		{
			name: "With operators",
			given: `
			T + [a:2 b:3]:2
			[x y] : [2/d 3/d]
			`,
			frames: []string{
				"T + § : 2",
				"§ : §",
			},
			groups: []string{
				"a : 2 b : 3",
				"x y", "2 / d 3 / d",
			},
		},
		{
			name: "Edge cases",
			given: `
			[a b]
			[a b] + [c d] % [e f]
			10 + [1 2] * 3 [4 5]
			`,
			frames: []string{
				"§",
				"§ + § % §",
				"10 + § * 3", "§",
			},
			groups: []string{
				"a b",
				"a b", "c d", "e f",
				"1 2", "4 5",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)

			frames, groups := scanNew(tc.given)
			expFrames := shared.Map(tc.frames, strings.Fields)
			expGroups := shared.Map(tc.groups, strings.Fields)

			testutils.AssertInts(t, len(expFrames), len(frames))
			for i, act := range frames {
				exp := expFrames[i]
				testutils.AssertSlices(t, exp, act)
			}

			testutils.AssertInts(t, len(expGroups), len(groups))
			for i, act := range groups {
				exp := expGroups[i]
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
			// 413.5 ns/op	     592 B/op	       9 allocs/op
			name: "Simple sequence",
			given: `
			2+a*3 b/8 c+b
			a b c 2 + D * 2
			`,
		},
		{
			// 751.5 ns/op	     976 B/op	       9 allocs/op
			// 489.0 ns/op	     688 B/op	       7 allocs/op
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
			// 412.6 ns/op	     576 B/op	       8 allocs/op
			name: "Value + operator + value",
			given: `
			[a b] + 2
			[x y] * 3 + 5
			c + [0 1 2]
			`,
		},
		{
			// 685.2 ns/op	     912 B/op	      16 allocs/op
			// 559.2 ns/op	     752 B/op	      14 allocs/op
			name: "Value value",
			given: `
			[a & b] 2
			[a b] [c:2 d|3]
			c [x y z] d
			`,
		},
		{
			// 817.3 ns/op	    1072 B/op	      16 allocs/op
			// 654.1 ns/op	     880 B/op	      14 allocs/op
			name: "Mixed",
			given: `
			[a b] + [c d]
			e2 2 + [a b c] + [0 1] d3
			[x y] * 2 a3
			`,
		},
		{
			// 775.2 ns/op	     928 B/op	      14 allocs/op
			// 587.9 ns/op	     704 B/op	      12 allocs/op
			name: "With functions",
			given: `
			@chord(e2 7) [a b]
			[a b] @chord(e2)
			$bass() 20 [0 4 7]
			`,
		},
		{
			// 781.8 ns/op	    1056 B/op	      15 allocs/op
			// 625.7 ns/op	     864 B/op	      13 allocs/op
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
				scanNew(tc.given)
			}
		})
	}
}

func TestFlat(t *testing.T) {
	tests := []struct {
		name     string
		frames   []string
		groups   []string
		expected []string
	}{
		{
			name: "Dev",
			frames: []string{
				"2 + § + §",
				"§ + § * .5",
				"e2 : §",
				"§ : 3",
			},
			groups: []string{
				"e3 a3", "0 1 2",
				"g2 & f2", "3 4 5",
				"3 4 3",
				"c2 d2",
			},
			expected: []string{
				"2 + e3 + 0", "2 + a3 + 0",
				"2 + e3 + 1", "2 + a3 + 1",
				"2 + e3 + 2", "2 + a3 + 2",

				"g2 + 3 * .5", "&", "f2 + 3 * .5",
				"g2 + 4 * .5", "&", "f2 + 4 * .5",
				"g2 + 5 * .5", "&", "f2 + 5 * .5",

				"e2 : 3", "e2 : 4", "e2 : 3",
				"c2 : 3", "d2 : 3",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)

			frames := shared.Map(tc.frames, strings.Fields)
			groups := shared.Map(tc.groups, strings.Fields)
			expected := shared.Map(tc.expected, strings.Fields)

			actual := flat(frames, groups)

			testutils.AssertNested(t, expected, actual)
		})
	}
}

func BenchmarkFlat(b *testing.B) {
	tests := []struct {
		name   string
		frames []string
		groups []string
	}{
		{
			// 620.1 ns/op	    1456 B/op	      16 allocs/op
			name: "For comparison with old expand",
			frames: []string{
				"2 + § + §",
			},
			groups: []string{
				"e2 & b2", "5 8 0 3",
			},
		},
		{
			// 1165 ns/op	    2976 B/op	      26 allocs/op
			name: "Dev",
			frames: []string{
				"2 + § + §",
				"§ + § * .5",
				"e2 : §",
				"§ : 3",
			},
			groups: []string{
				"e3 a3", "0 1 2",
				"g2 § f2", "3 4 5",
				"3 4 3",
				"c2 d2",
			},
		},
	}

	for _, tc := range tests {
		frames := shared.Map(tc.frames, strings.Fields)
		groups := shared.Map(tc.groups, strings.Fields)

		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(frames, groups)
			}
		})
	}
}

func TestScanFlat(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected []string
	}{
		{
			name: "No groups is used",
			given: `
			e2:4 b3:2|2
			f1|3 & c2|2 d2
			`,
			expected: []string{
				"e2 : 4", "b3 : 2 | 2",
				// "f1 | 3", "&", "c2 | 2", "d2",
				// a split+plain:
				"f1 | 3 & c2 | 2", "d2",
			},
		},
		{
			name: "Riff with simple chord",
			given: `
			[e2 & b3] + [0 4 5]
			`,
			expected: []string{
				// "e2 + 0", "&", "b3 + 0",
				// "e2 + 4", "&", "b3 + 4",
				// "e2 + 5", "&", "b3 + 5",
				// a split+plain:
				"e2 & b3 + 0",
				"e2 & b3 + 4",
				"e2 & b3 + 5",
			},
		},
		{
			name: "All features",
			given: `
			e2|3 b3:2|2 & d2 # inline
			@song1(e2 2):8+T # sequence
			[a&b] + [2 4] # simple chord
			[e2 d2 c3]:3 # mod duration
			[a b] & [c d] # two groups
			# [e2+T a2:d] * 2 # groups with operators
			# [a b] + [0 1] * [I J] # triple group
			# [@chord(g3)] +12 # with sequences
			`,
			expected: []string{
				// "e2 | 3", "b3 : 2 | 2", "&", "d2",
				"e2 | 3", "b3 : 2 | 2 & d2", // < split+plain
				"@ song1 ( e2 2 ) : 8 + T",
				// "a + 2", "&", "b + 2", "a + 4", "&", "b + 4",
				"a & b + 2", "a & b + 4", // < split+plain
				"e2 : 3", "d2 : 3", "c3 : 3",
				// "a", "b", "&", "c", "d",
				"a & c", "b & c", "a & d", "b & d", // < split+plain
				// "e2 * 2 + T", "a2 * 2 : d",
				// "a + 0 * I", "b + 0 * I", "a + 1 * I", "b + 1 * I",
				// "a + 0 * J", "b + 0 * J", "a + 1 * J", "b + 1 * J",
				// "@ chord ( g3 ) + 12",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			expected := shared.Map(tc.expected, strings.Fields)

			out1 := split(tc.given)
			out2 := plain2(out1)
			for i, act := range out2 {
				if i >= len(expected) {
					break
				}
				exp := expected[i]
				testutils.AssertSlices(t, exp, act)
			}
			testutils.AssertInts(t, len(expected), len(out2))

			// frames, groups := scan(tc.given)
			// expressions := flat(frames, groups)

			// for i, act := range expressions {
			// 	if i >= len(expected) {
			// 		break
			// 	}
			// 	exp := expected[i]
			// 	testutils.AssertSlices(t, exp, act)
			// }
			// testutils.AssertInts(t, len(expected), len(expressions))
		})
	}
}

func BenchmarkScanFlat(b *testing.B) {
	tests := []struct {
		name  string
		given string
	}{
		// split+plain:
		// 429.6 ns/op	     768 B/op	       8 allocs/op
		// 694.1 ns/op	    1168 B/op	      19 allocs/op
		// 2995 ns/op	    5032 B/op	      55 allocs/op
		// split+plain2:
		// 432.3 ns/op	     768 B/op	       8 allocs/op
		// 694.8 ns/op	    1136 B/op	      19 allocs/op
		// 2829 ns/op	    4632 B/op	      57 allocs/op
		{
			// 458.5 ns/op	     672 B/op	      10 allocs/op
			name: "No groups is used",
			given: `
			e2:4 b3:2|2
			f1|3 & c2|2 d2
			`,
		},
		{
			// 695.2 ns/op	    1632 B/op	      19 allocs/op
			name: "Riff with simple chord",
			given: `
			[e2 & b3] + [0 4 5]
			`,
		},
		{
			// 2316 ns/op	    4096 B/op	      38 allocs/op
			name: "All features",
			given: `
			e2|3 b3:2|2 & d2 # inline
			@song1(e2 2):8+T # sequence
			[a&b] + [2 4] # simple chord
			[e2 d2 c3]:3 # mod duration
			[a b] & [c d] # two groups
			`,
		},
	}

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				// frames, groups := scan(tc.given)
				// flat(frames, groups)
				out := split(tc.given)
				plain2(out)
			}
		})
	}
}
