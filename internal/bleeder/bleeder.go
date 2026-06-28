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
	vibes map[string]*Vibe
	sqncs map[string]*Sequence
}

// Create new Bleeder instance
func NewBleeder(bleed *Bleed) *Bleeder {
	// logs.TraceFrom(logs.INFO, "called")
	b := &Bleeder{
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
	return b.GenSeqIR(MAIN_NAME, "")
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
	var prevCh string           // previous operation character
	var prevIns *ir.Instruction // previous instruction
	var prevLinkName string     // previous link name
	var prevLinkArgs []string   // previous link args
	outIrp := ir.NewProgram()   // generated IR Program

	for _, line := range tokens {
		for _, cell := range line {
			ch := string(cell[0])
			args := splitArgs(cell[1:])
			switch ch {
			/* PLAY */
			case chPlay:
				cT += aT
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
				prevCh = ch
				aT = ins.Dur

			/* LINK */
			case chLink:
				cT += aT
				prevLinkName = getArg(args, 0, "")
				prevLinkArgs = args[1:]
				irp, err := b.evalLink(prevLinkName, prevLinkArgs)
				if err != nil {
					return nil, fmt.Errorf("%w in %s", err, cell)
				}
				irp.Shift(cT)
				outIrp.Merge(irp)
				prevCh = ch
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
					prevIns = ins
					aT = ins.Dur
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
	var cT, iT float64          // current time, initial time
	var prevCh string           // previous operation character
	var prevIns *ir.Instruction // previous instruction
	var prevLinkName string     // previous link name
	var prevLinkArgs []string   // previous link args
	outIrp := ir.NewProgram()   // generated IR Program

	for _, line := range tokens {
		cT = iT
		if cT == 0 {
			cT = 0
			prevCh = ""
			prevIns = nil
			prevLinkName = ""
			prevLinkArgs = nil
		}
		for ci, cell := range line {
			ch := string(cell[0])
			switch ch {
			/* FILL */
			case chPlay:
				if prevIns == nil {
					return nil, fmt.Errorf("%s without previous instruction", ch)
				}
				aT := evalArg(getArg(splitArgs(cell[1:]), 0, "1"))
				prevIns.Dur += aT
				prevCh = ch
				cT += aT

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
				aT := evalArg(getArg(splitArgs(cell[1:]), 0, "1"))
				cT += aT

			/* WITH */
			case chWith:
				if ci != len(line)-1 {
					return nil, fmt.Errorf("%s should be the last operation", ch)
				}
				continue

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

		if len(line) > 0 && line[len(line)-1] == chWith {
			iT = cT
		} else {
			iT = 0
		}
	}
	outIrp.Sort()
	return outIrp, nil
}

// evaluate args and produce play instruction
func (b *Bleeder) evalPlay(midiArg, durArg, volArg string) (*ir.Instruction, error) {
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
