package cmd

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared"
	"bleeder/internal/shared/logs"
	"fmt"
	"strings"
)

// Used for parsing hack to split content by operations
const REPLACER_CHAR = "\\"

// Core DSL processor and IRs generator
type Bleeder struct {
	bleed    *Bleed                 // reference to Bleed
	cfg      *Config                // reference to Config
	main     string                 // main sequence name
	programs map[string]*ir.Program // cached IR programs
	replacer *strings.Replacer      // cached string replacer
}

// Create new Bleeder instance
func NewBleeder(cfg *Config) *Bleeder {
	return &Bleeder{
		cfg:      cfg,
		programs: make(map[string]*ir.Program),
		replacer: strings.NewReplacer(
			cfg.Mapping.Play, REPLACER_CHAR+cfg.Mapping.Play,
			cfg.Mapping.Wave, REPLACER_CHAR+cfg.Mapping.Wave,
			cfg.Mapping.Seq, REPLACER_CHAR+cfg.Mapping.Seq,
			cfg.Mapping.Wait, REPLACER_CHAR+cfg.Mapping.Wait,
			cfg.Mapping.RepeatLine, REPLACER_CHAR+cfg.Mapping.RepeatLine,
			cfg.Mapping.Repeat, REPLACER_CHAR+cfg.Mapping.Repeat,
		),
	}
}

// Load bleed sequences
func (b *Bleeder) Bleed(bleed *Bleed) (*Bleeder, error) {
	logs.Debug("Bleeder Bleed")
	// parse included bleeds
	if bleed.Meta.Include != nil {
		// TODO: will be fixed during parser rewrite
		return nil, fmt.Errorf("not implemented yet")
		// for _, v := range bleed.Meta.Include {
		// 	_, err := b.Bleed(v.Bleed)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
	}
	// parse main section to cache sequences
	b.bleed = bleed
	b.main = bleed.Meta.Main
	_, err := b.GenMainIR()
	return b, err // nil, err // b, nil
}

// Get IR of the main sequence
func (b *Bleeder) GenMainIR() (*ir.Program, error) {
	logs.Debug("Bleeder GenMainIR")
	// get main IR from cache or build it
	return b.GenSeqIR(b.main, nil, 0)
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, args []string, t float64) (*ir.Program, error) {
	logs.Debug("Bleeder GenSeqIR %s, %v", name, args)
	// try IR from cache
	key := name
	if len(args) > 0 {
		key += ":" + strings.Join(args, ",")
	}
	fmt.Printf("LOOKING FOR %s\n", key)
	if irp, ok := b.programs[key]; ok {
		return irp, nil
	}
	// get sequence from bleed
	seq, ok := b.bleed.Sequences[name]
	if !ok {
		return nil, fmt.Errorf("sequence is not found: %s", name)
	}
	// expands arguments
	pairs := make([]string, len(seq.Args))
	copy(pairs, seq.Args)
	for i, arg := range args {
		pairs[i*2+1] = arg
	}
	content := strings.NewReplacer(pairs...).Replace(seq.Content)
	irp, t, err := b.GenRawIR(content, t)
	if err != nil {
		return nil, err
	}
	for i := 1; i < seq.Repeat; i++ {
		pr2, t2, err := b.GenRawIR(content, t)
		if err != nil {
			return nil, err
		}
		t += t2
		irp.Merge(pr2)
	}
	// cache generated IR
	b.programs[key] = irp
	return irp, nil
}

// Get IR of raw DSL
func (b *Bleeder) GenRawIR(content string, t float64) (*ir.Program, float64, error) {
	logs.Debug("Bleeder GenRawIR\n%s", content)
	lines := strings.Split(content, "\n")
	defDur := b.cfg.Parser.DefaultDur
	defVol := b.cfg.Parser.DefaultVol
	irp := ir.NewProgram()
	accDelay := 0.0

	for _, line := range lines {
		// skip empty lines
		if len(line) == 0 {
			continue
		}
		// split by instruction characters
		ins := &ir.Instruction{Info: "Start"}
		raw := strings.Split(b.replacer.Replace(line), REPLACER_CHAR)[1:]
		for _, r := range raw {
			v := strings.Fields(r)
			switch v[0] {
			// parse PLAY >
			case b.cfg.Mapping.Play:
				accDelay = 0
				ins = &ir.Instruction{
					Freq: parseNoteArg(v, 1, "c4"),
					Dur:  int(parseFloatArg(v, 2, defDur)),
					Vol:  parseFloatArg(v, 3, defVol),
					Time: int(t),
					Info: r, // Just for debug
				}
				irp.Add(ins)
			// parse WAVE ~
			case b.cfg.Mapping.Wave:
				accDelay = 0
				ins = &ir.Instruction{
					Freq: parseFloatArg(v, 1, 440),
					Dur:  int(parseFloatArg(v, 2, defDur)),
					Vol:  parseFloatArg(v, 3, defVol),
					Time: int(t),
					Info: r, // Just for debug
				}
				irp.Add(ins)
			// parse SEQ
			case b.cfg.Mapping.Seq:
				accDelay = 0
				args := v[2:]
				pr2, err := b.GenSeqIR(v[1], args, t)
				if err != nil {
					return nil, t, err
				}
				irp.Merge(pr2)
			// parse WAIT
			case b.cfg.Mapping.Wait:
				w := parseFloatArg(v, 1, float64(ins.Dur))
				t += w
				accDelay = +w
			// parse REPEAT
			case b.cfg.Mapping.Repeat:
				_, mod := parseInstructionArg(v, 1, "")
				ins = &ir.Instruction{
					Freq: audio.TransposeFreq(ins.Freq, mod),
					Dur:  ins.Dur, // TODO
					Vol:  ins.Vol,
					Time: int(t),
					Info: "REPEAT " + r,
				}
				t += accDelay
				irp.Add(ins)
			// parse REPEAT LINE
			case b.cfg.Mapping.RepeatLine:
				return nil, t, fmt.Errorf("not implemented yet: %s", b.cfg.Mapping.RepeatLine)
			default:
				fmt.Printf("Unknown instruction: %s\n", v[0])
			}
		}
	}
	return irp, t, nil
}

// private helpers

func parseInstructionArg(v []string, idx int, def string) (arg string, mod float64) {
	if idx >= len(v) {
		return def, 0.0
	}
	i := strings.IndexAny(v[idx], "+-")
	if i == -1 {
		return v[idx], 0.0
	}
	return v[idx][:i], shared.Str2Float(v[idx][i:], 0.0)
}

func parseNoteArg(v []string, idx int, def string) float64 {
	s, mod := parseInstructionArg(v, idx, def)
	i := audio.NoteToMidi(s) + int(mod)
	return audio.MidiToFreq(i)
}

func parseFloatArg(v []string, idx int, def float64) float64 {
	s, mod := parseInstructionArg(v, idx, shared.Float2Str(def))
	return shared.Str2Float(s, 0.0) + mod
}
