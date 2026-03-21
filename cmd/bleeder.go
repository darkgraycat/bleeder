package cmd

import (
	"bleeder/internal/ir"
	"fmt"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	cfg  *Config
	irs  map[string]*ir.Program
	main string
}

// Create new Bleeder instance
func NewBleeder(cfg *Config) *Bleeder {
	return &Bleeder{
		cfg: cfg,
		irs: make(map[string]*ir.Program),
	}
}

// Bleed is a main init method
func (b *Bleeder) Bleed(bleed *Bleed) (*Bleeder, error) {
	// find main sequence
	b.main = bleed.Meta.Main
	main, ok := bleed.Sequence[b.main]
	if !ok {
		return nil, fmt.Errorf("Sequence is not found: %s", b.main)
	}
	// parse included bleeds
	if bleed.Meta.Bleeds != nil {
		for _, v := range bleed.Meta.Bleeds {
			_, err := b.Bleed(v.Bleed)
			if err != nil {
				return nil, err
			}
		}
	}
	// parse main section
	IR, err := b.GetRawIR(main.Content)
	if err != nil {
		return b, err
	}
	b.irs[b.main] = IR
	return b, nil
}

// Get combined
func (b *Bleeder) GetFullIR() (*ir.Program, error) {
	return b.GetSeqIR(b.main, nil)
}

func (b *Bleeder) GetSeqIR(seq string, args []string) (*ir.Program, error) {
	key := fmt.Sprintf("%s:%v", seq, args)
	if IR, ok := b.irs[key]; ok {
		return IR, nil
	}
	// TODO: mb build it and write into cache?
	// Its needed in case file contains unused sequences

	// ah I need... no.. if I build everything in main
	// like do a preload
	// and methods only runs from cache

	// by good design this one should call GetRawIR
	return nil, fmt.Errorf("Sequence is not found: %s", seq)
}

func (b *Bleeder) GetRawIR(lines []string) (*ir.Program, error) {
	IR := ir.NewProgram(ir.Context{
		SampleRate: b.cfg.Audio.SampleRate,
	})
	for _, line := range lines {
		parts := strings.Split(line, " ")
		instr := parts[0]
		rest := parts[1:]
		fmt.Printf("Parsing %s %v", instr, rest)
		switch instr {
		case b.cfg.Instructions.Play:
			IR.Add(&ir.Instruction{}) // TODO actual implementation
		case b.cfg.Instructions.Wait:
			IR.Add(&ir.Instruction{}) // TODO actual implementation
		case b.cfg.Instructions.Repeat:
			IR.Add(&ir.Instruction{}) // TODO actual implementation
		case b.cfg.Instructions.RepeatLine:
			IR.Add(&ir.Instruction{}) // TODO actual implementation
		default:
			return nil, fmt.Errorf("Unknown instruction: %s", instr)
		}

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
