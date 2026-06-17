package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared"
	"fmt"
	"strconv"
	"strings"
)

// Context required to parse content
type ParserContext struct {
	ResolveFunc func(name string, args []string) (*ir.Program, error)
}

// Parse sequence content into IR Program
func ParseContent(content string, context *ParserContext) (*ir.Program, error) {
	irp := ir.NewProgram()
	replaced := lcFormatter.Replace(content)
	if len(replaced) < 1 {
		return irp, nil
	}

	offset := 0
	lastDelay := 0
	lastInsOp := ""
	ins := &ir.Instruction{Info: "None"}

	for raw := range strings.SplitSeq(replaced[1:], lcSplit) {
		op := string(raw[0])
		args := strings.Fields(raw[1:])

		switch op {
		case lcMidi: // > midi operator
			ins = &ir.Instruction{
				Midi: audio.MidiToFreq(int(getOpArg(args, 0, 60))),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: offset,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			irp.Add(ins)

		case lcNote: // : note operator
			ins = &ir.Instruction{
				Midi: audio.MidiToFreq(int(getOpNoteArg(args, 0, "c4"))),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: offset,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			irp.Add(ins)

		case lcFreq: // ~ freq operator
			ins = &ir.Instruction{
				Midi: getOpArg(args, 0, audio.BaseToneFreq),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: offset,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			irp.Add(ins)

		case lcLink: // @ link operator
			if context == nil || context.ResolveFunc == nil {
				return nil, fmt.Errorf("%s not supported without context", op)
			}
			if len(args) == 0 {
				return nil, fmt.Errorf("%s requires a sequence name", op)
			}
			irpNested, err := context.ResolveFunc(args[0], args[1:])
			if err != nil {
				return nil, err
			}
			irpNested = irpNested.Copy()
			irpNested.Shift(offset)
			lastInsOp = op
			irp.Merge(irpNested)

		case lcWait: // _ wait operator
			lastDelay = int(getOpArg(args, 0, float64(ins.Dur)))
			offset += lastDelay

		case lcLast: // | last operator
			freq := ins.Midi
			switch lastInsOp {
			case lcMidi, lcNote:
				freq = audio.MidiToFreq(int(getOpArg(args, 0, float64(audio.FreqToMidi(ins.Midi)))))
			case lcFreq:
				freq = getOpArg(args, 0, freq)
			case lcLink:
				return nil, fmt.Errorf("%s after %s is not implemented yet", lcLast, lastInsOp)
			}
			ins = &ir.Instruction{
				Midi: freq,
				Dur:  int(getOpArg(args, 1, float64(ins.Dur))),
				Vol:  getOpArg(args, 2, ins.Vol),
				Time: offset,
				Info: "REPEAT" + raw,
			}
			offset += lastDelay
			irp.Add(ins)
		}
	}

	for i, ins := range irp.Instructions() {
		fmt.Printf("%d - %s\n", i, ins)
	}

	return irp, nil
}

// TODO: not used yet
// Expand sequence arguments to produce raw content
func ExpandArgs(content string, args []string) (string, error) {
	// TODO: anyway we need to understand how to implement @ in ParseContent first
	// to do so, we need provide ParseContent with context of whole bleed + incuded
	for i, arg := range args {
		fmt.Printf("Arg %d - %s\n", i, arg)
		lhs, rhs, op := splitOpArgs(arg)
		fmt.Printf("L %s\top %s\tR %s\n", lhs, op, rhs)
	}

	// pairs := append([]string(nil), args...)
	return "", nil
}

// TODO: not used yet
// Parse sequence raw arguments into []string
func ParseRawArgs(s string) ([]string, error) {
	args := make([]string, 0)
	for part := range strings.FieldsSeq(s) {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			return nil, fmt.Errorf("invalid arg: %q", part)
		}
		args = append(args, k, v)
	}
	return args, nil
}

// helpers

// get nth numeric argument as float64
func getOpArg(args []string, idx int, def float64) float64 {
	if idx >= len(args) {
		return def
	}
	lhs, rhs, op := splitOpArgs(args[idx])
	return modOpArg(
		shared.Str2Float(lhs, def),
		shared.Str2Float(rhs, 0.0),
		op,
	)
}

// TODO: know how we could "inject" midi instead of note
// get nth note argument as float64
func getOpNoteArg(args []string, idx int, def string) float64 {
	d := float64(audio.NoteToMidi(def))
	if idx >= len(args) {
		return d
	}
	lhs, rhs, op := splitOpArgs(args[idx])
	midi := audio.NoteToMidi(lhs)
	return modOpArg(
		shared.Str2Float(strconv.Itoa(midi), d),
		shared.Str2Float(rhs, 0.0),
		op,
	)
}

// produce content with applied sequence variables
func applySeqVars(content string, vars string, values []string) string {
	defs := strings.Fields(vars)
	pairs := make([]string, 0, len(defs)*2)
	// fill with defined values
	for i, def := range defs[:min(len(values), len(defs))] {
		name, _, _ := strings.Cut(def, "=")
		pairs = append(pairs, name, values[i])
	}
	// fill with default values
	for _, def := range defs[len(values):] {
		name, defaultVal, _ := strings.Cut(def, "=")
		pairs = append(pairs, name, defaultVal)
	}
	return strings.NewReplacer(pairs...).Replace(content)
}

// parse sequence variables and produce pairs for replacer
func parseSeqVars(vars string, values []string) []string {
	defs := strings.Fields(vars)
	pairs := make([]string, len(defs)*2)
	for i, def := range defs {
		k, v, _ := strings.Cut(def, "=")
		if len(values) < i {
			v = values[i]
		}
		pairs[i*2] = k
		pairs[i*2+1] = v
	}
	return pairs
}

// split string by +-*/ operators
func splitOpArgs(s string) (lhs, rhs, op string) {
	for i := range s {
		switch s[i] {
		case '+', '-', '*', '/':
			return s[:i], s[i+1:], s[i : i+1]
		}
	}
	return s, "", ""
}

// apply modificator on two arguments
func modOpArg(a, b float64, op string) float64 {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		return a / b
	default:
		return a
	}
}
