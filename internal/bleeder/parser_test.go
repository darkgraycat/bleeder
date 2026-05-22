package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/testutils"
	"fmt"
	"strings"
	"testing"
)

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
			name: "check | time modifications",
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
				act := fmt.Sprintf("%fhz %vt %vd", ins.Freq, ins.Time, ins.Dur)
				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}

func TestNormalizeLaneContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected []string
	}{
		{
			name:  "normalize oneline string",
			given: ">60_ :e#3_2~440.0 2_3 |*2_",
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
				>40*2 3
				:f#4-2 2_
				~440 _4 |880 *2
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
		// {
		// 	name:     "normalize invalid content",
		// 	content:  "60 :32",
		// 	expected: nil,
		// },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizeLaneContent(tc.given)
			for i, v := range actual {
				actual[i] = strings.TrimSpace(v)
			}
			testutils.AssertSlices(t, tc.expected, actual)
			for i, exp := range tc.expected {
				act := strings.TrimSpace(actual[i])
				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}

func TestTokenizeLaneContent(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected [][]string
	}{
		{
			name:  "tokenize oneline string",
			given: ">60_ :e#3_2~440.0 2_3 |*2_",
			expected: [][]string{
				{">", "60"},
				{"_"},
				{":", "e#3"},
				{"_", "2"},
				{"~", "440.0", "2"},
				{"_", "3"},
				{"|", "*2"},
				{"_"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tokenizeLaneContent(tc.given)
			for i, exp := range tc.expected {
				testutils.AssertSlices(t, exp, actual[i])
			}
		})
	}
}
