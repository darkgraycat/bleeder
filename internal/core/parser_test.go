package core

import (
	"bleeder/internal/audio"
	"bleeder/internal/shared/testutils"
	"fmt"
	"math"
	"testing"
)

func TestTokenizeContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name: "Simple multiline Lane",
			given: `
			>c4|>eb4 >60:2 <+8
			@chord:as2 | >a2 >gs2`,
			expected: [][]string{
				{">c4", "|", ">eb4", ">60:2", "<+8"},
				{"@chord:as2", "|", ">a2", ">gs2"},
			},
		},
		{
			name: "Commented multiline Lane",
			given: `
			#>c4|>e4 >60:2 <+8
			@chord:a2 | >a2 #>g2`,
			expected: [][]string{
				{"@chord:a2", "|", ">a2"},
			},
		},
		{
			name: "Simple multiline Riff",
			given: `
			c4        e4 60 68
			@chord:a2 _  a2 g2`,
			expected: [][]string{
				{"c4", "e4", "60", "68"},
				{"@chord:a2", "_", "a2", "g2"},
			},
		},
		{
			name: "Commented multiline Riff",
			given: `
			c4        e4 60 68
			# @chord:a2 _  a2 g2`,
			expected: [][]string{
				{"c4", "e4", "60", "68"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := tokenizeContent(tc.given)

			testutils.AssertInts(t, len(tc.expected), len(actual))
			for i, act := range actual {
				exp := tc.expected[i]
				testutils.AssertSlices(t, exp, act)
			}
		})
	}
}

func BenchmarkTokenizeContent(b *testing.B) {
	tests := []string{
		// 273.4 ns/op	     368 B/op	       5 allocs/op
		`
			>c4|>eb4 >60:2 <+8
			@chord:as2 | >a2 >gs2`,
		// 222.3 ns/op	     272 B/op	       4 allocs/op
		`
			#>c4|>e4 >60:2 <+8
			@chord:a2 | >a2 #>g2`,

		// 236.2 ns/op	     320 B/op	       5 allocs/op
		`
			c4        e4 60 68
			@chord:a2 _  a2 g2`,
		// 206.4 ns/op	     288 B/op	       4 allocs/op
		`
			c4        e4 60 68
			# @chord:a2 _  a2 g2`,
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				tokenizeContent(tc)
			}
		})
	}
}

func TestParseVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		values   []string
		expected map[string]float64
	}{
		{
			name:   "parse simple vars",
			given:  "note:e2 dur:1",
			values: []string{"cs3", "2"},
			expected: map[string]float64{
				"note": float64(audio.NoteToMidi("cs3")),
				"dur":  2,
			},
		},
		{
			name:   "parse simple vars with default fallback",
			given:  "n:e2 m:60 d:2",
			values: []string{"a3"},
			expected: map[string]float64{
				"n": float64(audio.NoteToMidi("a3")),
				"m": 60,
				"d": 2,
			},
		},
		{
			name:   "parse dependent vars",
			given:  "a:20 b:a+10 c:b+20",
			values: []string{"40"},
			expected: map[string]float64{
				"a": 40,
				"b": 50,
				"c": 70,
			},
		},
		{
			name:   "parse dependent vars with default fallback",
			given:  "a:20 b:a+10 c:b+20",
			values: []string{},
			expected: map[string]float64{
				"a": 20,
				"b": 30,
				"c": 50,
			},
		},
		{
			name:   "parse dependent vars with overrides",
			given:  "a:20 b:a+10",
			values: []string{"40", "80"},
			expected: map[string]float64{
				"a": 40,
				"b": 80,
			},
		},
		{
			name:   "values with modificators",
			given:  "a:20",
			values: []string{"40+7"},
			expected: map[string]float64{
				"a": 47,
			},
		},
		{
			name:   "parse complex vars",
			given:  "a:20 b:a+10 c:b+20 d:44 e:12",
			values: []string{"40+5", "-2", "", "80"},
			expected: map[string]float64{
				"a": 45, // absolute 40+5
				"b": 53, // a+10-2
				"c": 73, // b+20
				"d": 80, // absolute 80
				"e": 12, // default
			},
		},
		{
			name:   "parse both sides",
			given:  "a:5 b:a+10 c:10-a",
			values: []string{"3"},
			expected: map[string]float64{
				"a": 3,
				"b": 13,
				"c": 7,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := parseVars(tc.given, tc.values)
			testutils.AssertMaps(t, tc.expected, actual)
		})
	}
}

func BenchmarkParseVars(b *testing.B) {
	tests := []struct {
		given  string
		values []string
	}{
		{
			// 211.2 ns/op	     288 B/op	       3 allocs/op
			given:  "note:e2 dur:1",
			values: []string{"cs3", "2"},
		},
		{
			// 258.9 ns/op	     304 B/op	       3 allocs/op
			given:  "n:e2 m:60 d:2",
			values: []string{"a3"},
		},
		{
			// 488.3 ns/op	     304 B/op	       3 allocs/op
			given:  "a:20 b:a+10 c:b+20",
			values: []string{"40"},
		},
		{
			// 479.8 ns/op	     304 B/op	       3 allocs/op
			given:  "a:20 b:a+10 c:b+20",
			values: []string{},
		},
		{
			// 221.6 ns/op	     288 B/op	       3 allocs/op
			given:  "a:20 b:a+10",
			values: []string{"40", "80"},
		},
		{
			// 177.5 ns/op	     272 B/op	       3 allocs/op
			given:  "a:20",
			values: []string{"40+7"},
		},
		{
			// 695.4 ns/op	     344 B/op	       4 allocs/op
			given:  "a:20 b:a+10 c:b+20 d:44 e:12",
			values: []string{"40+5", "+1", "", "80"},
		},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				parseVars(tc.given, tc.values)
			}
		})
	}
}

func TestApplyVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		vars     map[string]float64
		expected string
	}{
		{
			name: "apply vars on multiline Lane content",
			vars: map[string]float64{"note": 60, "dur": 8},
			given: `
			>note dur |
			>note+7 dur/2`,
			expected: `
			>60 8 |
			>60+7 8/2`,
		},
		{
			name: "apply vars on multiline Riff content",
			vars: map[string]float64{"a": 60, "b": 80},
			given: `
			a _ _ a
			b b _ _
			`,
			expected: `
			60 _ _ 60
			80 80 _ _
			`,
		},
		{
			name: "apply vars avoiding sequence names",
			vars: map[string]float64{"a": 60, "b": 80, "d1": 2, "d": 3},
			given: `
			@bass |
			@drum |
			>a:d1 >a:d >b:d >a:d1
			`,
			expected: `
			@bass |
			@drum |
			>60:2 >60:3 >80:3 >60:2
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := applyVars(tc.given, tc.vars)
			testutils.AssertStrings(t, tc.expected, actual)
		})
	}
}

func BenchmarkApplyVars(b *testing.B) {
	tests := []struct {
		vars  map[string]float64
		given string
	}{
		{
			// 449.8 ns/op	     202 B/op	       6 allocs/op
			vars: map[string]float64{"note": 60, "dur": 8},
			given: `
			>note dur |
			>note+7 dur/2`,
		},
		{
			// 437.5 ns/op	     204 B/op	       7 allocs/op
			vars: map[string]float64{"a": 60, "b": 80},
			given: `
			a _ _ a
			b b _ _
			`,
		},
		{
			// 1094 ns/op	     300 B/op	       7 allocs/op
			vars: map[string]float64{"a": 60, "b": 80, "d1": 2, "d": 3},
			given: `
			@bass |
			@drum |
			>a:d1 >a:d >b:d >a:d1
			`,
		},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				applyVars(tc.given, tc.vars)
			}
		})
	}
}

func TestEvalVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected float64
	}{
		{
			name:     "evaluate plain numeric value",
			given:    "65.0",
			expected: 65.0,
		},
		{
			name:     "evaluate sum for midi",
			given:    "60+4",
			expected: 64,
		},
		{
			name:     "evaluate sum for note",
			given:    "cs2+7",
			expected: float64(audio.NoteToMidi("cs2")) + 7,
		},
		{
			name:     "evaluate complex in order",
			given:    "6*2-2/5*3+10/4",
			expected: 4,
		},
		{
			name:     "evaluate into NaN",
			given:    "lol+9",
			expected: math.NaN(),
		},
		{
			name:     "evaluate with preceding modifier",
			given:    "-2-4",
			expected: -6,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := evalVars(tc.given, nil)
			if math.IsNaN(tc.expected) && math.IsNaN(actual) {
				return
			}
			testutils.AssertFloats(t, tc.expected, actual)
		})
	}
}

func BenchmarkEvalVars(b *testing.B) {
	tests := []string{
		// 42.52 ns/op	       0 B/op	       0 allocs/op
		"65.0",
		// 61.83 ns/op	       0 B/op	       0 allocs/op
		"60+4",
		// 54.37 ns/op	       0 B/op	       0 allocs/op
		"cs2+7",
		// 54.24 ns/op	       0 B/op	       0 allocs/op
		"cs2+7",
		// 187.2 ns/op	       0 B/op	       0 allocs/op
		"6*2-2/5*3+10/4",
		// 91.28 ns/op	      51 B/op	       2 allocs/op
		"cf2+9",
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				evalVars(tc, nil)
			}
		})
	}
}

func TestGetArg(t *testing.T) {
	tests := []struct {
		name     string
		given    []string
		idx      int
		prev     string
		expected string
	}{
		{
			name:     "get existing arg",
			given:    []string{"60", "2", "0.8"},
			idx:      1,
			prev:     "1",
			expected: "2",
		},
		{
			name:     "out of bounds returns prev",
			given:    []string{"60"},
			idx:      2,
			prev:     "80",
			expected: "80",
		},
		{
			name:     "modifier prepends prev",
			given:    []string{"+7"},
			idx:      0,
			prev:     "60",
			expected: "60+7",
		},
		{
			name:     "absolute returns as-is",
			given:    []string{"c4"},
			idx:      0,
			prev:     "60",
			expected: "c4",
		},
		{
			name:     "fallback to prev on empty",
			given:    []string{"60", ""},
			idx:      1,
			prev:     "12",
			expected: "12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := getArg(tc.given, tc.idx, tc.prev)
			testutils.AssertStrings(t, tc.expected, actual)
		})
	}
}

func BenchmarkGetArg(b *testing.B) {
	tests := []struct {
		given []string
		idx   int
		prev  string
	}{
		{
			// 2.129 ns/op	       0 B/op	       0 allocs/op
			given: []string{"60", "2", "0.8"},
			idx:   1,
			prev:  "1",
		},
		{
			// 2.062 ns/op	       0 B/op	       0 allocs/op
			given: []string{"60"},
			idx:   2,
			prev:  "80",
		},
		{
			// 20.78 ns/op	       4 B/op	       1 allocs/op
			given: []string{"+7", "-1"},
			idx:   0,
			prev:  "60",
		},
		{
			// 2.061 ns/op	       0 B/op	       0 allocs/op
			given: []string{"c4"},
			idx:   0,
			prev:  "60",
		},
		{
			// 2.072 ns/op	       0 B/op	       0 allocs/op
			given: []string{"60", ""},
			idx:   1,
			prev:  "12",
		},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				getArg(tc.given, tc.idx, tc.prev)
			}
		})
	}
}
