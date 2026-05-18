package ir

import "fmt"

// Intermediate Representation Program
type Program struct {
	timeScale    float64              // multiplier applied to instruction tick values (Time, Dur) during playback; defaults to 1
	instructions []*Instruction       // instructions array
	indexesCache map[*Instruction]int // instructions cache
}

// Create new Program instance
func NewProgram() *Program {
	return &Program{
		timeScale:    1, // TODO: i think I can remove it from IR and move into Bleeder itself
		instructions: make([]*Instruction, 0),
		indexesCache: make(map[*Instruction]int),
	}
}

// Get time scale
func (p *Program) TimeScale() float64 {
	return p.timeScale
}

// Get an array of instructions
func (p *Program) Instructions() []*Instruction {
	return p.instructions
}

// Add next Instruction pointer(s) to the end
func (p *Program) Add(ins ...*Instruction) {
	offset := len(p.instructions)
	p.instructions = append(p.instructions, ins...)
	for i, ins := range ins {
		p.indexesCache[ins] = offset + i
	}
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

// Cut Instructions into new Program
func (p *Program) Cut(start, end int) *Program {
	sliced := p.instructions[start:end]
	np := NewProgram()
	for i, ins := range sliced {
		np.instructions = append(np.instructions, ins)
		np.indexesCache[ins] = i
	}
	return np
}

// Merge one or more Programs into current one
func (p *Program) Merge(irp ...*Program) {
	for _, src := range irp {
		offset := len(p.instructions)
		p.instructions = append(p.instructions, src.instructions...)
		for i, ins := range src.instructions {
			p.indexesCache[ins] = offset + i
		}
	}
}

// Get the number of Instructions in Program
func (p *Program) Length() int {
	return len(p.instructions)
}

// Get duration of whole Program
func (p *Program) Duration() int {
	dur := 0
	for _, ins := range p.instructions {
		end := ins.Time + ins.Dur
		if end > dur {
			dur = end
		}
	}
	return dur
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

// Get next Instruction after provided one
func (p *Program) Next(ins *Instruction) *Instruction {
	if idx, ok := p.indexesCache[ins]; ok && idx+1 < len(p.instructions) {
		return p.instructions[idx+1]
	}
	return nil
}

// Get previos Instruction after provided one
func (p *Program) Prev(ins *Instruction) *Instruction {
	if idx, ok := p.indexesCache[ins]; ok && idx-1 >= 0 {
		return p.instructions[idx-1]
	}
	return nil
}

// Instruction is a basic unit of Intermediate Representation
type Instruction struct {
	Freq float64 // frequence in Hz
	Vol  float64 // volume 0.0 > 1.0
	Dur  int     // duration in ticks
	Time int     // start time in ticks
	Info string  // additional information
}

// Format Instruction into string
func (ins Instruction) String() string {
	return fmt.Sprintf("Freq=%f Vol=%f Dur=%d Time=%d Info=%s",
		ins.Freq, ins.Vol, ins.Dur, ins.Time, ins.Info)
}
