package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"fmt"
	"math"
	"os"
	"os/exec"
)

type WAVPlayer struct {
	wav *audio.WAV
}

func NewWAVPlayer(sampleRate, channels int) *WAVPlayer {
	return &WAVPlayer{
		wav: audio.NewWAV(sampleRate, channels),
	}
}

func (p *WAVPlayer) Play(pr *ir.Program, start, end int) error {
	sr := p.wav.SampleRate()
	instructions := pr.GetInstructions()
	totalSamples := int(math.Ceil((pr.Last().Time + pr.Last().Dur) * sr))
	fmt.Printf("Total samples %d\n", totalSamples)

	for i, in := range instructions {
		fmt.Printf("%d - %f\t%f %f\t%s\n", i, in.Freq, in.Dur, in.Time, in.Info)
	}

	out := p.getSamples3(instructions, totalSamples, audio.WaveSaw)
	p.wav.Append(out)

	f, err := os.CreateTemp("", "note*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	p.wav.Write(f)
	return exec.Command("afplay", "-v", "0.5", f.Name()).Run()
}

func (p *WAVPlayer) Stop() error {
	return nil
}

func (p *WAVPlayer) getSamples1(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	out := make([]int16, total)
	for _, in := range instructions {
		offset := int(in.Time * sr)
		samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		for i, s := range samples {
			out[offset+i] += s
		}
	}
	return out
}

func (p *WAVPlayer) getSamples2(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	buf := make([]float64, total)
	out := make([]int16, total)
	for _, in := range instructions {
		offset := int(in.Time * sr)
		samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		for i, s := range samples {
			buf[offset+i] += float64(s)
		}
	}
	for i, s := range buf {
		clamped := math.Max(-math.MaxInt16, math.Min(math.MaxInt16, s))
		out[i] = int16(clamped)
	}
	return out
}

func (p *WAVPlayer) getSamples3(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	density := make([]int, total)
	buf := make([]float64, total)
	out := make([]int16, total)
	for _, in := range instructions {
		offset := int(in.Time * sr)
		samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		for i, s := range samples {
			buf[offset+i] += float64(s)
			density[offset+i]++
		}
	}
	for i, s := range buf {
		d := float64(density[i])
		// NaN when d=0, int16(NaN)=0 in Go
		out[i] = int16(s / d)
	}
	return out
}
