package cmd

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared"
	"fmt"
	"strings"
)

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
			cfg.Mapping.Play, "\\"+cfg.Mapping.Play,
			cfg.Mapping.Wave, "\\"+cfg.Mapping.Wave,
			cfg.Mapping.Seq, "\\"+cfg.Mapping.Seq,
			cfg.Mapping.Wait, "\\"+cfg.Mapping.Wait,
			cfg.Mapping.RepeatLine, "\\"+cfg.Mapping.RepeatLine,
			cfg.Mapping.Repeat, "\\"+cfg.Mapping.Repeat,
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
	fmt.Printf("CALL GetMainIR\n")
	// get main IR from cache or build it
	return b.GetSeqIR(b.main, nil, 0)
}

// Get IR of specified section with args
func (b *Bleeder) GetSeqIR(name string, args []string, t float64) (*ir.Program, error) {
	fmt.Printf("CALL GetSeqIR %s, %v\n", name, args)
	// try IR from cache
	key := name + ":" + strings.Join(args, ",")
	if pr, ok := b.programs[key]; ok {
		return pr, nil
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
	return b.GetRawIR(content, t)
}

// Get IR of raw DSL
func (b *Bleeder) GetRawIR(content string, t float64) (*ir.Program, error) {
	fmt.Printf("CALL GetRawIR\n%s\n", content)
	lines := strings.Split(content, "\n")
	defDur := b.cfg.Parser.DefaultDur
	defVol := b.cfg.Parser.DefaultVol
	pr := ir.NewProgram()
	rt := 0.0 // relative time
	lastRt := 0.0
	for _, line := range lines {
		// skip empty lines
		if len(line) == 0 {
			continue
		}
		// fmt.Printf("-- LINE %s\n", line)
		// split by instruction characters
		var in *ir.Instruction
		raw := strings.Split(b.replacer.Replace(line), "\\")[1:]
		for _, r := range raw {
			v := strings.Fields(r)
			// fmt.Printf("-- RAW  %s\n", v)
			switch v[0] {
			// parse PLAY >
			case b.cfg.Mapping.Play:
				in = &ir.Instruction{
					Freq: parseNoteArg(v, 1, "c4"),
					Dur:  parseFloatArg(v, 2, defDur),
					Vol:  parseFloatArg(v, 3, defVol),
					Time: t,
					Info: r, // Just for debug
				}
				pr.Add(in)
			// parse WAVE ~
			case b.cfg.Mapping.Wave:
				in = &ir.Instruction{
					Freq: parseFloatArg(v, 1, 440),
					Dur:  parseFloatArg(v, 2, defDur),
					Vol:  parseFloatArg(v, 3, defVol),
					Time: t,
					Info: r, // Just for debug
				}
				pr.Add(in)
			// parse SEQ
			case b.cfg.Mapping.Seq:
				args := v[2:]
				pr2, err := b.GetSeqIR(v[1], args, t)
				if err != nil {
					return nil, err
				}
				pr.Merge(pr2)
			// parse WAIT
			case b.cfg.Mapping.Wait:
				w := parseFloatArg(v, 1, defDur)
				t += w
				rt += w
				lastRt = rt
			// parse REPEAT
			case b.cfg.Mapping.Repeat:
				_, mod := parseInstructionArg(v, 1, "")
				in = &ir.Instruction{
					Freq: audio.TransposeFreq(in.Freq, mod),
					Dur:  in.Dur, // TODO
					Vol:  in.Vol,
					Time: t,
					Info: "REPEAT " + in.Info,
				}
				pr.Add(in)
				t += lastRt
				rt = 0
			// parse REPEAT LINE
			case b.cfg.Mapping.RepeatLine:
				// TODO
			default:
				fmt.Printf("Unknown instruction: %s\n", v[0])
			}
		}
	}
	return pr, nil
}

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
	fmt.Printf("DEBUG S %s, F %f\n", s, mod)
	i := audio.GetNoteIndex(s) + int(mod)
	return audio.FreqByNoteIndex(i)
}

func parseFloatArg(v []string, idx int, def float64) float64 {
	s, mod := parseInstructionArg(v, idx, shared.Float2Str(def))
	return shared.Str2Float(s, 0.0) + mod
}
