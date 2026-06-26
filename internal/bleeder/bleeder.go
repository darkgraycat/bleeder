package bleeder

import (
	"bleeder/internal/ir"
	// "bleeder/internal/shared/logs"
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
	// logs.TraceFrom(logs.INFO, "called")
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
	// logs.TraceFrom(logs.INFO, "called")
	return b.GenSeqIR(b.main, "")
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, vars string) (*ir.Program, error) {
	// logs.TraceFrom(logs.INFO, "called with %v, %v", name, vars)
	irp := b.cache.Get(name, vars)
	if irp != nil {
		return irp, nil
	}
	seq, ok := b.sqncs[name]
	if !ok {
		return nil, fmt.Errorf("%s sequence does not exist", name)
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
		return nil, fmt.Errorf("%s is unknown sequence type", name)
	}
	if err != nil {
		return nil, err
	}

	b.cache.Set(name, vars, irp)
	return irp, nil
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(tokens [][]string) (*ir.Program, error) {
	var prev string             // previos operation character
	var prevIns *ir.Instruction // previos instruction
	var prevLinkName string     // previos link name
	var prevLinkArgs []string   // previos link args

	concated := slices.Concat(tokens...) // TODO: think about passing it directry into loop
	outIrp := ir.NewProgram()            // generated IR Program
	cT, aT := 0.0, 0.0                   // current time, advance time

	// TODO: maybe its even better to do for-for loop - better error messages btw(line number)
	for _, cell := range concated {
		ch := string(cell[0])
		args := splitArgs(cell[1:])
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
				return nil, fmt.Errorf("%w in %s", err, cell)
			}
			ins.Time, ins.Info = cT, cell
			outIrp.Add(ins)
			aT = ins.Dur
			prevIns = ins

		case chLink:
			cT += aT
			prev = ch
			prevLinkName = getArg(args, 0, "")
			prevLinkArgs = args[1:]
			irp, err := b.evalLink(prevLinkName, prevLinkArgs)
			if err != nil {
				return nil, fmt.Errorf("%w in %s", err, cell)
			}
			irp.Shift(cT)
			outIrp.Merge(irp)
			aT = irp.Duration()

		case chPrev:
			cT += aT
			switch prev {
			case chPlay:
				ins, err := b.evalPlay(
					getArg(args, 0, strconv.FormatFloat(prevIns.Midi, 'g', 8, 64)),
					getArg(args, 1, strconv.FormatFloat(prevIns.Dur, 'g', 8, 64)),
					getArg(args, 2, strconv.FormatFloat(prevIns.Vol, 'g', 8, 64)),
				)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				ins.Time, ins.Info = cT, cell
				outIrp.Add(ins)
				aT = ins.Dur
				prevIns = ins
			case chLink:
				newArgs := make([]string, max(len(prevLinkArgs), len(args)))
				for i := range newArgs {
					newArgs[i] = getArg(args, i, getArg(prevLinkArgs, i, ""))
				}
				prevLinkArgs = newArgs
				irp, err := b.evalLink(prevLinkName, newArgs)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				irp.Shift(cT)
				outIrp.Merge(irp)
				aT = irp.Duration()
			default:
				return nil, fmt.Errorf("%s can't be used after `%s`", chPrev, prev)
			}

		case chVibe:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)

		case chRest:
			cT += aT
			aT = evalArg(getArg(args, 0, "1"))

		case chWith:
			aT = 0
		}
	}
	outIrp.Sort()
	return outIrp, nil
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	var prev string             // previos operation character
	var prevIns *ir.Instruction // previos instruction
	// var prevLinkName string     // previos link name
	// var prevLinkArgs []string   // previos link args
	outIrp := ir.NewProgram()
	for _, line := range tokens {
		fmt.Printf("LINE: %s\n", line)
		cT, aT := 0.0, 0.0 // current time, advance time
		for _, cell := range line {
			fmt.Printf("CELL: %s\n", cell)
			ch := string(cell[0])
			switch ch {
			case chPlay:
				switch prev {
				case "":
					prevIns.Dur += 1.0
				case chLink:
					return nil, fmt.Errorf("%s not after %s implemented yet", ch, chLink)
				}
			case chLink:
				return nil, fmt.Errorf("%s not implemented yet", ch)
			case chPrev:
				switch prev {
				case "":
					cT += aT
					args := splitArgs(cell[1:])
					ins, err := b.evalPlay(
						getArg(args, 0, strconv.FormatFloat(prevIns.Midi, 'g', 8, 64)),
						getArg(args, 1, strconv.FormatFloat(prevIns.Dur, 'g', 8, 64)),
						getArg(args, 2, strconv.FormatFloat(prevIns.Vol, 'g', 8, 64)),
					)
					if err != nil {
						return nil, fmt.Errorf("%w in %s", err, cell)
					}
					ins.Time, ins.Info = cT, cell
					outIrp.Add(ins)
					aT = 1
					prevIns = ins
				case chLink:
					return nil, fmt.Errorf("%s not implemented yet for @", ch)
				}
			case chVibe:
				return nil, fmt.Errorf("%s not implemented yet", ch)
			case chRest:
				cT += aT
				aT = 1 // TODO: do we need it?
			default:
				cT += aT
				prev = ""
				args := splitArgs(cell)
				ins, err := b.evalPlay(
					getArg(args, 0, ""),
					getArg(args, 1, "1"),
					getArg(args, 2, "1"),
				)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				ins.Time, ins.Info = cT, cell
				outIrp.Add(ins)
				aT = 1
				prevIns = ins
			}
		}
	}
	fmt.Printf("%v %v\n", prev, prevIns)
	outIrp.Sort()
	return outIrp, nil
}

// evaluate args and produce play instruction
func (b *Bleeder) evalPlay(midiArg, durArg, volArg string) (*ir.Instruction, error) {
	// fmt.Printf("evaluating m `%s`, d `%s`, v `%s`\n", midiArg, durArg, volArg)
	midi := evalArg(midiArg)
	dur := evalArg(durArg)
	vol := evalArg(volArg)
	if math.IsNaN(midi + dur + vol) {
		return nil, fmt.Errorf("eval >%.1f:%.1f:%.1f failed", midi, dur, vol)
	}
	return &ir.Instruction{Midi: midi, Dur: dur, Vol: vol}, nil
}

// evaluate args and produce linked program
func (b *Bleeder) evalLink(name string, args []string) (*ir.Program, error) {
	if name == "" {
		return nil, fmt.Errorf("eval with no sequence name failed")
	}
	irp, err := b.GenSeqIR(name, strings.Join(args, chArgs))
	if err != nil {
		return nil, err
	}
	return irp.Copy(), nil
}
