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
	startTime   time.Time
	mu          sync.Mutex
}

func NewWAVRenderer(sampleRate, channels int) *WAVRenderer {
	return &WAVRenderer{
		wav:         audio.NewWAV(sampleRate, channels),
		activeNotes: make([]ActiveNote, 0),
		startTime:   time.Now(),
	}
}

// Render generates continuous audio from IR instruction stream
func (wr *WAVRenderer) Render(r io.Reader, w io.Writer) error {
	// Write WAV header
	wr.wav.WriteHeader(w, 0)

	// Start goroutine to read IR events
	go wr.readInstructions(r)

	// Generate continuous audio at sample rate
	chunkDuration := 0.01 // 10ms chunks
	ticker := time.NewTicker(time.Duration(chunkDuration * float64(time.Second)))
	defer ticker.Stop()

	for range ticker.C {
		wr.writeInstructions(chunkDuration, w)
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
		wr.mu.Lock()
		wr.activeNotes = append(wr.activeNotes, ActiveNote{
			Instruction: ins,
			StartTime:   time.Now(),
		})
		wr.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ERROR] scanner: %v", err)
	}
}

// writeInstructions creates and writes one chunk of audio
func (wr *WAVRenderer) writeInstructions(durationSec float64, w io.Writer) {
	sr := wr.wav.SampleRate()
	chunkSize := int(durationSec * float64(sr))
	chunk := make([]int16, chunkSize)

	wr.mu.Lock()

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
		wr.mixNote(chunk, note.Instruction, elapsed)
	}
	wr.activeNotes = activeNotes

	wr.mu.Unlock()

	// Write chunk (silence if no active notes)
	wr.wav.WriteSamples(w, chunk)
	if f, ok := w.(*os.File); ok {
		f.Sync()
	}
}

// mixNote generates samples for one note and mixes into chunk
func (wr *WAVRenderer) mixNote(chunk []int16, ins *ir.Instruction, offsetSec float64) {
	sr := float64(wr.wav.SampleRate())
	chunkSize := len(chunk)

	freq := audio.MidfToFreq(ins.Midi)
	wave := audio.WaveSine
	if ins.Patch != nil && ins.Patch.WaveFunc != nil {
		wave = ins.Patch.WaveFunc
	}

	amp := ins.Vol * math.MaxInt16
	step := freq / sr
	phase := math.Mod(offsetSec*freq, 1.0)

	attack := int(sr * 0.02) // TODO: move to ADSR
	release := int(sr * 0.03) // TODO: move to ADSR
	attackStep := 1.0 / float64(attack)
	releaseStep := 1.0 / float64(release)

	for i := range chunkSize {
		envelope := 1.0
		if i < attack {
			envelope = attackStep * float64(i)
		} else if i >= chunkSize-release {
			envelope = releaseStep * float64(chunkSize-i)
		}
		sample := int16(wave(phase) * amp * envelope)

		// sample := int16(wave(phase) * amp)

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
