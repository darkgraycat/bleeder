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

	out := p.getSamples(instructions, totalSamples, audio.WaveSaw)
	p.wav.Append(out)

	f, err := os.CreateTemp("", "out*.wav")
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

func (p *WAVPlayer) getSamples(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	buf := make([]float64, total)
	out := make([]int16, total)
	clip := float64(math.MaxInt16)
	for _, in := range instructions {
		offset := int(in.Time * sr)
		// TODO
		// samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		samples := p.wav.GenerateSamples2(in.Freq, in.Dur, in.Vol, 0.03, 0.06, wave)
		for i, s := range samples {
			buf[offset+i] += float64(s)
		}
	}
	for i, s := range buf {
		s = math.Tanh(s / clip) * clip // soft-clipping
		out[i] = int16(s)
	}
	return out
}
