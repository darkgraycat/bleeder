package audio

import (
	"math"
	"strconv"
)

const C4 = 261.6255653006 // base

var freqTable = make(map[string]float64)
var indexesTable = make(map[int]float64)
var noteIndexTable = make(map[string]int)

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
			semitones := (octave-4)*12 + idx
			frequency := C4 * math.Pow(2, float64(semitones)/12)
			freqTable[note+strconv.Itoa(octave)] = frequency
			indexesTable[semitones+60] = frequency
			noteIndexTable[note+strconv.Itoa(octave)] = semitones + 60
		}
	}
}

// Get frequency by note name
func FreqByNoteName(note string) float64 {
	if f, ok := freqTable[note]; ok {
		return f
	}
	return C4
}

// Get frequency by note name
func FreqByNoteIndex(idx int) float64 {
	if f, ok := indexesTable[idx]; ok {
		return f
	}
	return C4
}

// Get note index by note name
func GetNoteIndex(note string) int {
	if i, ok := noteIndexTable[note]; ok {
		return i
	}
	return -1
}

// Transpose frequency by semitone steps
func TransposeFreq(freq, steps float64) float64 {
	return freq * math.Pow(2, steps/12)
}
