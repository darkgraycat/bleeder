package cmd

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"fmt"
	"strconv"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	bleed    *Bleed                 // reference to Bleed
	cfg      *Config                // reference to Config
	main     string                 // main sequence name
	programs map[string]*ir.Program // set of stored IR programs
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
		return nil, fmt.Errorf("Sequence is not found: %s", name)
	}
	// expands arguments
	content := make([]string, len(seq.Content))
	pairs := make([]string, len(seq.Args))
	copy(pairs, seq.Args)
	for i, arg := range args {
		pairs[i*2+1] = arg
	}
	r := strings.NewReplacer(pairs...)
	for i, line := range seq.Content {
		content[i] = r.Replace(line)
	}
	return b.GetRawIR(content, t)
}

// Get IR of raw DSL
func (b *Bleeder) GetRawIR(lines []string, t float64) (*ir.Program, error) {
	fmt.Printf("CALL GetRawIR\n%s\n", strings.Join(lines, "\n"))
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
					Freq: audio.NoteToFreq(v[1]),
					Dur:  parseFloatArg(v, 2, 0, defDur),
					Vol:  parseFloatArg(v, 3, 0, defVol),
					Time: t,
					Info: r, // Just for debug
				}
				pr.Add(in)
			// parse WAVE ~
			case b.cfg.Mapping.Wave:
				in = &ir.Instruction{
					Freq: parseFloatArg(v, 1, 0, 440),
					Dur:  parseFloatArg(v, 2, 0, defDur),
					Vol:  parseFloatArg(v, 3, 0, defVol),
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
				w := parseFloatArg(v, 1, 0, defDur)
				t += w
				rt += w
				lastRt = rt
			// parse REPEAT
			case b.cfg.Mapping.Repeat:
				in = &ir.Instruction{
					Freq: in.Freq,
					Dur:  in.Dur,
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

func parseFloatArg(v []string, i int, prev, def float64) float64 {
	if i >= len(v) {
		return prev + def
	}
	f, err := strconv.ParseFloat(v[i], 64)
	if err != nil {
		return prev + def
	}
	return f
}
