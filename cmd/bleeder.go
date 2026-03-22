package cmd

import (
	"bleeder/internal/ir"
	"fmt"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	irs   map[string]*ir.Program
	cfg   *Config
	bleed *Bleed
	main  string
}

// Create new Bleeder instance
func NewBleeder(cfg *Config) *Bleeder {
	return &Bleeder{
		irs: make(map[string]*ir.Program),
		cfg: cfg,
	}
}

// Bleed is a main init method
func (b *Bleeder) Bleed(bleed *Bleed) (*Bleeder, error) {
	// parse included bleeds
	if bleed.Meta.Bleeds != nil {
		for _, v := range bleed.Meta.Bleeds {
			_, err := b.Bleed(v.Bleed)
			if err != nil {
				return nil, err
			}
		}
	}
	// parse main section to cache sequences
	b.bleed = bleed
	b.main = bleed.Meta.Main
	_, err := b.GetMainIR()
	return b, err // nil, err // b, nil
}

// Get IR of the main sequence
func (b *Bleeder) GetMainIR() (*ir.Program, error) {
	return b.GetSeqIR(b.main, nil)
}

// Get IR of specified section with args
func (b *Bleeder) GetSeqIR(name string, args []string) (*ir.Program, error) {
	key := fmt.Sprintf("%s:%v", name, args)
	if IR, ok := b.irs[key]; ok {
		return IR, nil
	}
	seq, ok := b.bleed.Sequence[name]
	if !ok {
		return nil, fmt.Errorf("Sequence is not found: %s", name)
	}
	content := seq.Content
	// TODO expand arguments
	IR, err := b.GetRawIR(content)
	if err != nil {
		return nil, err
	}
	b.irs[key] = IR
	return IR, nil
	// content := b.bleed
	// IR := b.GetRawIR() // need access to bleed file

	// TODO: mb build it and write into cache?
	// Its needed in case file contains unused sequences
	// NO. We need to cache ALL sequences because of RawIR

	// ah I need... no.. if I build everything in main
	// like do a preload
	// and methods only runs from cache
}

// Get IR of raw DSL
func (b *Bleeder) GetRawIR(lines []string) (*ir.Program, error) {
	IR := ir.NewProgram()
	for _, line := range lines {
		parts := strings.Split(line, " ")
		// TODO: we need a different way of parsing
		// Because > note+2 1 vol+0.1 | +2 : 1
		// has 3 commands
		instr := parts[0]
		rest := parts[1:]
		fmt.Printf("Parsing %s %v", instr, rest)

		// TODO actual implementation
		switch instr {
		case b.cfg.Mapping.Play:
			IR.Add(&ir.Instruction{Tag: fmt.Sprintf("Play %v", rest)})
		case b.cfg.Mapping.Wait:
			IR.Add(&ir.Instruction{Tag: fmt.Sprintf("Wait %v", rest)})
		case b.cfg.Mapping.Repeat:
			IR.Add(&ir.Instruction{Tag: fmt.Sprintf("Repeat %v", rest)})
		case b.cfg.Mapping.RepeatLine:
			IR.Add(&ir.Instruction{Tag: fmt.Sprintf("RepeatLine %v", rest)})
		default:
			return nil, fmt.Errorf("Unknown instruction: %s", instr)
		}

		// in case or @seq_reference // 2
		// IR2 := GetSeqIR(seq_reference, args)
		// IR.Merge(IR2)

	}
	// by good design this one should be main workhorse
	return IR, nil
}

func printBleed(bleed *Bleed) {
	for seq, d := range bleed.Sequence {
		fmt.Printf("Seq %s\n", seq)
		fmt.Printf("\tArgs %v\n", d.Args)
		fmt.Printf("\tRepeat %v\n", d.Repeat)
		for n, line := range d.Content {
			fmt.Printf("\t%d: %v\n", n, line)
		}
	}
}
