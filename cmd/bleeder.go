package cmd

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared"
	"fmt"
	"strconv"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	irs   map[string]*ir.Program
	cfg   *Config
	bleed *Bleed
	main  string
	r     *strings.Replacer
}

// Create new Bleeder instance
func NewBleeder(cfg *Config) *Bleeder {
	return &Bleeder{
		irs: make(map[string]*ir.Program),
		cfg: cfg,
		r: strings.NewReplacer(
			cfg.Mapping.Play, "\\"+cfg.Mapping.Play,
			cfg.Mapping.Wait, "\\"+cfg.Mapping.Wait,
			cfg.Mapping.Wave, "\\"+cfg.Mapping.Wave,
			cfg.Mapping.Repeat, "\\"+cfg.Mapping.Repeat,
			cfg.Mapping.RepeatLine, "\\"+cfg.Mapping.RepeatLine,
			cfg.Mapping.Debug, "\\"+cfg.Mapping.Debug,
		),
	}
}

// Load bleed sequences
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
	fmt.Printf("FN CALL GetMainIR()\n")
	// get main IR from cache or build it
	return b.GetSeqIR(b.main, nil)
}

// Get IR of specified section with args
func (b *Bleeder) GetSeqIR(name string, args []string) (*ir.Program, error) {
	fmt.Printf("FN CALL GetSeqIR(%s)\n", name)
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
	content := shared.ReplaceByMap(seq.Args, seq.Content...)
	return b.GetRawIR(content)
}

// Get IR of raw DSL
func (b *Bleeder) GetRawIR(lines []string) (*ir.Program, error) {
	fmt.Printf("FN CALL GetRawIR(%s)\n", lines)
	IR := ir.NewProgram()
	ad := 0.0 // accumulated delay
	ld := 0.0 // last delay
	for _, line := range lines {
		// skip empty lines
		if len(line) == 0 {
			continue
		}
		// fmt.Printf("-- LINE %s\n", line)
		// split by instruction characters
		var inst *ir.Instruction
		raw := strings.Split(b.r.Replace(line), "\\")[1:]
		for _, r := range raw {
			v := strings.Fields(r)
			// fmt.Printf("-- RAW  %s\n", v)
			switch v[0] {
			// parse PLAY >
			case b.cfg.Mapping.Play:
				inst = &ir.Instruction{
					// TODO: convert note into freq
					Freq: parseArg(v, 1, 0, 440.0),
					Dur:  parseArg(v, 2, 0, 1.0),
					Vol:  parseArg(v, 3, 0, 1.0),
					Time: ad,
					Info: r, // Just for debug
				}
				IR.Add(inst)
			// parse WAVE ~
			case b.cfg.Mapping.Wave:
				inst = &ir.Instruction{
					Freq: parseArg(v, 1, 0, 440.0),
					Dur:  parseArg(v, 2, 0, 1.0),
					Vol:  parseArg(v, 3, 0, 1.0),
					Time: ad,
					Info: r, // Just for debug
				}
				IR.Add(inst)
			// parse WAIT
			case b.cfg.Mapping.Wait:
				ld = parseArg(v, 1, 0, 1.0)
				ad += ld
			// parse REPEAT
			case b.cfg.Mapping.Repeat:
				inst = &ir.Instruction{
					Freq: inst.Freq,
					Dur:  inst.Dur,
					Vol:  inst.Vol,
					Time: ad,
					Info: "REPEAT " + inst.Info,
				}
				ad += ld
				ld = 0.0
				IR.Add(inst)
			// parse REPEAT LINE
			case b.cfg.Mapping.RepeatLine:
				// TODO
			default:
				fmt.Printf("Unknown instruction: %s\n", v[0])
			}
		}

		// in case or @seq_reference // 2
		// IR2 := GetSeqIR(seq_reference, args)
		// IR.Merge(IR2)
	}
	return IR, nil
}

func parseArg(v []string, i int, prev, def float64) float64 {
	if i >= len(v) {
		return prev + def
	}
	f, err := strconv.ParseFloat(v[i], 64)
	if err != nil {
		return prev + def
	}
	return f
}
