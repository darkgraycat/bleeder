package bleeder

import (
	"bleeder/internal/ir"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	main  string
	vibes map[string]*Vibe
	sqncs map[string]*Sequence
	cache *Cache[*ir.Program]
}

// Create new Bleeder instance
func NewBleeder(bleed *Bleed) *Bleeder {
	// logs.Trace(logs.INFO, "called")
	b := &Bleeder{
		main:  bleed.Meta.Main,
		cache: NewCache[*ir.Program](),
		vibes: make(map[string]*Vibe, len(bleed.Vibes)),
		sqncs: make(map[string]*Sequence, len(bleed.Lanes)+len(bleed.Riffs)),
	}
	for k, v := range bleed.Vibes {
		b.vibes[k] = &v
	}
	for k, v := range bleed.Lanes {
		b.sqncs[k] = &v
	}
	for k, v := range bleed.Riffs {
		b.sqncs[k] = &v
	}
	return b
}

// Get IR of the main sequence
func (b *Bleeder) GenMainIR() (*ir.Program, error) {
	// logs.Trace(logs.INFO, "called")
	return b.GenSeqIR(b.main, "")
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, vars string) (*ir.Program, error) {
	// logs.Trace(logs.INFO, "called with %v, %v", name, vars)
	irp := b.cache.Get(name, vars)
	if irp != nil {
		return irp, nil
	}
	seq, ok := b.sqncs[name]
	if !ok {
		return nil, fmt.Errorf("sequence does not exist: %s", name)
	}
	varsMap := parseVars(seq.Vars, splitArgs(vars))
	rawContent := applyVars(seq.Content, varsMap)

	tokens := tokenizeContent(rawContent)
	if tokens == nil {
		return nil, fmt.Errorf("sequence content is invalid or empty")
	}

	var err error
	switch seq.Type {
	case SEQ_LANE:
		irp, err = b.genLaneIR(tokens)
	case SEQ_RIFF:
		irp, err = b.genRiffIR(tokens)
	default:
		return nil, fmt.Errorf("unknown sequence type: %s", name)
	}
	if err != nil {
		return nil, err
	}

	b.cache.Set(name, vars, irp)
	return irp, nil
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(tokens [][]string) (*ir.Program, error) {
	// logs.Trace(logs.INFO, "called with %v", tokens)
	concated := slices.Concat(tokens...)
	seqIrp := ir.NewProgram()
	cT, aT := 0, 0              // current time, advance time
	var prev string             // previos operation character
	var prevIns *ir.Instruction // previos instruction
	var prevLinkName string     // previos link name
	var prevLinkArgs []string   // previos link args
	// logs.Debug("Lane tokens %s\n", concated)

	for _, raw := range concated {
		ch := string(raw[0])
		args := splitArgs(raw[1:])
		switch ch {
		case chPlay:
			cT += aT
			prev = ch
			ins, err := b.evalPlay(
				getArg(args, 0, ""),
				getArg(args, 1, "1"),
				getArg(args, 2, "1"),
			)
			if err != nil {
				return nil, fmt.Errorf("parsing error %v for %s", err, raw)
			}
			ins.Time, ins.Info = cT, raw
			seqIrp.Add(ins)
			aT = ins.Dur
			prevIns = ins

		case chLink:
			cT += aT
			prev = ch
			prevLinkName = getArg(args, 0, "")
			prevLinkArgs = args[1:]
			irp, err := b.evalLink(prevLinkName, prevLinkArgs)
			if err != nil {
				return nil, fmt.Errorf("parsing error %v for %s", err, raw)
			}
			irp.Shift(cT)
			seqIrp.Merge(irp)
			aT = irp.Duration()

		case chPrev:
			cT += aT
			switch prev {
			case chPlay:
				ins, err := b.evalPlay(
					getArg(args, 0, strconv.FormatFloat(prevIns.Midi, 'g', 8, 64)),
					getArg(args, 1, strconv.FormatInt(int64(prevIns.Dur), 10)),
					getArg(args, 2, strconv.FormatFloat(prevIns.Vol, 'g', 8, 64)),
				)
				if err != nil {
					return nil, fmt.Errorf("parsing error %v for %s", err, raw)
				}
				ins.Time, ins.Info = cT, raw
				seqIrp.Add(ins)
				aT = ins.Dur
				prevIns = ins
			case chLink:
				newArgs := make([]string, max(len(prevLinkArgs), len(args)))
				for i := range newArgs {
					newArgs[i] = getArg(args, i, getArg(prevLinkArgs, i, ""))
				}
				irp, err := b.evalLink(prevLinkName, newArgs)
				if err != nil {
					return nil, fmt.Errorf("parsing error %v for %s", err, raw)
				}
				irp.Shift(cT)
				seqIrp.Merge(irp)
				aT = irp.Duration()
			default:
				return nil, fmt.Errorf("%s can't be used after `%s`", chPrev, prev)
			}

		case chVibe:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)

		case chRest:
			cT += aT
			aT = int(evalArg(getArg(args, 0, "1")))

		case chWith:
			aT = 0
		}
	}
	seqIrp.Sort()
	return seqIrp, nil
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	// logs.Trace(logs.INFO, "called with\n%v", tokens)
	// logs.Debug("Lane tokens %s\n", tokens)
	return nil, fmt.Errorf("sequence riff type not implemented yet")
}

// evaluate args and produce play instruction
func (b *Bleeder) evalPlay(midiArg, durArg, volArg string) (*ir.Instruction, error) {
	midi := evalArg(midiArg)
	dur := evalArg(durArg)
	vol := evalArg(volArg)
	if math.IsNaN(midi + dur + vol) {
		return nil, fmt.Errorf("cannot eval %s %s %s", midiArg, durArg, volArg)
	}
	return &ir.Instruction{Midi: midi, Dur: int(dur), Vol: vol}, nil
}

// evaluate args and produce linked program
func (b *Bleeder) evalLink(name string, args []string) (*ir.Program, error) {
	if name == "" {
		return nil, fmt.Errorf("cannot eval sequence without a name")
	}
	irp, err := b.GenSeqIR(name, strings.Join(args, ":"))
	if err != nil {
		return nil, err
	}
	return irp.Copy(), nil
}
