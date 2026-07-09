package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"log"
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
	sr := p.wav.SampleRate()
	instructions := irp.Instructions()
	_, duration := irp.MinMaxTime()
	totalSamples := int(duration * float64(sr))
	log.Printf("[INIT:PLAY] instructions %d, samples %d, duration %f\n", irp.Length(), totalSamples, duration)

	wave := audio.WaveFuncMix(audio.WaveParabola, audio.WaveSoftSquare)
	out := p.getSamples(instructions, totalSamples, wave)

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
	for i, s := range buf {
		s = math.Tanh(s/clip) * clip // soft-clipping
		out[i] = int16(s)
	}
	return out
}
