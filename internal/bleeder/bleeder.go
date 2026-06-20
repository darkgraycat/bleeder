package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
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
	logs.Trace(logs.INFO, "called")
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
	logs.Trace(logs.INFO, "called")
	return b.GenSeqIR(b.main, "")
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, vars string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with %v, %v", name, vars)
	irp := b.cache.Get(name, vars)
	if irp != nil {
		return irp, nil
	}
	seq, ok := b.sqncs[name]
	if !ok {
		return nil, fmt.Errorf("sequence is not exist: %s", name)
	}
	fmt.Printf("SEQ %s, %s\n", name, vars)
	varsMap := parseVars(seq.Vars, splitArgs(vars))
	rawContent := applyVars(seq.Content, varsMap)
	fmt.Printf("%v\n%v\n", varsMap, rawContent)

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
	// TODO - remove
	for i, ins := range irp.Instructions() {
		fmt.Printf("INS [%d] - %s\n", i, ins)
	}

	b.cache.Set(name, vars, irp)
	return irp, nil
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(tokens [][]string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with %v", tokens)
	concated := slices.Concat(tokens...)
	irp := ir.NewProgram()
	ins := &ir.Instruction{Info: "None"}
	t, ta := 0, 0             // time and additional time
	var prev string           // previos operation character
	var prevLinkName string   // previos link name
	var prevLinkArgs []string // previos link args
	logs.Debug("Lane tokens %s\n", concated)

	for _, raw := range concated {
		ch := string(raw[0])
		args := splitArgs(raw[1:])
		switch ch {
		case chPlay:
			t += ta
			prev = ch
			midi := evalArg(getArg(args, 0, ""))
			dur := evalArg(getArg(args, 1, "1"))
			vol := evalArg(getArg(args, 2, "1"))
			if math.IsNaN(midi + dur + vol) {
				return nil, fmt.Errorf("cannot parse arguments for %s", raw)
			}
			ins = &ir.Instruction{Midi: midi, Dur: int(dur), Vol: vol, Time: t, Info: raw}
			irp.Add(ins)
			ta = int(dur)

		case chLink:
			t += ta
			prev = ch
			prevLinkName = getArg(args, 0, "")
			if prevLinkName == "" {
				return nil, fmt.Errorf("%s requires a sequence name", ch)
			}
			prevLinkArgs = args[1:]
			irpNested, err := b.GenSeqIR(prevLinkName, strings.Join(prevLinkArgs, ":"))
			if err != nil {
				return nil, err
			}
			irpNested = irpNested.Copy()
			irpNested.Shift(t)
			irp.Merge(irpNested)
			ta = irpNested.Duration()

		case chPrev:
			t += ta
			switch prev {
			case chPlay:
				midi := evalArg(getArg(args, 0, strconv.FormatFloat(ins.Midi, 'g', 8, 64)))
				dur := evalArg(getArg(args, 1, strconv.FormatInt(int64(ins.Dur), 10)))
				vol := evalArg(getArg(args, 2, strconv.FormatFloat(ins.Vol, 'g', 8, 64)))
				if math.IsNaN(midi + dur + vol) {
					return nil, fmt.Errorf("cannot parse arguments for %s", raw)
				}
				ins = &ir.Instruction{Midi: midi, Dur: int(dur), Vol: vol, Time: t, Info: raw}
				irp.Add(ins)
				ta = int(dur)
			case chLink:
				newArgs := make([]string, max(len(prevLinkArgs), len(args)))
				for i := range newArgs {
					newArgs[i] = getArg(args, i, getArg(prevLinkArgs, i, ""))
				}
				irpNested, err := b.GenSeqIR(prevLinkName, strings.Join(newArgs, ":"))
				if err != nil {
					return nil, err
				}
				irpNested = irpNested.Copy()
				irpNested.Shift(t)
				irp.Merge(irpNested)
				ta = irpNested.Duration()
			default:
				return nil, fmt.Errorf("%s can't be used after %s", chPrev, prev)
			}

		case chVibe:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)

		case chRest:
			t += ta
			ta = int(evalArg(getArg(args, 0, "1")))

		case chWith:
			ta = 0
		}
	}
	irp.Sort()
	return irp, nil
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%v", tokens)
	logs.Debug("Lane tokens %s\n", tokens)
	return nil, fmt.Errorf("sequence riff type not implemented yet")
}
