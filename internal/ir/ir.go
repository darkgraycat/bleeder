package ir

type Program []*Instruction

func NewProgram() *Program {
	p := Program(make([]*Instruction, 0))
	return &p
}

func (p *Program) Add(i *Instruction) {
	*p = append(*p, i)
}

func (p *Program) Merge(src *Program) {
	*p = append(*p, *src...)
}

type Instruction struct {
	Tag  string
	Note int
	Freq int
	Dur  float32
	Vol  float32
}
