package cmd

import (
	"bleeder/internal/ir"
	"fmt"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	irs map[string]*ir.Program
	cfg *Config
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
	// find main sequence
	main, ok := bleed.Sequence[bleed.Meta.Main]
	if !ok {
		return nil, fmt.Errorf("Sequence is not found: %s", bleed.Meta.Main)
	}
	// parse included bleeds
	if bleed.Meta.Bleeds != nil {
		for _, v := range bleed.Meta.Bleeds {
			_, err := b.Bleed(v)
			if err != nil {
				return nil, err
			}
		}
	}
	// parse main section
	_, err := b.GetRawIR(main.Content)
	if err != nil {
		return b, err
	}
	// 2. do I need to cache main?
	// 6. I can store it with literal name "main"
	// 7. But I feel its wrong

	return b, nil
}

// Get combined 
func (b *Bleeder) GetFullIR() (*ir.Program, error) {
	// 1. so how can I get full without original bleed file details
	// 3. then I can just GetSeqIr()
	// 4. but what name?
	// 5. name is also stored in bleed file

	// is just and accessor

	// huh
	// need to exactly know what and where to cache
	// and where we are really need to operate with bleed file

	// by good design this one should call GetSeqIR
	return nil, nil
}

func (b *Bleeder) GetSeqIR(seq string, args []string) (*ir.Program, error) {
	key := fmt.Sprintf("%s:%v", seq, args)
	IR, ok := b.irs[key]
	if ok {
		return IR, nil
	}
	// TODO: mb build it and write into cache?
	// Its needed in case file contains unused sequences

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

// OLD

func (br *Bleeder) ParseBleed(bleed *Bleed) (*ir.Program, error) {
	return br.ParseSequence(
		bleed.Meta.Main,
		bleed,
	)
}

func (br *Bleeder) ParseSequence(name string, bleed *Bleed) (*ir.Program, error) {
	seq, ok := bleed.Sequence[name]
	if !ok {
		return nil, fmt.Errorf("No sequence %s is found", name)
	}
	ir := ir.NewProgram(ir.Context{
		SampleRate: br.cfg.Audio.SampleRate,
	})
	for _, line := range seq.Content {
		// how to replace with seq.Args before parsing line?
		lineIr, err := br.ParseLine(line, bleed)
		if err != nil {
			return nil, err
		}
		ir.Merge(lineIr)
	}
	return ir, nil
}

func (br *Bleeder) ParseLine(line string, bleed *Bleed) (*ir.Program, error) {
	// TODO: but what about arguments?
	return nil, nil
}

// func parseBleed(cfg *Config, bleed *Bleed) (*ir.Program, error) {
// 	ir := ir.NewProgram(ir.Context{})
// 	included, err := parseInclude(bleed)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// TODO: parse included sequences into ir
// 	for _, b := range included {
// 		fmt.Printf("B.Seq %v", b.Sequence)
// 		for seq, v := range b.Sequence {
// 			if _, ok := bleed.Sequence[seq]; !ok {
// 				bleed.Sequence[seq] = v
// 			}
// 		}
// 	}
//
// 	PrintBleed(bleed)
// 	return ir, nil
// }

// func parseInclude(bleed *Bleed) ([]*Bleed, error) {
// 	bleeds := make([]*Bleed, len(bleed.Include))
// 	for _, b := range bleed.Include {
// 		newBleed, err := LoadBleed(b.Path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		bleeds = append(bleeds, newBleed)
// 	}
// 	return bleeds, nil
// }

func PrintBleed(bleed *Bleed) {
	for seq, d := range bleed.Sequence {
		fmt.Printf("Seq %s\n", seq)
		fmt.Printf("\tArgs %v\n", d.Args)
		fmt.Printf("\tRepeat %v\n", d.Repeat)
		for n, line := range d.Content {
			fmt.Printf("\t%d: %v\n", n, line)
		}
	}
}
