package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"math"
	"os"
)

type WAVPlayer struct {
	wav *audio.WAV
}

func NewWAVPlayer(sampleRate, channels int) *WAVPlayer {
	return &WAVPlayer{
		wav: audio.NewWAV(sampleRate, channels),
	}
}

func (p *WAVPlayer) Play(irp *ir.Program, start, end int) error {
	logs.Info("Play")
	sr := p.wav.SampleRate()
	instructions := irp.Instructions()
	_, duration := irp.MinMaxTime()
	totalSamples := int(duration * float64(sr))
	logs.Info("Total instructions %d", irp.Length())
	logs.Info("Total samples %d", totalSamples)
	logs.Info("Total duration %f", duration)

	logs.Debug("get samples")
	wave := audio.WaveFuncMix(audio.WaveCubic, audio.WaveSoftSquare)
	out := p.getSamples(instructions, totalSamples, wave)

	logs.Debug("append samples")
	p.wav.Append(out)

	return p.wav.Write(os.Stdout)
}

func (p *WAVPlayer) Stop() error {
	return nil
}

func (p *WAVPlayer) getSamples(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	forDebugTimeTempVariableAtAll := 1.0

	sr := p.wav.SampleRate()

	total = total / int(forDebugTimeTempVariableAtAll)

	buf := make([]float64, total)
	out := make([]int16, total)
	clip := float64(math.MaxInt16)
	logs.Debug("geting samples")

	for _, ins := range instructions {
		offset := int(ins.Time * float64(sr) / forDebugTimeTempVariableAtAll)
		dur := ins.Dur / forDebugTimeTempVariableAtAll
		// TODO
		// samples := p.wav.GenerateSamples(ins.Freq, ins.Dur, ins.Vol, wave)
		samples := p.wav.GenerateSamplesEnvelope(audio.MidfToFreq(ins.Midi), dur, ins.Vol, 0.01, 0.01, wave)
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
