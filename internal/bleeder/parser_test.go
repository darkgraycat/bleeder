package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/testutils"
	"fmt"
	"testing"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		content  string
		context  *ParserContext
		expected []string
	}{
		{
			name:    "midi > operator",
			content: ">60",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.MidiToFreq(60)),
			},
		},
		{
			name:    "note : operator",
			content: ":e#3",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.NoteToFreq("e#3")),
			},
		},
		{
			name:    "freq ~ operator",
			content: "~440.17",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", 440.17),
			},
		},
		{
			name:    "wait _ operator",
			content: "~40_ ~60 2_ ~80 _3 ~100 4/2",
			expected: []string{
				fmt.Sprintf("%fhz 0t 1d", 40.0),
				fmt.Sprintf("%fhz 1t 2d", 60.0),
				fmt.Sprintf("%fhz 3t 1d", 80.0),
				fmt.Sprintf("%fhz 6t 2d", 100.0),
			},
		},
		{
			name:    "last | operator",
			content: "~40 2_1 |*2 /2",
			expected: []string{
				fmt.Sprintf("%fhz 0t 2d", 40.0),
				fmt.Sprintf("%fhz 1t 1d", 80.0),
			},
		},
		{
			name:    "link @ operator",
			content: "@chord5 e2 2",
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
			content: `
				>40*2 3
				:f#4-2 2_
				~440 _4 |/2 *2
			`,
			expected: []string{
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(80)),
				fmt.Sprintf("%fhz 0t 2d", audio.NoteToFreq("e4")),
				fmt.Sprintf("%fhz 2t 1d", 440.0),
				fmt.Sprintf("%fhz 6t 2d", 220.0),
			},
		},
		{
			name:    "check :| for chord composition",
			content: ":b#4 3 |+7 |+5",
			expected: []string{
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+0)),
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+7)),
				fmt.Sprintf("%fhz 0t 3d", audio.MidiToFreq(audio.NoteToMidi("b#4")+7+5)),
			},
		},
		{
			name:    "check :| for arpeggio composition",
			content: ":c2 2_1 |+7 |+5",
			expected: []string{
				fmt.Sprintf("%fhz 0t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+0)),
				fmt.Sprintf("%fhz 1t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+7)),
				fmt.Sprintf("%fhz 2t 2d", audio.MidiToFreq(audio.NoteToMidi("c2")+7+5)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			irp, err := ParseContent(tc.content, tc.context)
			irp.Shift(tc.offset)
			testutils.AssertErrNil(t, err)

			instructions := irp.Instructions()
			testutils.AssertInts(t, len(instructions), len(tc.expected))

			for i, ins := range instructions {
				exp := tc.expected[i]
				act := fmt.Sprintf("%fhz %vt %vd", ins.Freq, ins.Time, ins.Dur)

				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}

// func TestExpandArgs(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		argsRaw string
// 		content string
// 		result  string
// 	}{
// 		{
// 			name:    "expand modified argument",
// 			argsRaw: "a:60 b:a/2",
// 			content: ">a_ >b_",
// 			result:  ">60_ >30_",
// 		},
// 		// {
// 		// 	name:    "expand two plain arguments",
// 		// 	argsRaw: "note:e2 d:2",
// 		// 	content: ":note d_",
// 		// 	result:  ":e2 2_",
// 		// },
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			args, err := ParseRawArgs(tc.argsRaw)
// 			testutils.AssertErrNil(t, err)

// 			exp := tc.result
// 			act, err := ExpandArgs(tc.content, args)
// 			testutils.AssertErrNil(t, err)

// 			testutils.AssertStrings(t, exp, act)
// 		})
// 	}
// }
