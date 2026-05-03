package ir

// Intermediate Representation Program
type Program struct {
	instructions []*Instruction       // instructions array
	indexesCache map[*Instruction]int // instructions cache
}

// Create new Program instance
func NewProgram() *Program {
	return &Program{
		instructions: make([]*Instruction, 0),
		indexesCache: make(map[*Instruction]int),
	}
}

// Get an array of instructions
func (p *Program) Instructions() []*Instruction {
	return p.instructions
}

// Add next Instruction pointer to the end
func (p *Program) Add(in *Instruction) {
	p.instructions = append(p.instructions, in)
	p.indexesCache[in] = len(p.instructions) - 1
}

// Cut Instructions into new Program
func (p *Program) Cut(start, end int) *Program {
	sliced := p.instructions[start:end]
	np := NewProgram()
	for i, in := range sliced {
		np.instructions = append(np.instructions, in)
		np.indexesCache[in] = i
	}
	return np
}

// Merge another Program into current one
func (p *Program) Merge(src *Program) {
	offset := len(p.instructions)
	p.instructions = append(p.instructions, src.instructions...)
	for i, in := range src.instructions {
		p.indexesCache[in] = offset + i
	}
}

// Get the number of Instructions in Program
func (p *Program) Length() int {
	return len(p.instructions)
}

// Get duration of whole Program
func (p *Program) Duration() int {
	dur := 0
	for _, in := range p.instructions {
		end := in.Time + in.Dur
		if end > dur {
			dur = end
		}
	}
	return dur
}

// Get first Instruction
func (p *Program) First() *Instruction {
	if n := len(p.instructions); n > 0 {
		return p.instructions[0]
	}
	return nil
}

// Get last Instruction
func (p *Program) Last() *Instruction {
	if n := len(p.instructions); n > 0 {
		return p.instructions[n-1]
	}
	return nil
}

// Get next Instruction after provided one
func (p *Program) Next(in *Instruction) *Instruction {
	if idx, ok := p.indexesCache[in]; ok && idx+1 < len(p.instructions) {
		return p.instructions[idx+1]
	}
	return nil
}

// Get previos Instruction after provided one
func (p *Program) Prev(in *Instruction) *Instruction {
	if idx, ok := p.indexesCache[in]; ok && idx-1 >= 0 {
		return p.instructions[idx-1]
	}
	return nil
}

// Instruction is a basic unit of Intermediate Representation
type Instruction struct {
	Freq float64 // frequence
	Vol  float64 // volume
	Dur  int     // duration
	Time int     // start time
	Info string  // additional information
}

// TODO: make Dur and Time ints
