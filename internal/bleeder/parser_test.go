package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
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
			>c4 |>e4 >60:2 <+8
			@chord:a2 | >a2 >g2
			`,
			expected: [][]string{
				{">c4", "|", ">e4", ">60:2", "<+8"},
				{"@chord:a2", "|", ">a2", ">g2"},
			},
		},
		{
			name: "Simple multiline Riff",
			given: `
			c4        e4 60 68
			@chord:a2 _  a2 g2
			`,
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

			fmt.Printf("RESULT of %s\n", tc.name)
			for i, v := range actual {
				fmt.Printf("R [%d] - %s\n", i, strings.Join(v, ", "))
			}

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
			values: nil,
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
		// 182.7 ns/op	     288 B/op	       3 allocs/op
		{s: "note:e2 dur:1", values: []string{"c#3", "2"}},
		// 259.5 ns/op	     304 B/op	       3 allocs/op
		{s: "n:e2 m:60 d:2", values: []string{"a3"}},
		// 470.7 ns/op	     304 B/op	       3 allocs/op
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
