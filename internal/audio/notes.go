package audio

import (
	"math"
	"strconv"
)

const a4 = 440.0

var freqTable = make(map[string]float64)

func init() {
	noteIndex := map[string]int{
		"c": 0, "c#": 1, "db": 1,
		"d": 2, "d#": 3, "eb": 3,
		"e": 4,
		"f": 5, "f#": 6, "gb": 6,
		"g": 7, "g#": 8, "ab": 8,
		"a": 9, "a#": 10, "bb": 10,
		"b": 11,
	}
	for note, idx := range noteIndex {
		for octave := 0; octave <= 9; octave++ {
			semitones := float64((octave-4)*12 + idx - 9)
			freqTable[note+strconv.Itoa(octave)] = a4 * math.Pow(2, semitones/12)
		}
	}
}

func NoteToFreq(note string) float64 {
	if f, ok := freqTable[note]; ok {
		return f
	}
	return a4
}
