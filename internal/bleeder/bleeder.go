package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"fmt"
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

	varsMap := parseVars(seq.Vars, strings.Split(vars, ","))
	fmt.Printf("VARS: %v\n", varsMap)

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
		irp, err = b.genLaneIR("TODO")
	case SEQ_RIFF:
		irp, err = b.genRiffIR("TODO")
	default:
		return nil, fmt.Errorf("unknown sequence type: %s", name)
	}
	if err == nil {
		b.cache.Set(name, vars, irp)
	}
	return irp, err
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	normalized := normalizeLaneContent(content)
	if normalized == nil {
		return nil, fmt.Errorf("sequence content is invalid or empty")
	}
	fmt.Printf("LANE %s\n", normalized)
	return nil, fmt.Errorf("sequence lane type not implemented yet")
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	normalized := normalizeRiffContent(content)
	if normalized == nil {
		return nil, fmt.Errorf("sequence content is invalid or empty")
	}
	fmt.Printf("RIFF %s\n", normalized)
	return nil, fmt.Errorf("sequence riff type not implemented yet")
}
