package cmd

import (
	"bleeder/internal/ir"
	"bleeder/internal/utils"
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
	// get main IR from cache or build it
	return b.GetSeqIR(b.main, nil)
}

// Get IR of specified section with args
func (b *Bleeder) GetSeqIR(name string, args []string) (*ir.Program, error) {
	// try IR from cache
	key := name + ":" + strings.Join(args, ",")
	if IR, ok := b.irs[key]; ok {
		return IR, nil
	}
	// get sequence from bleed
	seq, ok := b.bleed.Sequences[name]
	if !ok {
		return nil, fmt.Errorf("Sequence is not found: %s", name)
	}
	// expands arguments
	content := utils.ReplaceByMap(seq.Args, seq.Content...)
	return b.GetRawIR(content)
}

// Get IR of raw DSL
func (b *Bleeder) GetRawIR(lines []string) (*ir.Program, error) {
	IR := ir.NewProgram()
	for _, line := range lines {
		var instr *ir.Instruction
		for token := range strings.SplitSeq(line, " ") {
			switch token {
			case b.cfg.Mapping.Play:
				instr = &ir.Instruction{}
			case b.cfg.Mapping.Wait:
			case b.cfg.Mapping.Repeat:
			case b.cfg.Mapping.RepeatLine:
			default:
				IR.Add(instr)
				// TODO: fill up current instruction
				// return nil, fmt.Errorf("Unknown instruction: %v", instr)
			}
		}

		// in case or @seq_reference // 2
		// IR2 := GetSeqIR(seq_reference, args)
		// IR.Merge(IR2)

	}
	// by good design this one should be main workhorse
	return IR, nil
}
