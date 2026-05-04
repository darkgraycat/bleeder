package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/shared/testutils"
	"fmt"
	"testing"
)

func TestParseContentOperators(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		instructions []string
	}{
		{
			name:    "midi > operator",
			content: ">60",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.MidiToFreq(60)),
			},
		},
		{
			name:    "note : operator",
			content: ":e#3",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.NoteToFreq("e#3")),
			},
		},
		{
			name:    "freq ~ operator",
			content: "~440.17",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 1d", 440.17),
			},
		},
		{
			name:    "last | operator",
			content: "~40 2_1 |*2 /2",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 2d", 40.0),
				fmt.Sprintf("%fhz 1t 1d", 80.0),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr, err := ParseContent(tc.content, 0)
			testutils.AssertErrNil(t, err)

			instructions := pr.Instructions()
			testutils.AssertInts(t, len(instructions), len(tc.instructions))

			for i, in := range instructions {
				exp := tc.instructions[i]
				act := fmt.Sprintf("%fhz %vt %vd", in.Freq, in.Time, in.Dur)

				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}
