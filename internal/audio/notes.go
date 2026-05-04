package audio

import (
	"math"
	"strconv"
)

const C4freq = 261.6255653006 // base

var noteToFreq = make(map[string]float64)
var noteToMidi = make(map[string]int)
var midiToFreq = make(map[int]float64)
var midiToNote = make(map[int]string)

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
	for noteName, idx := range noteIndex {
		for octave := 0; octave <= 9; octave++ {
			semitones := (octave-4)*12 + idx
			freq := C4freq * math.Pow(2, float64(semitones)/12)
			note := noteName + strconv.Itoa(octave)
			midi := semitones + 60
			noteToFreq[note] = freq
			noteToMidi[note] = midi
			midiToFreq[midi] = freq
			midiToNote[midi] = note
		}
	}
}

// Get frequency by note name
func NoteToFreq(note string) float64 {
	if f, ok := noteToFreq[note]; ok {
		return f
	}
	return C4freq
}

// Get midi number by note name
func NoteToMidi(note string) int {
	if i, ok := noteToMidi[note]; ok {
		return i
	}
	return -1
}

// Get frequency by midi number
func MidiToFreq(idx int) float64 {
	if f, ok := midiToFreq[idx]; ok {
		return f
	}
	return C4freq
}

// Get note name by midi number
func MidiToNote(idx int) string {
	if f, ok := midiToNote[idx]; ok {
		return f
	}
	return "c4"
}

// Get midi float number by freq
func FreqToMidi(freq float64) float64 {
	return 60 + 12*math.Log2(freq/C4freq)
}

// Transpose frequency by semitone steps
func TransposeFreq(freq, steps float64) float64 {
	return freq * math.Pow(2, steps/12)
}
