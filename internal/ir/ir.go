package ir

// Intermediate Representation Program
type Program struct {
	insts []*Instruction
	index map[*Instruction]int
}

// Create new Program instance
func NewProgram() *Program {
	return &Program{
		insts: make([]*Instruction, 0),
		index: make(map[*Instruction]int),
	}
}

// Get an array of instructions
func (p *Program) GetInstructions() []*Instruction {
	return p.insts
}

// Add next Instruction pointer to the end
func (p *Program) Add(inst *Instruction) {
	p.insts = append(p.insts, inst)
	p.index[inst] = len(p.insts) - 1
}

// Merge another Program into current one
func (p *Program) Merge(src *Program) {
	offset := len(p.insts)
	p.insts = append(p.insts, src.insts...)
	for i, inst := range src.insts {
		p.index[inst] = offset + i
	}
}

// Get the number of Instructions in Program
func (p *Program) Length() int {
	return len(p.insts)
}

// Get first Instruction
func (p *Program) First() *Instruction {
	if len(p.insts) == 0 {
		return nil
	}
	return p.insts[0]
}

// Get last Instruction
func (p *Program) Last() *Instruction {
	if len(p.insts) == 0 {
		return nil
	}
	return p.insts[len(p.insts)-1]
}

// Get next Instruction after provided one
func (p *Program) Next(inst *Instruction) *Instruction {
	if idx, ok := p.index[inst]; ok && idx+1 < len(p.insts) {
		return p.insts[idx+1]
	}
	return nil
}

// Get previos Instruction after provided one
func (p *Program) Prev(inst *Instruction) *Instruction {
	if idx, ok := p.index[inst]; ok && idx-1 >= 0 {
		return p.insts[idx-1]
	}
	return nil
}

// Instruction is a basic unit of Intermediate Representation
type Instruction struct {
	Note  int     // integer representation of note (C4 is 60)
	Freq  int     // frequence of sound to play
	Dur   float32 // duration of sound
	Vol   float32 // volume of sound
	Start float32 // time to start
	Info  string  // additional information
}
