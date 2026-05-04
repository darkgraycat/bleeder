package bleeder

import (
	"bleeder/internal/audio"
	"fmt"
	"testing"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		instructions []string
	}{
		{
			name:    "test > op",
			content: ">60",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.MidiToFreq(60)),
			},
		},
		{
			name:    "test : op",
			content: ":e#3",
			instructions: []string{
				fmt.Sprintf("%fhz 0t 1d", audio.NoteToFreq("e#3")),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr, err := ParseContent(tc.content, 0)
			if err != nil {
				t.Fatalf("ParseContent error: %v", err)
			}

			instructions := pr.Instructions()

			if len(instructions) != len(tc.instructions) {
				t.Fatalf("expected %d instructions, got %d",
					len(tc.instructions), len(instructions))
			}

			for i, in := range instructions {
				exp := tc.instructions[i]
				act := fmt.Sprintf("%fhz %vt %vd", in.Freq, in.Time, in.Dur)

				if exp != act {
					t.Fatalf("mismatch at index %d:\nexpected: %s\nactual:   %s",
						i, exp, act)
				}
			}
		})
	}
}
