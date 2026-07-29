package renderer

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bufio"
	"io"
	"log"
	"math"
	"os"
	"time"
)

type WAVRenderer struct {
	wav *audio.WAV
}

func NewWAVRenderer(sampleRate, channels int) *WAVRenderer {
	return &WAVRenderer{
		wav: audio.NewWAV(sampleRate, channels),
	}
}

func (wr *WAVRenderer) Render(r io.Reader, w io.Writer) error {
	startTime := time.Now()
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		ins, err := ir.Deserialize(line)
		if err != nil {
			log.Printf("[ERROR] deserealization: %v", err)
			continue
		}

		log.Printf("[+%.3fs] %v\n", time.Since(startTime).Seconds(), ins)
		samples := wr.Samples(ins)
		wr.wav.WriteSamples(w, samples)
		if f, ok := w.(*os.File); ok {
			f.Sync()
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (wr *WAVRenderer) Render2(samples []int16, w io.Writer) error {
	err := wr.wav.WriteSamples(w, samples)
	if err != nil {
		return err
	}
	if f, ok := w.(*os.File); ok {
		f.Sync()
	}
	return nil
}

func (wr *WAVRenderer) Start(w io.Writer) {
	wr.wav.WriteHeader(w, 0)
}

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

// WriteChunk writes a fixed-duration chunk of samples
// If instructions provided, mix them. Otherwise write silence.
// currentTime is the current playback position in seconds
func (wr *WAVRenderer) WriteChunk(durationSec float64, currentTime float64, instructions []*ir.Instruction, w io.Writer) error {
	sr := wr.wav.SampleRate()
	chunkSize := int(durationSec * float64(sr))
	chunk := make([]int16, chunkSize)

	// Generate and mix samples for all active instructions
	for _, ins := range instructions {
		// Calculate offset within this note
		offsetInNote := currentTime - ins.Time
		if offsetInNote < 0 {
			continue // Not started yet
		}

		// Generate samples for this chunk starting from the offset
		samples := wr.SamplesAtOffset(ins, offsetInNote, durationSec)

		// Mix into chunk
		for i := 0; i < chunkSize && i < len(samples); i++ {
			mixed := int(chunk[i]) + int(samples[i])
			if mixed > math.MaxInt16 {
				chunk[i] = math.MaxInt16
			} else if mixed < math.MinInt16 {
				chunk[i] = math.MinInt16
			} else {
				chunk[i] = int16(mixed)
			}
		}
	}

	// Write chunk (silence if no instructions)
	return wr.wav.WriteSamples(w, chunk)
}

// SamplesAtOffset generates samples for a specific time range within an instruction
func (wr *WAVRenderer) SamplesAtOffset(ins *ir.Instruction, offsetSec float64, durationSec float64) []int16 {
	sr := wr.wav.SampleRate()
	chunkSize := int(durationSec * float64(sr))
	out := make([]int16, chunkSize)

	freq := audio.MidfToFreq(ins.Midi)
	wave := audio.WaveSine
	if ins.Patch != nil && ins.Patch.WaveFunc != nil {
		wave = ins.Patch.WaveFunc
	}

	clip := float64(math.MaxInt16)
	amp := ins.Vol * math.MaxInt16
	step := freq / float64(sr)

	// Start phase based on offset
	phase := math.Mod(offsetSec*freq, 1.0)

	for i := range out {
		v := wave(phase) * amp
		out[i] = int16(math.Tanh(v/clip) * clip)
		phase += step
		if phase >= 1 {
			phase -= 1
		}
	}

	return out
}
