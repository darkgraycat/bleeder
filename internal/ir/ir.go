package ir

type Program struct {
	ctx          Context
	instructions []*Instruction
}

func NewProgram(ctx Context) *Program {
	return &Program{
		ctx:          ctx,
		instructions: make([]*Instruction, 0),
	}
}

func (p *Program) Merge(src *Program) {
	p.instructions = append(p.instructions, src.instructions...)
}

func (p *Program) Add(i *Instruction) {
	p.instructions = append(p.instructions, i)
}

type Context struct {
	SampleRate int
}

type Instruction struct {
	Tag  string
	Note int
	Freq int
	Dur  float32
	Vol  float32
}

