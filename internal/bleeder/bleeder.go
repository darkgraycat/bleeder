package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"fmt"
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
	logs.Trace(logs.INFO, "called")
	irp := b.cache.Get(name, vars)
	if irp != nil {
		return irp, nil
	}
	seq, ok := b.sqncs[name]
	if !ok {
		return nil, fmt.Errorf("sequence is not exist: %s", name)
	}
	fmt.Printf("SEQ %s, %s\n", name, vars)
	varsMap := parseVars(seq.Vars, strings.Split(vars, opArgs))
	rawContent := applyVars(seq.Content, varsMap)
	fmt.Printf("%v\n%v\n", varsMap, rawContent)

	tokens := tokenizeContent(rawContent)
	if tokens == nil {
		return nil, fmt.Errorf("sequence content is invalid or empty")
	}

	// 1. substitute "e2"-like notation with midi numbers for seq.Vars -> prepared vars ?
	// note - we are not going to do pre-substution
	// 2. do all arithmetic modifications for prepared vars -> calculated vars
	// Ex: a=e2 b=2 c=a+2.5 with a3 1	 ->  a=57 b=1 c=58.5
	//	or a=e2 b=2 c=a+2.5 with a3 1 e2 ->  a=57 b=1 c=40 -- overrides "a+2.5" completelly
	// note we cant do evalArg on content right now, because it depends on flow

	// 3. substitute seq.Content with calculated vars -> prepared content
	// Note: prepared content still might have modifications that cant be calculated without actual parsing, ex: "|+7" for lane

	// 4. parse prepared content into IR
	var err error
	switch seq.Type {
	case SEQ_LANE:
		irp, err = b.genLaneIR(tokens)
	case SEQ_RIFF:
		irp, err = b.genRiffIR(tokens)
	default:
		return nil, fmt.Errorf("unknown sequence type: %s", name)
	}
	if err == nil {
		b.cache.Set(name, vars, irp)
	}
	return irp, err
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(tokens [][]string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%v", tokens)
	concated := slices.Concat(tokens...)
	irp := ir.NewProgram()
	ins := &ir.Instruction{Info: "None"}

	for _, raw := range concated {
		op := string(raw[0])
		args := strings.Split(raw[1:], opArgs)
		fmt.Printf("OP %s - %v\n", op, args)
		switch op {
		case opPlay:
			ins = &ir.Instruction{
				Midi: evalArg(args[0]),
				// TODO
			}
			// TODO
			irp.Add(ins)
		case opLast:
		case opLink:
			irpNested, err := b.GenSeqIR(args[0], strings.Join(args[1:], ":"))
			if err != nil {
				return nil, err
			}
			irpNested = irpNested.Copy()
			irpNested.Shift(0) // TODO
			// lastInsOp = op
			irp.Merge(irpNested)
		case opVibe:
		case opRest:
		case opWith:
		}
	}
	fmt.Printf("LANE TOKENS %s\n", concated)
	return irp, fmt.Errorf("sequence lane type not implemented yet")
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(tokens [][]string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%v", tokens)
	fmt.Printf("RIFF TOKENS %s\n", tokens)
	return nil, fmt.Errorf("sequence riff type not implemented yet")
}
