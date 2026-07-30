package renderer

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bufio"
	"io"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

type ActiveNote struct {
	Instruction *ir.Instruction
	StartTime   time.Time
}

type WAVRenderer struct {
	wav         *audio.WAV
	activeNotes []ActiveNote
	notesMutex  sync.Mutex
	startTime   time.Time
}

func NewWAVRenderer(sampleRate, channels int) *WAVRenderer {
	return &WAVRenderer{
		wav:         audio.NewWAV(sampleRate, channels),
		activeNotes: make([]ActiveNote, 0),
		startTime:   time.Now(),
	}
}

// Stream generates continuous audio from IR instruction stream
func (wr *WAVRenderer) Stream(r io.Reader, w io.Writer) error {
	// Write WAV header
	wr.wav.WriteHeader(w, 0)

	// Start goroutine to read IR events
	go wr.readInstructions(r)

	// Generate continuous audio at sample rate
	chunkDuration := 0.01 // 10ms chunks
	ticker := time.NewTicker(time.Duration(chunkDuration * float64(time.Second)))
	defer ticker.Stop()

	for range ticker.C {
		wr.generateAndWriteChunk(chunkDuration, w)
	}

	return nil
}

// readInstructions reads IR events from stdin and adds to active notes
func (wr *WAVRenderer) readInstructions(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		ins, err := ir.Deserialize(line)
		if err != nil {
			log.Printf("[ERROR] deserialize: %v", err)
			continue
		}

		// Add to active notes
		wr.notesMutex.Lock()
		wr.activeNotes = append(wr.activeNotes, ActiveNote{
			Instruction: ins,
			StartTime:   time.Now(),
		})
		wr.notesMutex.Unlock()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] scanner: %v", err)
	}
}

// TODO: remove
func (wr *WAVRenderer) Samples(ins *ir.Instruction) []int16 {
	sr := wr.wav.SampleRate()
	out := make([]int16, int(ins.Dur*float64(sr)))
	freq := audio.MidfToFreq(ins.Midi)

	wave := audio.WaveSine
	if ins.Patch != nil && ins.Patch.WaveFunc != nil {
		wave = ins.Patch.WaveFunc
	}

	clip := float64(math.MaxInt16)
	amp := ins.Vol * math.MaxInt16
	step := freq / float64(sr)
	phase := 0.0
	for i := range out {
		v := wave(phase) * amp
		out[i] = int16(math.Tanh(v/clip) * clip) // soft-clipping
		phase += step
		if phase >= 1 {
			phase -= 1
		}
	}
	return out
}

// generateAndWriteChunk creates and writes one chunk of audio
func (wr *WAVRenderer) generateAndWriteChunk(durationSec float64, w io.Writer) {
	sr := wr.wav.SampleRate()
	chunkSize := int(durationSec * float64(sr))
	chunk := make([]int16, chunkSize)

	wr.notesMutex.Lock()

	// Remove finished notes and generate samples for active ones
	activeNotes := wr.activeNotes[:0]
	for _, note := range wr.activeNotes {
		elapsed := time.Since(note.StartTime).Seconds()

		// Remove if finished
		if elapsed >= note.Instruction.Dur {
			continue
		}

		activeNotes = append(activeNotes, note)

		// Generate and mix samples for this note
		wr.mixNote(chunk, note.Instruction, elapsed, durationSec)
	}
	wr.activeNotes = activeNotes

	wr.notesMutex.Unlock()

	// Write chunk (silence if no active notes)
	wr.wav.WriteSamples(w, chunk)
	if f, ok := w.(*os.File); ok {
		f.Sync()
	}
}

// mixNote generates samples for one note and mixes into chunk
func (wr *WAVRenderer) mixNote(chunk []int16, ins *ir.Instruction, offsetSec float64, durationSec float64) {
	sr := wr.wav.SampleRate()
	chunkSize := len(chunk)

	freq := audio.MidfToFreq(ins.Midi)
	wave := audio.WaveSine
	if ins.Patch != nil && ins.Patch.WaveFunc != nil {
		wave = ins.Patch.WaveFunc
	}

	amp := ins.Vol * math.MaxInt16
	step := freq / float64(sr)
	phase := math.Mod(offsetSec*freq, 1.0)

	for i := range chunkSize {
		sample := int16(wave(phase) * amp)

		// Mix with soft clipping
		mixed := int(chunk[i]) + int(sample)
		if mixed > math.MaxInt16 {
			chunk[i] = math.MaxInt16
		} else if mixed < math.MinInt16 {
			chunk[i] = math.MinInt16
		} else {
			chunk[i] = int16(mixed)
		}

		phase += step
		if phase >= 1 {
			phase -= 1
		}
	}
}
