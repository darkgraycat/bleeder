package ir

import (
	"bleeder/internal/audio"
	"fmt"
	"strconv"
	"strings"
)

// Instruction is a basic unit of Intermediate Representation
type Instruction struct {
	Midi  float64 // fractional midi
	Dur   float64 // duration in ticks (fractional)
	Vol   float64 // volume 0.0..1.0
	Time  float64 // absolute time in ticks (fractional)
	Info  string  // debug information
	Patch *Patch  // patch to use
}

// Format Instruction into string
func (ins Instruction) String() string {
	return fmt.Sprintf("Midi=%f Dur=%f Vol=%f Time=%f Info=%s Patch=%s",
		ins.Midi, ins.Dur, ins.Vol, ins.Time, ins.Info, ins.Patch)
}

// Serialize instruction to space-separated format
func (ins Instruction) Serialize() string {
	patchName := "sine"
	if ins.Patch != nil {
		patchName = ins.Patch.Name
	}
	return fmt.Sprintf("%f %f %f %f %s",
		ins.Time, ins.Midi, ins.Dur, ins.Vol, patchName)
}

// Deserialize instruction from space-separated format
func Deserialize(src string) (*Instruction, error) {
	parts := strings.Fields(src)
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid instruction format: expected 5 fields, got %d", len(parts))
	}

	time, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Time: %w", err)
	}

	midi, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Midi: %w", err)
	}

	dur, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Dur: %w", err)
	}

	vol, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Vol: %w", err)
	}

	patchName := parts[4]

	waveFunc, ok := audio.WaveFuncs[patchName]
	if !ok {
		waveFunc = audio.WaveSine
	}

	return &Instruction{
		Time:  time,
		Midi:  midi,
		Dur:   dur,
		Vol:   vol,
		Patch: &Patch{Name: patchName, WaveFunc: waveFunc},
	}, nil
}
