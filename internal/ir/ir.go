package ir

import (
	"bleeder/internal/audio"
	"fmt"
	"math"
	"slices"
)

// Intermediate Representation Program
type Program struct {
	instructions []*Instruction // instructions array
}

// Create new Program instance
func NewProgram() *Program {
	return &Program{
		instructions: make([]*Instruction, 0),
	}
}

// Get an array of instructions
func (p *Program) Instructions() []*Instruction {
	return p.instructions
}

// Add next Instruction pointer(s) to the end
func (p *Program) Add(ins ...*Instruction) {
	p.instructions = append(p.instructions, ins...)
}

// Copy returns a deep copy of the Program with new Instruction pointers
func (p *Program) Copy() *Program {
	np := NewProgram()
	for _, ins := range p.instructions {
		cp := *ins
		np.Add(&cp)
	}
	return np
}

// Merge one or more Programs into current one
func (p *Program) Merge(irp ...*Program) {
	for _, src := range irp {
		p.instructions = append(p.instructions, src.instructions...)
	}
}

// Get the number of Instructions in Program
func (p *Program) Length() int {
	return len(p.instructions)
}

// Get duration of whole Program
func (p *Program) Duration() float64 {
	minTime, maxTime := p.MinMaxTime()
	return maxTime - minTime
}

// Get start and end time of Program
func (p *Program) MinMaxTime() (float64, float64) {
	minTime := math.MaxFloat64
	maxTime := 0.0
	for _, ins := range p.instructions {
		if ins.Time < minTime {
			minTime = ins.Time
		}
		end := ins.Time + ins.Dur
		if end > maxTime {
			maxTime = end
		}
	}
	return minTime, maxTime
}

// Shift start time of each instruction
func (p *Program) Shift(offset float64) {
	if offset <= 0 || math.IsNaN(offset) {
		return
	}
	for _, ins := range p.instructions {
		ins.Time += offset
	}
}

// Stretch program in time
func (p *Program) Stretch(factor float64) {
	if factor <= 0 || math.IsNaN(factor) {
		return
	}
	for _, ins := range p.instructions {
		ins.Dur *= factor
		ins.Time *= factor
	}
}

// Transpose all instructions
func (p *Program) Transpose(semitones float64) {
	if semitones == 0 || math.IsNaN(semitones) {
		return
	}
	for _, ins := range p.instructions {
		ins.Midi += semitones
	}
}

// Sort instructions by absolute time
func (p *Program) Sort() {
	slices.SortFunc(p.instructions, func(a, b *Instruction) int {
		if a.Time < b.Time {
			return -1
		}
		if a.Time > b.Time {
			return 1
		}
		return 0
	})
}

// Get first Instruction
func (p *Program) First() *Instruction {
	if l := len(p.instructions); l > 0 {
		return p.instructions[0]
	}
	return nil
}

// Get last Instruction
func (p *Program) Last() *Instruction {
	if l := len(p.instructions); l > 0 {
		return p.instructions[l-1]
	}
	return nil
}

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
	return fmt.Sprintf("Midi=%f Vol=%f Dur=%f Time=%f Info=%s",
		ins.Midi, ins.Vol, ins.Dur, ins.Time, ins.Info)
}

// Instruction shape of the sound
type Patch struct {
	Name     string         // patch name
	WaveFunc audio.WaveFunc // wave function to use
}
