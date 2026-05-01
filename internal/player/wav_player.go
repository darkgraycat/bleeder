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

	out := p.getSamples(instructions, totalSamples, audio.WaveTriangle)
	p.wav.Append(out)

	f, err := os.CreateTemp("", "out*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()


	fmt.Printf("out[22050] = %d, wav.Samples[22050] = %d\n", out[22050], p.wav.Samples()[22050])

	p.wav.Write(f)
	return exec.Command("afplay", "-v", "0.5", f.Name()).Run()
}

func (p *WAVPlayer) Stop() error {
	return nil
}

func (p *WAVPlayer) getSamples(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	density := make([]int, total)
	buf := make([]float64, total)
	out := make([]int16, total)
	for idx, in := range instructions {
		offset := int(in.Time * sr)
		// TODO
		// samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		samples := p.wav.GenerateSamples2(in.Freq, in.Dur, in.Vol, 0.02, 0.02, wave)
		for i, s := range samples {
			if offset+i == 22050 {
				fmt.Printf("instruction %d adding to buf[22050]: s=%d, buf before=%f\n",
					idx, s, buf[offset+i])
			}
			buf[offset+i] += float64(s)
			density[offset+i]++
		}
	}
	for i, s := range buf {
		d := float64(density[i])
		// NaN when d=0, int16(NaN)=0 in Go
		if i == 22050 {
			fmt.Printf("i=22050: s=%f, d=%f, s/d=%f, int16=%d\n",
				s, d, s/d, int16(s/d))
		}
		out[i] = int16(s / d)
	}
	fmt.Printf("density[22050] = %d, final = %d\n", density[22050], out[22050])
	return out
}
