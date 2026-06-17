package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

// tests

func TestNormalizeLaneContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		message  string
		expected []string
	}{
		{
			name:  "normalize oneline string",
			given: ">60_ :e#3_2~440.0,2_3 |*2_",
			expected: []string{
				">60", "_",
				":e#3", "_2",
				"~440.0 2", "_3",
				"|*2", "_",
			},
		},
		{
			name: "normalize multiline string",
			given: `
				>40*2,3
				:f#4-2,2_
				~440 _4 |880,*2
			`,
			expected: []string{
				">40*2 3",
				":f#4-2 2",
				"_",
				"~440", "_4", "|880 *2",
			},
		},
		{
			name: "normalize empty content",
			given: `

			`,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizeLaneContent(tc.given)
			for i, v := range actual {
				actual[i] = strings.TrimSpace(v)
			}
			testutils.AssertSlices(t, tc.expected, actual)
			for i, exp := range tc.expected {
				testutils.AssertStrings(t, exp, actual[i])
			}
		})
	}
}

func TestNormalizeRiffContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		message  string
		expected []string
	}{
		{
			name: "normalize correct riff content",
			given: `
				--b>>
				-a>a>`,
			expected: []string{
				"--b>>",
				"-a>a>",
			},
		},
		{
			name: "normalize correct user formatted riff content",
			given: `
				b>>- b>>- b>>-
				-a-a -a-a aa-a`,
			expected: []string{
				"b>>-b>>-b>>-",
				"-a-a-a-aaa-a",
			},
		},
		{
			name:     "normalize oneliner riff content",
			given:    "a--a b--b",
			expected: []string{"a--ab--b"},
		},
		{
			name:     "normalize empty riff content",
			given:    "",
			expected: []string{""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizeRiffContent(tc.given)
			testutils.AssertSlices(t, tc.expected, actual)
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

func TestParseVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		values   []string
		expected map[string]float64
	}{
		{
			name:   "parse simple vars",
			given:  "note=e2 dur=1",
			values: []string{"c#3", "2"},
			expected: map[string]float64{
				"note": float64(audio.NoteToMidi("c#3")),
				"dur":  2,
			},
		},
		{
			name:   "parse simple vars with default fallback",
			given:  "n=e2 m=60 d=2",
			values: []string{"a3"},
			expected: map[string]float64{
				"n": float64(audio.NoteToMidi("a3")),
				"m": 60,
				"d": 2,
			},
		},
		{
			name:   "parse dependent vars",
			given:  "a=20 b=a+10 c=b+20",
			values: []string{"40"},
			expected: map[string]float64{
				"a": 40,
				"b": 50,
				"c": 70,
			},
		},
		{
			name:   "parse dependent vars with default fallback",
			given:  "a=20 b=a+10 c=b+20",
			values: nil,
			expected: map[string]float64{
				"a": 20,
				"b": 30,
				"c": 50,
			},
		},
		{
			name:   "parse dependent vars with overrides",
			given:  "a=20 b=a+10",
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

func TestApplyVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		vars     map[string]float64
		expected string
	}{
		{
			name:  "apply vars for lane sequence",
			given: ">mid,d _p",
			vars: map[string]float64{
				"mid": 77,
				"d":   3,
				"p":   2,
			},
			expected: ">77,,3, _2,",
		},
		{
			name: "apply vars for riff sequence",
			given: `
			--b- --b-
			aa-a a--a
			`,
			vars: map[string]float64{
				"a": 77,
				"b": 33,
			},
			expected: `
			--33,- --33,-
			77,77,-77, 77,--77,
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := applyVars(tc.given, tc.vars)
			testutils.AssertStrings(t, tc.expected, actual)
		})
	}
}

// OLD STUFF BELOW

func TestApplySeqVars(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		vars     string
		values   []string
		expected string
	}{
		{
			name:     "apply multiple for lane sequence",
			given:    ">midi d_p",
			vars:     "midi=60 d=1 p=2",
			expected: ">60 1_2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := applySeqVars(tc.given, tc.vars, tc.values)
			testutils.AssertStrings(t, tc.expected, actual)
		})
	}
}

func TestParseContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		context  *ParserContext
		expected []string
	}{
		{
			name:  "midi > operator",
			given: ">70",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.MidiToFreq(70)),
			},
		},
		{
			name:  "note : operator",
			given: ":e#3",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.NoteToFreq("e#3")),
			},
		},
		{
			name:  "freq ~ operator",
			given: "~440.17",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", 440.17),
			},
		},
		{
			name:  "wait _ operator",
			given: "~40_ ~60 2_ ~80 _3 ~100 4/2",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", 40.0),
				fmt.Sprintf("%fhz 1t 2d", 60.0),
				fmt.Sprintf("%fhz 3t 1d", 80.0),
				fmt.Sprintf("%fhz 6t 2d", 100.0),
			},
		},
		{
			name:  "last | operator",
			given: "~40 2_1 |*2 /2",
			expected: []string{
				fmt.Sprintf("%fhz 0t 2d", 40.0),
				fmt.Sprintf("%fhz 1t 1d", 80.0),
			},
		},
		{
			name:  "link @ operator",
			given: "@chord5 e2 2",
			context: &ParserContext{
				ResolveFunc: func(name string, args []string) (*ir.Program, error) {
					return ParseContent(":e2 2 |+7 |+5", nil)
				},
			},
			expected: []string{
				fmt.Sprintf("%fhz 0t 2d", audio.MidiToFreq(audio.NoteToMidi("e2")+0)),
				fmt.Sprintf("%fhz 0t 2d", audio.MidiToFreq(audio.NoteToMidi("e2")+7)),
				fmt.Sprintf("%fhz 0t 2d", audio.MidiToFreq(audio.NoteToMidi("e2")+7+5)),
			},
		},
		{
			name: "check >:~_| operators",
			given: `
				>40*2 3
				:f#4-2 2_
				~440 _4 |880 *2
			`,
			expected: []string{
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(80)),
				fmt.Sprintf("%fhz 0t 2d", audio.NoteToFreq("e4")),
				fmt.Sprintf("%fhz 2t 1d", 440.0),
				fmt.Sprintf("%fhz 6t 2d", 880.0),
			},
		},
		{
			name:  "check | time modifications",
			given: "~60 _1 |*2 +2 |/2 +2",
			expected: []string{
				fmt.Sprintf("%fhz %dt %dd", 60.0, 0, 1),
				fmt.Sprintf("%fhz %dt %dd", 120.0, 0, 1),
			},
		},
		{
			name:  "check :| for chord composition",
			given: ":b#4 3 |+7 |+5",
			expected: []string{
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+0)),
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+7)),
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+7+5)),
			},
		},
		{
			name:  "check :| for arpeggio composition",
			given: ":c2 2_1 |+7 |+5",
			expected: []string{
				fmt.Sprintf("%fhz 0t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+0)),
				fmt.Sprintf("%fhz 1t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+7)),
				fmt.Sprintf("%fhz 2t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+7+5)),
			},
		},
		{
			name: "multiple linked lane sequences with internal delays",
			given: `
				~100 _1
				@first _2
				@second _2
				~200
			`,
			context: &ParserContext{
				ResolveFunc: func(name string, args []string) (*ir.Program, error) {
					switch name {
					case "first":
						return ParseContent("~60 6_6 ~66_", nil)
					case "second":
						return ParseContent("~40 4_4 ~44_", nil)
					}
					return nil, fmt.Errorf("No sequence found: %s", name)
				},
			},
			expected: []string{
				fmt.Sprintf("%fhz %dt %dd", 100.0, 0, 1),
				fmt.Sprintf("%fhz %dt %dd", 60.0, 1, 6),
				fmt.Sprintf("%fhz %dt %dd", 66.0, 7, 1),
				fmt.Sprintf("%fhz %dt %dd", 40.0, 3, 4),
				fmt.Sprintf("%fhz %dt %dd", 44.0, 7, 1),
				fmt.Sprintf("%fhz %dt %dd", 200.0, 5, 1),
			},
		},
	}

	for _, tc := range tests {
		if tc.name != "multiple linked lane sequences with internal delays" {
			t.Skip()
		}
		t.Run(tc.name, func(t *testing.T) {
			irp, err := ParseContent(tc.given, tc.context)
			testutils.AssertErrNil(t, err)

			instructions := irp.Instructions()
			testutils.AssertInts(t, len(instructions), len(tc.expected))

			for i, exp := range tc.expected {
				ins := instructions[i]
				act := fmt.Sprintf("%fhz %vt %vd", ins.Midi, ins.Time, ins.Dur)
				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}

// benchmarks

func BenchmarkNormalizeLaneContent(b *testing.B) {
	tests := []string{
		// 203.1 ns/op	     144 B/op	       2 allocs/op
		`
		>40*2 3
		:f#4-2 2_
		~440 _4 |880 *2
		`,
		// 202.7 ns/op	     176 B/op	       2 allocs/op
		">60_ :e#3_2~440.0 2_3 |*2_",
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				normalizeLaneContent(tc)
			}
		})
	}
}

func BenchmarkNormalizeRiffContent(b *testing.B) {
	tests := []string{
		// 106.9 ns/op	      64 B/op	       2 allocs/op
		`
		b>>- b>>- b>>-
		-a-a -a-a aa-a
		`,
		// 68.21 ns/op	      48 B/op	       2 allocs/op
		`--b>>
		-a>a>`,
		// 53.71 ns/op	      24 B/op	       2 allocs/op
		"a--a b--b",
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case_%d", i), func(b *testing.B) {
			for b.Loop() {
				normalizeRiffContent(tc)
			}
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

func BenchmarkParseVars(b *testing.B) {
	tests := []struct {
		s      string
		values []string
	}{
		// 182.7 ns/op	     288 B/op	       3 allocs/op
		{s: "note=e2 dur=1", values: []string{"c#3", "2"}},
		// 259.5 ns/op	     304 B/op	       3 allocs/op
		{s: "n=e2 m=60 d=2", values: []string{"a3"}},
		// 470.7 ns/op	     304 B/op	       3 allocs/op
		{s: "a=20 b=a+10 c=b+20", values: []string{"40"}},
	}

	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				parseVars(tc.s, tc.values)
			}
		})
	}
}
