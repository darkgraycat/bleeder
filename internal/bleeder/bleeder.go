package bleeder

import (
	"bleeder/internal/ir"
	// "bleeder/internal/shared/logs"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	main  string
	vibes map[string]*Vibe
	sqncs map[string]*Sequence
}

// Create new Bleeder instance
func NewBleeder(bleed *Bleed) *Bleeder {
	// logs.TraceFrom(logs.INFO, "called")
	b := &Bleeder{
		main:  bleed.Meta.Main,
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

	switch seq.Type {
	case SEQ_LANE:
		return b.genLaneIR(tokens)
	case SEQ_RIFF:
		return b.genRiffIR(tokens)
	default:
		return nil, fmt.Errorf("%s is unknown sequence type", name)
	}
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(tokens [][]string) (*ir.Program, error) {
	var cT, aT float64          // current time, advance time
	var prevCh string           // previos operation character
	var prevIns *ir.Instruction // previos instruction
	var prevLinkName string     // previos link name
	var prevLinkArgs []string   // previos link args
	outIrp := ir.NewProgram()   // generated IR Program

	for _, line := range tokens {
		for _, cell := range line {
			ch := string(cell[0])
			args := splitArgs(cell[1:])
			switch ch {
			/* PLAY */
			case chPlay:
				cT += aT
				prevCh = ch
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

			/* LINK */
			case chLink:
				cT += aT
				prevCh = ch
				prevLinkName = getArg(args, 0, "")
				prevLinkArgs = args[1:]
				irp, err := b.evalLink(prevLinkName, prevLinkArgs)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				irp.Shift(cT)
				outIrp.Merge(irp)
				aT = irp.Duration()

			/* PREV */
			case chPrev:
				cT += aT
				switch prevCh {
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
				}

			/* VIBE */
			case chVibe:
				return nil, fmt.Errorf("%s operator is not implemented yet", ch)

			/* REST */
			case chRest:
				cT += aT
				aT = evalArg(getArg(args, 0, "1"))

			/* WITH */
			case chWith:
				aT = 0
			}
		}
	}
	outIrp.Sort()
	return outIrp, nil
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	outIrp := ir.NewProgram() // generated IR Program

	for _, line := range tokens {
		// TODO: remove
		fmt.Printf("LINE: %s\n", line)
		var cT float64              // current time
		var prevCh string           // previos operation character
		var prevIns *ir.Instruction // previos instruction
		var prevLinkName string     // previos link name
		var prevLinkArgs []string   // previos link args

		for _, cell := range line {
			// TODO: remove
			fmt.Printf("CELL: %s\n", cell)
			ch := string(cell[0])
			switch ch {
			/* FILL */
			case chPlay:
				if prevIns == nil {
					return nil, fmt.Errorf("%s without previos instruction", ch)
				}
				prevIns.Dur += 1
				prevCh = ch
				cT += 1

			/* LINK */
			case chLink:
				args := splitArgs(cell[1:])
				prevLinkName = getArg(args, 0, "")
				prevLinkArgs = args[1:]
				irp, err := b.evalLink(prevLinkName, prevLinkArgs)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				irp.Shift(cT)
				outIrp.Merge(irp)
				if last := irp.Last(); last != nil {
					prevIns = last
				}
				prevCh = ch
				cT += irp.Duration()

			/* PREV */
			case chPrev:
				args := splitArgs(cell[1:])
				switch prevCh {
				case chPlay, "":
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
					prevIns = ins
					cT += 1
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
					if last := irp.Last(); last != nil {
						prevIns = last
					}
					cT += irp.Duration()
				}

			/* VIBE */
			case chVibe:
				return nil, fmt.Errorf("%s not implemented yet", ch)

			/* REST */
			case chRest:
				cT += 1

			/* PLAY */
			default:
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
				prevIns = ins
				prevCh = chPlay
				cT += 1
			}
		}
		fmt.Printf("%v %v\n", prevCh, prevIns)
	}
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
