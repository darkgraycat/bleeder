package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/shared/testutils"
	"fmt"
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
			>c4|>e4 >60:2 <+8
			@chord:a2 | >a2 >g2`,
			expected: [][]string{
				{">c4", "|", ">e4", ">60:2", "<+8"},
				{"@chord:a2", "|", ">a2", ">g2"},
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
		// 3634864	       296.3 ns/op	     368 B/op	       5 allocs/op
		`
			>c4 |>e4 >60:2 <+8
			@chord:a2 | >a2 >g2
		`,
		// 5028037	       238.1 ns/op	     352 B/op	       5 allocs/op
		`
			c4        e4 60 68
			@chord:a2 _  a2 g2
		`,
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
			values: []string{"c#3", "2"},
			expected: map[string]float64{
				"note": float64(audio.NoteToMidi("c#3")),
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
			name:   "values with modificators to defaults",
			given:  "a:20",
			values: []string{"+7"},
			expected: map[string]float64{
				"a": 27,
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
		s      string
		values []string
	}{
		// 172.8 ns/op	     288 B/op	       3 allocs/op
		{s: "note:e2 dur:1", values: []string{"c#3", "2"}},
		// 242.6 ns/op	     304 B/op	       3 allocs/op
		{s: "n:e2 m:60 d:2", values: []string{"a3"}},
		// 442.0 ns/op	     304 B/op	       3 allocs/op
		{s: "a:20 b:a+10 c:b+20", values: []string{"40"}},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				parseVars(tc.s, tc.values)
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
			// 797.8 ns/op	    1010 B/op	      13 allocs/op
			vars: map[string]float64{"note": 60, "dur": 8},
			given: `
			>note dur |
			>note+7 dur/2`,
		},
		{
			// 1006 ns/op	    6808 B/op	      10 allocs/op
			vars: map[string]float64{"a": 60, "b": 80},
			given: `
			a _ _ a
			b b _ _
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

func TestEvalArg(t *testing.T) {
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
			given:    "c#2+7",
			expected: float64(audio.NoteToMidi("c#2")) + 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			actual := evalArg(tc.given)
			testutils.AssertFloats(t, tc.expected, actual)
		})
	}
}

func BenchmarkEvalArg(b *testing.B) {
	tests := []string{
		// 42.76 ns/op	       0 B/op	       0 allocs/op
		"65.0",
		// 53.73 ns/op	       0 B/op	       0 allocs/op
		"60+4",
		// 44.81 ns/op	       0 B/op	       0 allocs/op
		"c#2+7",
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				evalArg(tc)
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
			given:    []string{"+7", "-1"},
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
			// 6.490 ns/op	       0 B/op	       0 allocs/op
			given: []string{"60", "2", "0.8"},
			idx:   1,
			prev:  "1",
		},
		{
			// 2.046 ns/op	       0 B/op	       0 allocs/op
			given: []string{"60"},
			idx:   2,
			prev:  "80",
		},
		{
			// 23.73 ns/op	       4 B/op	       1 allocs/op
			given: []string{"+7", "-1"},
			idx:   0,
			prev:  "60",
		},
		{
			// 10.10 ns/op	       0 B/op	       0 allocs/op
			given: []string{"c4"},
			idx:   0,
			prev:  "60",
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
