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

var scanFlatCases = []struct {
	name, raw  string
	scan, flat []string
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

func TestScan2(t *testing.T) {
	for _, tc := range scanFlatCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			act := scan(tc.raw)
			exp := shared.Map(tc.scan, strings.Fields)
			testutils.AssertNested(t, exp, act)
		})
	}
}

func TestFlat2(t *testing.T) {
	for _, tc := range scanFlatCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.UseFlags(t)
			act := flat(scan(tc.raw))
			exp := shared.Map(tc.flat, strings.Fields)
			testutils.AssertNested(t, exp, act)
		})
	}
}

func BenchmarkScan2(b *testing.B) {
	// empty_input-8         	               73.65 ns/op	     192 B/op	       1 allocs/op
	// only_comment-8        	               76.35 ns/op	     192 B/op	       1 allocs/op
	// single_token-8        	               74.94 ns/op	     208 B/op	       2 allocs/op
	// values_on_one_line-8  	               129.0 ns/op	     240 B/op	       4 allocs/op
	// operators_glue-8      	               218.5 ns/op	     360 B/op	       5 allocs/op
	// all_operators-8       	               261.0 ns/op	     688 B/op	       4 allocs/op
	// inline_comment-8      	               106.5 ns/op	     224 B/op	       3 allocs/op
	// comment_lines_between-8        	       117.9 ns/op	     224 B/op	       3 allocs/op
	// blank_lines-8                  	       128.1 ns/op	     224 B/op	       3 allocs/op
	// tabs_and_spaces-8              	       237.4 ns/op	     272 B/op	       4 allocs/op
	// link_positional_args-8         	       251.1 ns/op	     312 B/op	       3 allocs/op
	// link_named_args-8              	       232.4 ns/op	     352 B/op	       3 allocs/op
	// patch-8                        	       239.5 ns/op	     312 B/op	       3 allocs/op
	// link_with_modification-8       	       254.4 ns/op	     640 B/op	       4 allocs/op
	// two_links_in_a_row-8           	       215.0 ns/op	     376 B/op	       4 allocs/op
	// empty_args-8                   	       142.3 ns/op	     272 B/op	       3 allocs/op
	// simple_group-8                 	       173.7 ns/op	     272 B/op	       6 allocs/op
	// group_with_spaces_inside-8     	       183.9 ns/op	     272 B/op	       6 allocs/op
	// group_single_item-8            	       147.1 ns/op	     248 B/op	       5 allocs/op
	// empty_group-8                  	       137.3 ns/op	     232 B/op	       4 allocs/op
	// group_with_suffix-8            	       251.9 ns/op	     320 B/op	       7 allocs/op
	// group_with_prefix-8            	       210.9 ns/op	     304 B/op	       6 allocs/op
	// group_both_sides-8             	       243.0 ns/op	     344 B/op	       6 allocs/op
	// item_with_operator-8           	       202.2 ns/op	     304 B/op	       6 allocs/op
	// group_then_value-8             	       220.1 ns/op	     288 B/op	       7 allocs/op
	// value_then_group-8             	       219.0 ns/op	     288 B/op	       7 allocs/op
	// repeat_modifier-8              	       245.7 ns/op	     344 B/op	       6 allocs/op
	// two_groups_joined-8            	       321.3 ns/op	     360 B/op	       9 allocs/op
	// two_groups_with_tails-8        	       405.8 ns/op	     496 B/op	      10 allocs/op
	// three_groups-8                 	       537.0 ns/op	     896 B/op	      14 allocs/op
	// modification_on_both-8          	  	   419.3 ns/op	     528 B/op	      10 allocs/op
	// adjacent_groups_no_operator-8   	  	   319.1 ns/op	     344 B/op	      10 allocs/op
	// adjacent_groups_with_space-8    	  	   313.8 ns/op	     344 B/op	      10 allocs/op
	// group_of_links-8                	  	   319.5 ns/op	     464 B/op	       7 allocs/op
	// group_of_links_with_suffix-8    	  	   608.2 ns/op	     752 B/op	      10 allocs/op
	// parallel_outside_groups-8       	  	   120.8 ns/op	     256 B/op	       3 allocs/op
	// parallel_with_modifications-8   	  	   194.6 ns/op	     344 B/op	       4 allocs/op
	// chord_group-8                   	  	   207.9 ns/op	     288 B/op	       7 allocs/op
	// chord_group_with_suffix-8       	  	   242.8 ns/op	     328 B/op	       7 allocs/op
	// chord_of_three-8                	  	   312.1 ns/op	     368 B/op	       9 allocs/op
	// chord_times_group-8             	  	   331.8 ns/op	     384 B/op	      10 allocs/op
	// note_then_group-8               	  	   210.4 ns/op	     304 B/op	       6 allocs/op
	// group_then_note-8               	  	   205.9 ns/op	     304 B/op	       6 allocs/op
	// groups_in_parallel-8            	  	   312.5 ns/op	     360 B/op	       9 allocs/op
	// chords_in_both_groups-8         	  	   620.2 ns/op	     784 B/op	      12 allocs/op
	// chord_in_second_group-8         	  	   407.8 ns/op	     384 B/op	      10 allocs/op
	// chord_in_third_group-8          	  	   551.9 ns/op	     912 B/op	      15 allocs/op
	// multi-line_mixed-8              	  	   594.9 ns/op	     912 B/op	      15 allocs/op
	// multi-line_groups_and_comments-8   	   430.8 ns/op	     448 B/op	      11 allocs/op
	for _, tc := range scanFlatCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				scan(tc.raw)
			}
		})
	}
}

func BenchmarkFlat2(b *testing.B) {
	// empty_input-8         			        	71.35 ns/op	     192 B/op	       1 allocs/op
	// only_comment-8        			        	68.80 ns/op	     192 B/op	       1 allocs/op
	// single_token-8        			        	54.12 ns/op	     192 B/op	       1 allocs/op
	// values_on_one_line-8  			        	56.28 ns/op	     192 B/op	       1 allocs/op
	// operators_glue-8      			        	56.36 ns/op	     192 B/op	       1 allocs/op
	// all_operators-8       			        	48.66 ns/op	     192 B/op	       1 allocs/op
	// inline_comment-8      			        	93.14 ns/op	     192 B/op	       1 allocs/op
	// comment_lines_between-8         		        66.97 ns/op	     192 B/op	       1 allocs/op
	// blank_lines-8                   		        66.89 ns/op	     192 B/op	       1 allocs/op
	// tabs_and_spaces-8               		        56.78 ns/op	     192 B/op	       1 allocs/op
	// link_positional_args-8          		        53.40 ns/op	     192 B/op	       1 allocs/op
	// link_named_args-8               		        53.47 ns/op	     192 B/op	       1 allocs/op
	// patch-8                         		        52.87 ns/op	     192 B/op	       1 allocs/op
	// link_with_modification-8        		        53.41 ns/op	     192 B/op	       1 allocs/op
	// two_links_in_a_row-8            		        57.81 ns/op	     192 B/op	       1 allocs/op
	// empty_args-8                    		        60.99 ns/op	     192 B/op	       1 allocs/op
	// simple_group-8                  		        146.8 ns/op	     224 B/op	       3 allocs/op
	// group_with_spaces_inside-8      		        128.9 ns/op	     224 B/op	       3 allocs/op
	// group_single_item-8             		        88.73 ns/op	     208 B/op	       2 allocs/op
	// empty_group-8                   		        60.93 ns/op	     192 B/op	       1 allocs/op
	// group_with_suffix-8             		        251.7 ns/op	     384 B/op	       7 allocs/op
	// group_with_prefix-8             		        138.7 ns/op	     288 B/op	       3 allocs/op
	// group_both_sides-8              		        198.0 ns/op	     480 B/op	       5 allocs/op
	// item_with_operator-8            		        133.6 ns/op	     256 B/op	       3 allocs/op
	// group_then_value-8              		        135.3 ns/op	     224 B/op	       3 allocs/op
	// value_then_group-8              		        133.7 ns/op	     224 B/op	       3 allocs/op
	// repeat_modifier-8               		        149.4 ns/op	     352 B/op	       3 allocs/op
	// two_groups_joined-8             		        307.5 ns/op	     416 B/op	       7 allocs/op
	// two_groups_with_tails-8         		        486.8 ns/op	    1184 B/op	      11 allocs/op
	// three_groups-8                  		        747.3 ns/op	    1248 B/op	      16 allocs/op
	// modification_on_both-8          		        547.4 ns/op	    1632 B/op	      11 allocs/op
	// adjacent_groups_no_operator-8   		        217.1 ns/op	     256 B/op	       5 allocs/op
	// adjacent_groups_with_space-8    		        216.9 ns/op	     256 B/op	       5 allocs/op
	// group_of_links-8                		        149.4 ns/op	     336 B/op	       3 allocs/op
	// group_of_links_with_suffix-8    		        677.0 ns/op	    2528 B/op	      11 allocs/op
	// parallel_outside_groups-8       		        50.43 ns/op	     192 B/op	       1 allocs/op
	// parallel_with_modifications-8   		        56.74 ns/op	     192 B/op	       1 allocs/op
	// chord_group-8                   		        133.8 ns/op	     224 B/op	       3 allocs/op
	// chord_group_with_suffix-8       		        180.8 ns/op	     320 B/op	       5 allocs/op
	// chord_of_three-8                		        303.8 ns/op	     576 B/op	       8 allocs/op
	// chord_times_group-8             		        352.0 ns/op	     608 B/op	       8 allocs/op
	// note_then_group-8               		        143.2 ns/op	     288 B/op	       3 allocs/op
	// group_then_note-8               		        280.9 ns/op	     320 B/op	       5 allocs/op
	// groups_in_parallel-8            		        315.2 ns/op	     416 B/op	       7 allocs/op
	// chords_in_both_groups-8         		        365.1 ns/op	     608 B/op	       8 allocs/op
	// chord_in_second_group-8         		        356.4 ns/op	     608 B/op	       8 allocs/op
	// chord_in_third_group-8          		        878.6 ns/op	    2016 B/op	      18 allocs/op
	// multi-line_mixed-8              		        259.4 ns/op	     384 B/op	       5 allocs/op
	// multi-line_groups_and_comments-8 	        271.5 ns/op	     352 B/op	       7 allocs/op
	for _, tc := range scanFlatCases {
		frames := scan(tc.raw)
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(frames)
			}
		})
	}
}

func BenchmarkScanFlat2(b *testing.B) {
	// empty_input-8         	               144.4 ns/op	     384 B/op	       2 allocs/op
	// only_comment-8        	               133.1 ns/op	     384 B/op	       2 allocs/op
	// single_token-8        	               114.7 ns/op	     400 B/op	       3 allocs/op
	// values_on_one_line-8  	               181.9 ns/op	     432 B/op	       5 allocs/op
	// operators_glue-8      	               272.6 ns/op	     552 B/op	       6 allocs/op
	// all_operators-8       	               311.2 ns/op	     880 B/op	       5 allocs/op
	// inline_comment-8      	               155.4 ns/op	     416 B/op	       4 allocs/op
	// comment_lines_between-8        	       173.3 ns/op	     416 B/op	       4 allocs/op
	// blank_lines-8                  	       190.4 ns/op	     416 B/op	       4 allocs/op
	// tabs_and_spaces-8              	       217.7 ns/op	     464 B/op	       5 allocs/op
	// link_positional_args-8         	       206.1 ns/op	     504 B/op	       4 allocs/op
	// link_named_args-8              	       243.2 ns/op	     544 B/op	       4 allocs/op
	// patch-8                        	       205.1 ns/op	     504 B/op	       4 allocs/op
	// link_with_modification-8       	       551.4 ns/op	     832 B/op	       5 allocs/op
	// two_links_in_a_row-8           	       314.3 ns/op	     568 B/op	       5 allocs/op
	// empty_args-8                   	       239.6 ns/op	     464 B/op	       4 allocs/op
	// simple_group-8                 	       303.5 ns/op	     496 B/op	       9 allocs/op
	// group_with_spaces_inside-8     	       315.6 ns/op	     496 B/op	       9 allocs/op
	// group_single_item-8            	       237.8 ns/op	     456 B/op	       7 allocs/op
	// empty_group-8                  	       172.8 ns/op	     424 B/op	       5 allocs/op
	// group_with_suffix-8            	       492.8 ns/op	     704 B/op	      14 allocs/op
	// group_with_prefix-8            	        1384 ns/op	     592 B/op	       9 allocs/op
	// group_both_sides-8             	       478.4 ns/op	     824 B/op	      11 allocs/op
	// item_with_operator-8           	       368.3 ns/op	     560 B/op	       9 allocs/op
	// group_then_value-8             	       370.7 ns/op	     512 B/op	      10 allocs/op
	// value_then_group-8             	       351.7 ns/op	     512 B/op	      10 allocs/op
	// repeat_modifier-8              	       384.3 ns/op	     696 B/op	       9 allocs/op
	// two_groups_joined-8            	       619.9 ns/op	     776 B/op	      16 allocs/op
	// two_groups_with_tails-8        	       887.8 ns/op	    1680 B/op	      21 allocs/op
	// three_groups-8                 	        1298 ns/op	    2144 B/op	      30 allocs/op
	// modification_on_both-8         	       957.2 ns/op	    2160 B/op	      21 allocs/op
	// adjacent_groups_no_operator-8  	       528.5 ns/op	     600 B/op	      15 allocs/op
	// adjacent_groups_with_space-8   	       535.1 ns/op	     600 B/op	      15 allocs/op
	// group_of_links-8               	       456.8 ns/op	     800 B/op	      10 allocs/op
	// group_of_links_with_suffix-8   	        1297 ns/op	    3280 B/op	      21 allocs/op
	// parallel_outside_groups-8      	       163.2 ns/op	     448 B/op	       4 allocs/op
	// parallel_with_modifications-8  	       242.7 ns/op	     536 B/op	       5 allocs/op
	// chord_group-8                  	       344.7 ns/op	     512 B/op	      10 allocs/op
	// chord_group_with_suffix-8      	       424.3 ns/op	     648 B/op	      12 allocs/op
	// chord_of_three-8               	       625.5 ns/op	     944 B/op	      17 allocs/op
	// chord_times_group-8            	       695.6 ns/op	     992 B/op	      18 allocs/op
	// note_then_group-8              	       351.4 ns/op	     592 B/op	       9 allocs/op
	// group_then_note-8              	       385.1 ns/op	     624 B/op	      11 allocs/op
	// groups_in_parallel-8           	       619.0 ns/op	     776 B/op	      16 allocs/op
	// chords_in_both_groups-8        	       805.3 ns/op	    1392 B/op	      20 allocs/op
	// chord_in_second_group-8        	       699.5 ns/op	     992 B/op	      18 allocs/op
	// chord_in_third_group-8         	        1471 ns/op	    2928 B/op	      33 allocs/op
	// multi-line_mixed-8             	       834.0 ns/op	    1296 B/op	      20 allocs/op
	// multi-line_groups_and_comments-8        679.3 ns/op	     800 B/op	      18 allocs/op
	for _, tc := range scanFlatCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				flat(scan(tc.raw))
			}
		})
	}
}
