package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"fmt"
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
	return b.GenSeqIR(b.main, nil)
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, args []string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called")
	return nil, nil
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	return nil, fmt.Errorf("Sequence lane type not implemented yet")
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	return nil, fmt.Errorf("Sequence riff type not implemented yet")
}
