package cmd

import (
	"bleeder/internal/ir"
	"fmt"
)

type Bleeder struct {
	cfg *Config
	irs map[string]*ir.Program
}

func NewBleeder(cfg *Config) *Bleeder {
	return &Bleeder{
		cfg: cfg,
		irs: make(map[string]*ir.Program),
	}
}

func (b *Bleeder) Bleed(bleed Bleed) error {

	return nil
}

func (b *Bleeder) IntoIRFull() error {
	return nil
}

func (b *Bleeder) IntoIRSeq() error {
	return nil
}

func (b *Bleeder) IntoIRRaw() error {
	return nil
}

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
