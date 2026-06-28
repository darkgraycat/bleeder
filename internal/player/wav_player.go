package player

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
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

func (p *WAVPlayer) Play(irp *ir.Program, start, end int) error {
	logs.Info("Play")
	sr := p.wav.SampleRate()
	// ts := ipr.TimeScale()
	instructions := irp.Instructions()
	duration := irp.Duration()
	totalSamples := int(duration * float64(sr))
	logs.Info("Total instructions %d", irp.Length())
	logs.Info("Total samples %d", totalSamples)
	logs.Info("Total duration %f", duration)

	for i, ins := range irp.Instructions() {
		fmt.Printf("[%d] \t%.1f : %.1f : %.1f\n", i, ins.Midi, ins.Dur, ins.Vol)
	}

	logs.Debug("get samples")
	wave := audio.WaveFuncMix(audio.WaveSoftSquare, audio.WaveParabola)
	out := p.getSamples(instructions, totalSamples, wave)

	logs.Debug("append samples")
	p.wav.Append(out)

	logs.Debug("create file")
	// f, err := os.CreateTemp("", "out*.wav")
	f, err := os.Create("test.wav")
	if err != nil {
		return err
	}
	// defer os.Remove(f.Name())
	defer f.Close()

	logs.Debug("write file")
	p.wav.Write(f)

	// TODO: fix bug - it not exiting when its done
	logs.Debug("execute")
	return exec.Command("afplay", "-v", "0.5", f.Name()).Run()
}

func (p *WAVPlayer) Stop() error {
	return nil
}

func (p *WAVPlayer) getSamples2(irp *ir.Program, wave audio.WaveFunc) []int16 {
	sr := p.wav.SampleRate()
	timeScale := 2.0 // TODO
	total := int(float64(sr) * irp.Duration())
	buf := make([]float64, total)
	out := make([]int16, total)
	clip := float64(math.MaxInt16)

	for _, ins := range irp.Instructions() {
		offset := int(ins.Time*timeScale) * sr
		samples := p.wav.GenerateSamplesEnvelope(
			ins.Midi,
			float64(ins.Dur)*timeScale,
			float64(ins.Vol),
			0.03, 0.06,
			wave,
		)
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

func (p *WAVPlayer) getSamples(instructions []*ir.Instruction, total int, wave audio.WaveFunc) []int16 {
	forDebugTimeTempVariableAtAll := 4.0

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
		samples := p.wav.GenerateSamplesEnvelope(audio.MidfToFreq(ins.Midi), dur, ins.Vol, 0.03, 0.06, wave)
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
