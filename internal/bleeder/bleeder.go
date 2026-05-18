package bleeder

import (
	"bleeder/internal/ir"
	"bleeder/internal/shared/logs"
	"fmt"
	"strings"
)

// Core DSL processor and IRs generator
type Bleeder struct {
	context *ParserContext
	lanes   map[string]*Sequence
	riffs   map[string]*Sequence
	irps    map[string]*ir.Program
	main    string
}

// Create new Bleeder instance
func NewBleeder() *Bleeder {
	b := &Bleeder{
		lanes: make(map[string]*Sequence),
		riffs: make(map[string]*Sequence),
		irps:  make(map[string]*ir.Program),
	}
	b.context = &ParserContext{
		ResolveFunc: b.GenSeqIR,
	}
	return b
}

// Load bleed into Bleeder
func (b *Bleeder) Bleed(bleed *Bleed) (*Bleeder, error) {
	logs.Trace(logs.INFO, "called")
	// store contents
	for key, lane := range bleed.Lanes {
		b.lanes[key] = &lane
	}
	for key, riff := range bleed.Riffs {
		b.riffs[key] = &riff
	}

	// store entrypoint
	b.main = bleed.Meta.Main

	// TODO: process and cache sequences
	return b, nil
}

// Get IR of the main sequence
func (b *Bleeder) GenMainIR() (*ir.Program, error) {
	logs.Trace(logs.INFO, "called")
	return b.GenSeqIR(b.main, nil)
}

// Get IR of specified section with args
func (b *Bleeder) GenSeqIR(name string, args []string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called")
	irp := b.getCachedIR(name, args)
	if irp != nil {
		return irp, nil
	}
	seq, seqType := b.getCachedSequence(name)
	if seq == nil {
		return nil, fmt.Errorf("Sequence is not found: %s", name)
	}

	logs.Debug("---- %s", seqType)
	// 1. parse sequence with args
	//   - expand args
	//   - replace content
	// 2. write into cache
	// 3. return it
	return nil, nil
}

// Get IR of raw DSL
func (b *Bleeder) GenIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called")
	return ParseContent(content, &ParserContext{
		ResolveFunc: func(name string, args []string) (*ir.Program, error) {
			// TODO
			return nil, nil
		},
	})
}

func (b *Bleeder) getCachedIR(name string, args []string) *ir.Program {
	key := name
	if len(args) > 0 {
		key += ":" + strings.Join(args, ",")
	}
	if irp, ok := b.irps[key]; ok {
		return irp
	}
	return nil
}

func (b *Bleeder) getCachedSequence(name string) (*Sequence, SequenceType) {
	if lane, ok := b.lanes[name]; ok {
		return lane, SEQ_LANE
	}
	if riff, ok := b.riffs[name]; ok {
		return riff, SEQ_RIFF
	}
	return nil, SEQ_UNKNOWN
}
