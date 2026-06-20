package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"fmt"
	"math"
	"slices"
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
	varsMap := parseVars(seq.Vars, strings.Split(vars, chArgs))
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
	fmt.Printf("IR Len%d\n", irp.Length())
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
	prev := ""
	t, ta := 0, 0
	logs.Debug("Lane tokens %s\n", concated)

	for _, raw := range concated {
		ch := string(raw[0])
		args := strings.Split(raw[1:], chArgs)
		fmt.Printf("[OP] %s - %v\n", ch, args)
		switch ch {
		case chPlay:
			midi := evalArg(getArg(args, 0, ""))
			dur := evalArg(getArg(args, 1, "1"))
			vol := evalArg(getArg(args, 2, "1"))
			if math.IsNaN(midi) {
				return nil, fmt.Errorf("cannot parse midi tone for %s", raw)
			}
			ins = &ir.Instruction{Midi: midi, Dur: int(dur), Vol: vol, Time: t, Info: raw}
			ta += int(dur)
			irp.Add(ins)
			prev = ch
		case chPrev:
			switch prev {
			case chPlay:
			case chLink:
				return nil, fmt.Errorf("%s after %s is not implemented yet", chPrev, prev)
			default:
				return nil, fmt.Errorf("%s can't be used after %s", chPrev, prev)
			}
		case chLink:
			irpNested, err := b.GenSeqIR(args[0], strings.Join(args[1:], ":"))
			if err != nil {
				return nil, err
			}
			irpNested = irpNested.Copy()
			irpNested.Shift(0) // TODO
			// lastInsOp = op
			irp.Merge(irpNested)
		case chVibe:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)
		case chRest:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)
		case chWith:
			return nil, fmt.Errorf("%s operator is not implemented yet", ch)
		}
	}
	return irp, nil
	// return irp, fmt.Errorf("sequence lane type not implemented yet")
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%v", tokens)
	logs.Debug("Lane tokens %s\n", tokens)
	return nil, fmt.Errorf("sequence riff type not implemented yet")
}
