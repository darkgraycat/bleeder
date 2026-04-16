package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"fmt"
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
	totalSamples := int((pr.Last().Time + pr.Last().Dur) * sr)
	buf := make([]int16, totalSamples+1)

	for i, in := range instructions {
		fmt.Printf("%d %f - %f %f\n", i, in.Freq, in.Dur, in.Time)

		offset := int(in.Time * sr)
		samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, audio.WaveSine)
		for i, s := range samples {
			buf[offset+i] += s // += because notes can overlap
		}
	}
	p.wav.Append(buf)

	f, err := os.CreateTemp("", "note*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	p.wav.Write(f)
	return exec.Command("afplay", "-v", "0.3", f.Name()).Run()
}

func (p *WAVPlayer) Stop() error {
	return nil
}
