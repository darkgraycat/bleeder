package core

import (
	"bleeder/internal/shared"
	"bleeder/internal/shared/testutils"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestExperiments(t *testing.T) {
	tests := []string{
		"[a b c]|2",
		// "[a b] & [c d]",
		// "[a b] & [c & d]",
		// "[a & b] & [c d]",
		// "[a & b] & [c & d]",
		// "[a + b] & [c & d]",
		// "a & [b c]",       // dont: a & b c. how to make it a&b a&c?
		// "[a b] + 2 & [c d] +3",
		// "a&b + [2 4]", // 4/5 works
		// "[a & b] + [2 4]", // works

		// "2 + [a b] * [c d e] + 3",
		// "e2 & [a b c] + [7 12]",
		// current: a&b+2 a&b+4
		// if i want it: a+2 & b+2 a+4 & b+4

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
			out1 := scan(tc)
			out2 := flat(out1)
			fmt.Printf("INPUT: %v\n", tc)
			fmt.Printf("SPLIT: %v\n", out1)
			fmt.Printf("PLAIN: %v\n", out2)
			for i, expr := range out2 {
				fmt.Printf("[%d] ", i)
				for _, tok := range expr {
					fmt.Printf("%s ", tok)
				}
				fmt.Println()
			}

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

var scanTestCases = []struct {
	name     string
	given    string
	expected []string
}{
	{ // 169.5 ns/op	     256 B/op	       5 allocs/op
		name:     "Simple sequence",
		given:    "a2 b2 c3 d4",
		expected: []string{"a2", "b2", "c3", "d4"},
	},
	{ // 244.1 ns/op	     392 B/op	       5 allocs/op
		name:     "Simple sequence with modifications",
		given:    "a2+A b2+B:T 80/d",
		expected: []string{"a2 + A", "b2 + B : T", "80 / d"},
	},
	{ // 271.5 ns/op	     480 B/op	       5 allocs/op
		name:     "Simple sequence with repeated and parallel notes",
		given:    "a2 & b2|2 c3|3 & d4 & e4",
		expected: []string{"a2 & b2 | 2", "c3 | 3 & d4 & e4"},
	},
	{ // 541.2 ns/op	     800 B/op	       7 allocs/op
		name:  "Sequence with links",
		given: "@chord(a2) @chord{n a2 vol .5} @song(a3) @song{n a3 vol .5}",
		expected: []string{
			"@ chord ( a2 )", "@ chord { n a2 vol .5 }",
			"@ song ( a3 )", "@ song { n a3 vol .5 }",
		},
	},
	{ // 487.4 ns/op	     960 B/op	       6 allocs/op
		name:  "Sequence with parallel links/patches and modifications",
		given: "$bass(.4 .6) @chord(a2):8 & @bassRiff(a2 5)-12",
		expected: []string{
			"$ bass ( .4 .6 )",
			"@ chord ( a2 ) : 8 & @ bassRiff ( a2 5 ) - 12",
		},
	},
	{ // 309.1 ns/op	     360 B/op	       7 allocs/op
		name:     "Grouped notes with modifications",
		given:    "[a2 b2 c3]+2:8",
		expected: []string{"[", "a2", "b2", "c3", "] + 2 : 8"},
	},
	{ // 709.1 ns/op	     400 B/op	       9 allocs/op
		name:     "Grouped notes repeated with modifications",
		given:    "[a2 b2]:2 + [0 2]",
		expected: []string{"[", "a2", "b2", "] : 2 + [", "0", "2", "]"},
	},
	{ // 377.3 ns/op	     368 B/op	       9 allocs/op
		name:     "Grouped notes played in parallel",
		given:    "[a2 & b2 & c3]:3",
		expected: []string{"[", "a2", "&", "b2", "&", "c3", "] : 3"},
	},
	{ // 471.2 ns/op	     400 B/op	       9 allocs/op
		name:     "Grouped notes intersected with another group",
		given:    "[a2 b2] & [c2 d2]+2",
		expected: []string{"[", "a2", "b2", "] & [", "c2", "d2", "] + 2"},
	},
	{ // 548.8 ns/op	     624 B/op	      10 allocs/op
		name:  "Grouped links repeated with modifications",
		given: "[@chord(a2) @chord(b2)]:4 + [0 2]",
		expected: []string{
			"[", "@ chord ( a2 )", "@ chord ( b2 )",
			"] : 4 + [", "0", "2", "]",
		},
	},
	{ // 495.8 ns/op	     912 B/op	      13 allocs/op
		name:  "Grouped notes mixed with ungrouped",
		given: "[a2 b2]:2 c2:4 12+[d2 e2]",
		expected: []string{
			"[", "a2", "b2", "] : 2", "c2 : 4",
			"12 + [", "d2", "e2", "]",
		},
	},
}

func TestScan(t *testing.T) {
	for _, tc := range scanTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			act := scan(tc.given)
			exp := shared.Map(tc.expected, strings.Fields)
			testutils.AssertNested(t, exp, act)
		})
	}
}

func BenchmarkScan(b *testing.B) {
	for _, tc := range scanTestCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				scan(tc.given)
			}
		})
	}
}

var flatTestCases = []struct {
	name     string
	given    string
	expected []string
}{
	{ // 53.01 ns/op	     192 B/op	       1 allocs/op
		// 58.54 ns/op	     192 B/op	       1 allocs/op
		// 58.94 ns/op	     192 B/op	       1 allocs/op (pre-alloc)
		name:     "No groups is used",
		given:    "a2 b2 c3+5",
		expected: []string{"a2", "b2", "c3 + 5"},
	},
	{ // 339.1 ns/op	     576 B/op	      11 allocs/op
		// 334.3 ns/op	     552 B/op	      10 allocs/op
		// 250.3 ns/op	     384 B/op	       7 allocs/op
		name:     "Group with modification",
		given:    "[a2 b2 c2]:8",
		expected: []string{"a2 : 8", "b2 : 8", "c2 : 8"},
	},
	{ // 948.5 ns/op	    1680 B/op	      25 allocs/op
		// 892.8 ns/op	    1560 B/op	      21 allocs/op
		// 751.7 ns/op	    1248 B/op	      16 allocs/op
		name:  "Intersection of multiple small groups",
		given: "[a2 b2] + [0 2] * [x y]",
		expected: []string{
			"a2 + 0 * x", "b2 + 0 * x",
			"a2 + 2 * x", "b2 + 2 * x",
			"a2 + 0 * y", "b2 + 0 * y",
			"a2 + 2 * y", "b2 + 2 * y",
		},
	},
	{ // 722.1 ns/op	    1896 B/op	      17 allocs/op
		// 644.5 ns/op	    1848 B/op	      15 allocs/op
		// 541.4 ns/op	    1632 B/op	      11 allocs/op
		name:  "Groups with modifications",
		given: "12+[a2 b2]:8 + [4 6]/c",
		expected: []string{
			"12 + a2 : 8 + 4 / c", "12 + b2 : 8 + 4 / c",
			"12 + a2 : 8 + 6 / c", "12 + b2 : 8 + 6 / c",
		},
	},
	{ // 621.5 ns/op	    1064 B/op	      17 allocs/op
		// 595.4 ns/op	    1016 B/op	      15 allocs/op
		// 431.8 ns/op	     800 B/op	      11 allocs/op
		name:  "Groups with parallel intersected notes",
		given: "[a2 b2] & [c2 d2]:2",
		expected: []string{
			"a2 & c2 : 2", "b2 & c2 : 2",
			"a2 & d2 : 2", "b2 & d2 : 2",
		},
	},
	{ // 867.2 ns/op	    1880 B/op	      22 allocs/op
		// 762.4 ns/op	    1784 B/op	      17 allocs/op
		// 713.0 ns/op	    1472 B/op	      12 allocs/op
		name:  "Groups to play multiple chords",
		given: "[a2 & b2] + [0 1 2]",
		expected: []string{
			"a2 + 0", "&", "b2 + 0",
			"a2 + 1", "&", "b2 + 1",
			"a2 + 2", "&", "b2 + 2",
		},
	},
	{ // 588.2 ns/op	    1016 B/op	      18 allocs/op
		// 493.8 ns/op	     920 B/op	      13 allocs/op
		// 471.4 ns/op	     608 B/op	       8 allocs/op
		name:  "Groups with all notes parallel",
		given: "[a2 & b2] & [c2 & d2]",
		expected: []string{
			"a2 & c2", "&", "b2 & c2", "&",
			"a2 & d2", "&", "b2 & d2",
		},
	},
	{ // 231.3 ns/op	     416 B/op	       8 allocs/op
		// 223.2 ns/op	     392 B/op	       7 allocs/op
		// 199.3 ns/op	     320 B/op	       5 allocs/op
		name:     "Group with one parallel note",
		given:    "[a2 b2] & c2",
		expected: []string{"a2 & c2", "b2 & c2"},
	},
	{ // 279.1 ns/op	     768 B/op	       8 allocs/op
		// 275.7 ns/op	     744 B/op	       7 allocs/op
		// 327.1 ns/op	     672 B/op	       5 allocs/op
		name:  "Group with parallel note and modifications",
		given: "a2:4 & [b2 c2]|2",
		expected: []string{
			"a2 : 4 & b2 | 2",
			"a2 : 4 & c2 | 2",
		},
	},
	{ // 799.4 ns/op	    2216 B/op	      21 allocs/op
		// 726.8 ns/op	    2136 B/op	      17 allocs/op
		// 596.1 ns/op	    1824 B/op	      12 allocs/op
		name:  "Groups with various modifications",
		given: "7+[a2 & b2]|4 + [0 1]*7:2",
		expected: []string{
			"7 + a2 | 4 + 0 * 7 : 2", "&", "7 + b2 | 4 + 0 * 7 : 2",
			"7 + a2 | 4 + 1 * 7 : 2", "&", "7 + b2 | 4 + 1 * 7 : 2",
		},
	},
}

func TestFlat(t *testing.T) {
	for _, tc := range flatTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			act := flat(scan(tc.given))
			exp := shared.Map(tc.expected, strings.Fields)
			testutils.AssertNested(t, exp, act)
		})
	}
}

func BenchmarkFlat(b *testing.B) {
	for _, tc := range flatTestCases {
		frames := scan(tc.given)
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(frames)
			}
		})
	}
}

func BenchmarkScanFlat(b *testing.B) {
	// Simple_sequence-8         	                                 5005995	       219.0 ns/op	     448 B/op	       6 allocs/op
	// Simple_sequence_with_modifications-8         	             3845840	       300.2 ns/op	     584 B/op	       6 allocs/op
	// Simple_sequence_with_repeated_and_parallel_notes-8         	 3772772	       315.4 ns/op	     672 B/op	       6 allocs/op
	// Sequence_with_links-8                                      	 2037150	       593.3 ns/op	     992 B/op	       8 allocs/op
	// Sequence_with_parallel_links/patches_and_modifications-8   	 2250682	       528.1 ns/op	    1152 B/op	       7 allocs/op
	// Grouped_notes_with_modifications-8                         	 1939105	       614.9 ns/op	    1032 B/op	      18 allocs/op
	// Grouped_notes_repeated_with_modifications-8                	 1484674	       806.6 ns/op	    1208 B/op	      22 allocs/op
	// Grouped_notes_played_in_parallel-8                         	 1641362	       746.4 ns/op	    1168 B/op	      23 allocs/op
	// Grouped_notes_intersected_with_another_group-8             	 1279154	       932.0 ns/op	    1464 B/op	      26 allocs/op
	// Grouped_links_repeated_with_modifications-8                	 1000000	      1008 ns/op	    1816 B/op	      23 allocs/op
	// Grouped_notes_mixed_with_ungrouped-8                       	 1368540	       878.7 ns/op	    1520 B/op	      26 allocs/op
	// No_groups_is_used-8                                        	 5472892	       219.5 ns/op	     480 B/op	       6 allocs/op
	// Group_with_modification-8                                  	 2087001	       575.2 ns/op	     904 B/op	      18 allocs/op
	// Intersection_of_multiple_small_groups-8                    	  806496	      1483 ns/op	    2576 B/op	      39 allocs/op
	// Groups_with_modifications-8                                	 1000000	      1080 ns/op	    2424 B/op	      27 allocs/op
	// Groups_with_parallel_intersected_notes-8                   	 1311914	       914.3 ns/op	    1464 B/op	      26 allocs/op
	// Groups_to_play_multiple_chords-8                           	  968802	      1241 ns/op	    2664 B/op	      34 allocs/op
	// Groups_with_all_notes_parallel-8                           	 1000000	      1043 ns/op	    1864 B/op	      31 allocs/op
	// Group_with_one_parallel_note-8                             	 2712619	       443.3 ns/op	     728 B/op	      14 allocs/op
	// Group_with_parallel_note_and_modifications-8               	 2244489	       535.2 ns/op	    1152 B/op	      14 allocs/op
	// Groups_with_various_modifications-8                        	  925500	      1309 ns/op	    2792 B/op	      32 allocs/op
	tests := slices.Concat(scanTestCases, flatTestCases)
	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(scan(tc.given))
			}
		})
	}
}

var cases = []struct {
	name, given  string
	split, plain []string
}{
	// --- no groups
	{"empty input", ``,
		nil,
		nil,
	},
	{"only comment", `# nothing here`,
		nil,
		nil,
	},
	{"single token", `a2`,
		[]string{"a2"},
		[]string{"a2"},
	},
	{"values on one line", `a2 b2 c3`,
		[]string{"a2", "b2", "c3"},
		[]string{"a2", "b2", "c3"},
	},
	{"operators glue", `a2+1 b2:4 c3|2`,
		[]string{"a2 + 1", "b2 : 4", "c3 | 2"},
		[]string{"a2 + 1", "b2 : 4", "c3 | 2"},
	},
	{"all operators", `a+1-2*3/4%5^6`,
		[]string{"a + 1 - 2 * 3 / 4 % 5 ^ 6"},
		[]string{"a + 1 - 2 * 3 / 4 % 5 ^ 6"},
	},
	{"inline comment", `a2 b2 # c3 d3`,
		[]string{"a2", "b2"},
		[]string{"a2", "b2"},
	},
	{"comment lines between", "a2\n# skip\nb2",
		[]string{"a2", "b2"},
		[]string{"a2", "b2"},
	},
	{"blank lines", "a2\n\n\nb2",
		[]string{"a2", "b2"},
		[]string{"a2", "b2"},
	},
	{"tabs and spaces", "a2 \t b2   +  1",
		[]string{"a2", "b2 + 1"},
		[]string{"a2", "b2 + 1"},
	},

	// --- links and patches (always one frame)
	{"link positional args", `@chord(a2 3)`,
		[]string{"@ chord ( a2 3 )"},
		[]string{"@ chord ( a2 3 )"},
	},
	{"link named args", `@chord{note g3 dur 7}`,
		[]string{"@ chord { note g3 dur 7 }"},
		[]string{"@ chord { note g3 dur 7 }"},
	},
	{"patch", `$bass(0.77 12)`,
		[]string{"$ bass ( 0.77 12 )"},
		[]string{"$ bass ( 0.77 12 )"},
	},
	{"link with modification", `@song(e2 2):8+T`,
		[]string{"@ song ( e2 2 ) : 8 + T"},
		[]string{"@ song ( e2 2 ) : 8 + T"},
	},
	{"two links in a row", `@a(1) @b(2)`,
		[]string{"@ a ( 1 )", "@ b ( 2 )"},
		[]string{"@ a ( 1 )", "@ b ( 2 )"},
	},
	{"empty args", `@mysong()`,
		[]string{"@ mysong ( )"},
		[]string{"@ mysong ( )"},
	},

	// --- single group
	{"simple group", `[a b]`,
		[]string{"[", "a", "b", "]"},
		[]string{"a", "b"},
	},
	{"group with spaces inside", `[ a   b ]`,
		[]string{"[", "a", "b", "]"},
		[]string{"a", "b"},
	},
	{"group single item", `[a]`,
		[]string{"[", "a", "]"},
		[]string{"a"},
	},
	{"empty group", `[]`,
		[]string{"[", "]"},
		nil,
	},
	{"group with suffix", `[a b c]:3`,
		[]string{"[", "a", "b", "c", "] : 3"},
		[]string{"a : 3", "b : 3", "c : 3"},
	},
	{"group with prefix", `2 + [a b]`,
		[]string{"2 + [", "a", "b", "]"},
		[]string{"2 + a", "2 + b"},
	},
	{"group both sides", `2 + [a b] * 3`,
		[]string{"2 + [", "a", "b", "] * 3"},
		[]string{"2 + a * 3", "2 + b * 3"},
	},
	{"item with operator", `[a b+3]`,
		[]string{"[", "a", "b + 3", "]"},
		[]string{"a", "b + 3"},
	},
	{"group then value", `[a b] c`,
		[]string{"[", "a", "b", "]", "c"},
		[]string{"a", "b", "c"},
	},
	{"value then group", `c [a b]`,
		[]string{"c", "[", "a", "b", "]"},
		[]string{"c", "a", "b"},
	},
	{"repeat modifier", `a|2 + [x y]`,
		[]string{"a | 2 + [", "x", "y", "]"},
		[]string{"a | 2 + x", "a | 2 + y"},
	},

	// --- several groups (leftmost varies fastest)
	{"two groups joined", `[a b] + [c d]`,
		[]string{"[", "a", "b", "] + [", "c", "d", "]"},
		[]string{"a + c", "b + c", "a + d", "b + d"},
	},
	{"two groups with tails", `[a b] + 2 * [c d] : 8`,
		[]string{"[", "a", "b", "] + 2 * [", "c", "d", "] : 8"},
		[]string{"a + 2 * c : 8", "b + 2 * c : 8", "a + 2 * d : 8", "b + 2 * d : 8"},
	},
	{"three groups", `[a b] + [0 1] * [x y]`,
		[]string{"[", "a", "b", "] + [", "0", "1", "] * [", "x", "y", "]"},
		[]string{"a + 0 * x", "b + 0 * x", "a + 1 * x", "b + 1 * x", "a + 0 * y", "b + 0 * y", "a + 1 * y", "b + 1 * y"},
	},
	{"modification on both", `12+[a2 b2]:8 + [4 6]/c`,
		[]string{"12 + [", "a2", "b2", "] : 8 + [", "4", "6", "] / c"},
		[]string{"12 + a2 : 8 + 4 / c", "12 + b2 : 8 + 4 / c", "12 + a2 : 8 + 6 / c", "12 + b2 : 8 + 6 / c"},
	},
	{"adjacent groups no operator", `[a b][c d]`,
		[]string{"[", "a", "b", "]", "[", "c", "d", "]"},
		[]string{"a", "b", "c", "d"},
	},
	{"adjacent groups with space", `[a b] [c d]`,
		[]string{"[", "a", "b", "]", "[", "c", "d", "]"},
		[]string{"a", "b", "c", "d"},
	},

	// --- links inside groups
	{"group of links", `[@chord(g3) @bass()]`,
		[]string{"[", "@ chord ( g3 )", "@ bass ( )", "]"},
		[]string{"@ chord ( g3 )", "@ bass ( )"},
	},
	{"group of links with suffix", `[@chord{note g3 dur 7} @mysong(g2 a4)] : 8 + [a b] * 3`,
		[]string{"[", "@ chord { note g3 dur 7 }", "@ mysong ( g2 a4 )", "] : 8 + [", "a", "b", "] * 3"},
		[]string{"@ chord { note g3 dur 7 } : 8 + a * 3", "@ mysong ( g2 a4 ) : 8 + a * 3",
			"@ chord { note g3 dur 7 } : 8 + b * 3", "@ mysong ( g2 a4 ) : 8 + b * 3"},
	},

	// --- parallel (&)
	{"parallel outside groups", `a2 & b2`,
		[]string{"a2 & b2"},
		[]string{"a2 & b2"},
	},
	{"parallel with modifications", `a2:4 & b2|2 c3`,
		[]string{"a2 : 4 & b2 | 2", "c3"},
		[]string{"a2 : 4 & b2 | 2", "c3"},
	},
	{"chord group", `[a & b]`,
		[]string{"[", "a", "&", "b", "]"},
		[]string{"a", "&", "b"},
	},
	{"chord group with suffix", `[a & b]:2`,
		[]string{"[", "a", "&", "b", "] : 2"},
		[]string{"a : 2", "&", "b : 2"},
	},
	{"chord of three", `[a & b & c] + 1`,
		[]string{"[", "a", "&", "b", "&", "c", "] + 1"},
		[]string{"a + 1", "&", "b + 1", "&", "c + 1"},
	},
	{"chord times group", `[a & b] + [2 4]`,
		[]string{"[", "a", "&", "b", "] + [", "2", "4", "]"},
		[]string{"a + 2", "&", "b + 2", "a + 4", "&", "b + 4"}},
	{"note then group", `a & [b c]`,
		[]string{"a & [", "b", "c", "]"},
		[]string{"a & b", "a & c"},
	},
	{"group then note", `[b c] & a`,
		[]string{"[", "b", "c", "] & a"},
		[]string{"b & a", "c & a"},
	},
	{"groups in parallel", `[a b] & [c d]`,
		[]string{"[", "a", "b", "] & [", "c", "d", "]"},
		[]string{"a & c", "b & c", "a & d", "b & d"},
	},
	{"chords in both groups", `[a & b] & [c & d]`,
		[]string{"[", "a", "&", "b", "] & [", "c", "&", "d", "]"},
		[]string{"a & c", "&", "b & c", "&", "a & d", "&", "b & d"},
	},
	// pin current behaviour: the lone & lands BETWEEN the blocks, not between pair members
	{"chord in second group", `[a b] & [c & d]`,
		[]string{"[", "a", "b", "] & [", "c", "&", "d", "]"},
		[]string{"a & c", "b & c", "&", "a & d", "b & d"},
	},
	{"chord in third group", `[a b] + [c d] * [e & f]`,
		[]string{"[", "a", "b", "] + [", "c", "d", "] * [", "e", "&", "f", "]"},
		[]string{"a + c * e", "b + c * e", "a + d * e", "b + d * e", "&", "a + c * f", "b + c * f", "a + d * f", "b + d * f"},
	},

	// --- multi-line
	{"multi-line mixed", "x y\n2 + [a b]\n3 + [c d]\nz",
		[]string{"x", "y", "2 + [", "a", "b", "]", "3 + [", "c", "d", "]", "z"},
		[]string{"x", "y", "2 + a", "2 + b", "3 + c", "3 + d", "z"},
	},
	{"multi-line groups and comments", "[a b] # first\n# note\n[c d]:2 # second",
		[]string{"[", "a", "b", "]", "[", "c", "d", "] : 2"}, []string{"a", "b", "c : 2", "d : 2"},
	},
}
