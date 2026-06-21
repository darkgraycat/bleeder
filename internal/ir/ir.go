package ir

import (
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
func (p *Program) Duration() int {
	minTime := math.MaxInt
	maxTime := 0
	for _, ins := range p.instructions {
		if ins.Time < minTime {
			minTime = ins.Time
		}
		end := ins.Time + ins.Dur
		if end > maxTime {
			maxTime = end
		}
	}
	return maxTime - minTime
}

// Shift start time of each instruction
func (p *Program) Shift(t int) {
	if t <= 0 {
		return
	}
	for _, ins := range p.instructions {
		ins.Time += t
	}
}

// Sort instructions by absolute time
func (p *Program) Sort() {
	slices.SortFunc(p.instructions, func(a, b *Instruction) int {
		return a.Time - b.Time
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
	Midi  float64           // fractional midi
	Dur   int               // duration in ticks
	Vol   float64           // volume 0.0..1.0
	Time  int               // absolute time in ticks
	Info  string            // debug information
	Patch *InstructionPatch // patch to use
}

// Format Instruction into string
func (ins Instruction) String() string {
	return fmt.Sprintf("Midi=%f Vol=%d Dur=%d Time=%d Info=%s",
		ins.Midi, ins.Vol, ins.Dur, ins.Time, ins.Info)
}

// Instruction shape of the sound
type InstructionPatch struct {
	Name string // patch name
}
