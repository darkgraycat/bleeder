package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
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
	logs.Info("Play")
	sr := p.wav.SampleRate()
	instructions := pr.Instructions()
	duration := pr.Duration()
	totalSamples := duration * sr
	logs.Info("Total instructions %d", pr.Length())
	logs.Info("Total samples %d", totalSamples)
	logs.Info("Total duration %f", duration)

	logs.Debug("get samples")
	out := p.getSamples(instructions, totalSamples, audio.WaveSaw)

	logs.Debug("append samples")
	p.wav.Append(out)

	logs.Debug("create file")
	f, err := os.CreateTemp("", "out*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logs.Debug("write file")
	p.wav.Write(f)

	logs.Debug("execute")
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
	logs.Debug("geting samples")

	for _, in := range instructions {
		offset := in.Time * sr
		// TODO
		// samples := p.wav.GenerateSamples(in.Freq, in.Dur, in.Vol, wave)
		samples := p.wav.GenerateSamplesEnvelope(in.Freq, float64(in.Dur), float64(in.Vol), 0.03, 0.06, wave)
		for i, s := range samples {
			buf[offset+i] += float64(s)
		}
	}
	logs.Debug("normalising")
	for i, s := range buf {
		s = math.Tanh(s/clip) * clip // soft-clipping
		out[i] = int16(s)
	}
	return out
}
