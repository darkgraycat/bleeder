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

// Create new Bleeder instance from a loaded Bleed
func NewBleeder(bleed *Bleed) *Bleeder {
	logs.Trace(logs.INFO, "called")
	b := &Bleeder{
		lanes: make(map[string]*Sequence),
		riffs: make(map[string]*Sequence),
		irps:  make(map[string]*ir.Program),
		main:  bleed.Meta.Main,
	}
	for key, lane := range bleed.Lanes {
		b.lanes[key] = &lane
	}
	for key, riff := range bleed.Riffs {
		b.riffs[key] = &riff
	}
	b.context = &ParserContext{
		ResolveFunc: b.GenSeqIR,
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
	irp := b.getCachedIR(name, args)
	if irp != nil {
		return irp, nil
	}
	seq, seqType := b.getCachedSequence(name)
	if seq == nil {
		return nil, fmt.Errorf("Sequence is not found: %s", name)
	}

	// TODO: we need to know seqType to expand arguments
	irp, err := b.GenIR("", seqType)

	// var err error
	switch seqType {
	case SEQ_LANE:
		irp, err = b.genLaneIR("") // TODO: use expanded content
	case SEQ_RIFF:
		irp, err = b.genRiffIR("") // TODO: use expanded content
	}
	if err != nil {
		return nil, err
	}

	// 1. parse sequence with args
	//   - expand args
	//   - replace content
	// 2. write into cache
	// 3. return it
	return nil, nil
}

// Get IR of raw DSL by type
func (b *Bleeder) GenIR(content string, seqType SequenceType) (*ir.Program, error) {
	switch seqType {
	case SEQ_LANE:
		return b.genLaneIR(content)
	case SEQ_RIFF:
		return b.genRiffIR(content)
	}
	return nil, fmt.Errorf("Unknown sequence type: %d", seqType)
}

// Get IR from raw Lane-DSL
func (b *Bleeder) genLaneIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	// irp := ir.NewProgram()

	return ParseContent(content, b.context)
}

// Get IR from raw Riff-DSL
func (b *Bleeder) genRiffIR(content string) (*ir.Program, error) {
	logs.Trace(logs.INFO, "called with\n%s", content)
	return nil, fmt.Errorf("Sequence riff type not implemented yet")
}

// get IR cached by name and args
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

// get Sequence with SequenceType cached by name
func (b *Bleeder) getCachedSequence(name string) (*Sequence, SequenceType) {
	if lane, ok := b.lanes[name]; ok {
		return lane, SEQ_LANE
	}
	if riff, ok := b.riffs[name]; ok {
		return riff, SEQ_RIFF
	}
	return nil, SEQ_UNKNOWN
}
