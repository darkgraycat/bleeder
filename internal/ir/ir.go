package ir

// Intermediate Representation Program
type Program struct {
	instr []*Instruction
	index map[*Instruction]int
}

// Create new Program instance
func NewProgram() *Program {
	p := Program{
		instr: make([]*Instruction, 0),
		index: make(map[*Instruction]int),
	}
	return &p
}

// Get an array of instructions
func (p *Program) GetInstructions() []*Instruction {
	return p.instr
}

// Add next Instruction pointer to the end
func (p *Program) Add(insrt *Instruction) {
	p.instr = append(p.instr, insrt)
	p.index[insrt] = len(p.instr) - 1
}

// Merge another Program into current one
func (p *Program) Merge(src *Program) {
	offset := len(p.instr)
	p.instr = append(p.instr, src.instr...)
	for i, instr := range src.instr {
		p.index[instr] = offset + i
	}
}

// Get the number of Instructions in Program
func (p *Program) Length() int {
	return len(p.instr)
}

// Get first Instruction
func (p *Program) First() *Instruction {
	if len(p.instr) == 0 {
		return nil
	}
	return p.instr[0]
}

// Get last Instruction
func (p *Program) Last() *Instruction {
	if len(p.instr) == 0 {
		return nil
	}
	return p.instr[len(p.instr)-1]
}

// Get next Instruction after provided one
func (p *Program) Next(instr *Instruction) *Instruction {
	if idx, ok := p.index[instr]; ok && idx+1 < len(p.instr) {
		return p.instr[idx+1]
	}
	return nil
}

// Get previos Instruction after provided one
func (p *Program) Prev(instr *Instruction) *Instruction {
	if idx, ok := p.index[instr]; ok && idx-1 >= 0 {
		return p.instr[idx-1]
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
