package bleeder

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"bleeder/internal/shared"
	"fmt"
	"strconv"
	"strings"
)

const (
	opMidi = ">"
	opNote = ":"
	opFreq = "~"
	opLink = "@"
	opWait = "_"
	opLast = "|"
)

const opSplitter = "\\"

var replacer = strings.NewReplacer(
	"\n", " ", // trim newline
	"\t", " ", // trim tabchar
	opMidi, opSplitter+opMidi,
	opNote, opSplitter+opNote,
	opFreq, opSplitter+opFreq,
	opLink, opSplitter+opLink,
	opWait, opSplitter+opWait,
	opLast, opSplitter+opLast,
)

type ParserContext struct {
}

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

// Parse sequence raw content into IR Program
func ParseRawContent(s string, t int) (*ir.Program, error) {
	pr := ir.NewProgram()
	replaced := replacer.Replace(s)
	if len(replaced) < 1 {
		return pr, nil
	}

	lastInsOp := opWait
	lastDelay := 0
	ins := &ir.Instruction{Info: "Noop"}

	for raw := range strings.SplitSeq(replaced[1:], opSplitter) {
		op := string(raw[0])
		args := strings.Fields(raw[1:])
		fmt.Printf("Line %s - %v\n", op, args) // TODO: remove log

		switch op {
		// >
		case opMidi:
			ins = &ir.Instruction{
				Freq: audio.MidiToFreq(int(getOpArg(args, 0, 60))),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			pr.Add(ins)
		// :
		case opNote:
			ins = &ir.Instruction{
				Freq: audio.MidiToFreq(int(getOpNoteArg(args, 0, "c4"))),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			pr.Add(ins)
		// ~
		case opFreq:
			ins = &ir.Instruction{
				Freq: getOpArg(args, 0, audio.C4freq),
				Dur:  int(getOpArg(args, 1, 1)),
				Vol:  getOpArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastInsOp = op
			lastDelay = 0
			pr.Add(ins)
		// @
		case opLink:
			// Q: how to implement? we need to have a context here somehow
			lastInsOp = op
			return nil, fmt.Errorf("not implemented yet: %s", op)
		// _
		case opWait:
			lastDelay = int(getOpArg(args, 0, float64(ins.Dur)))
			t += lastDelay
		// |
		case opLast:
			freq := ins.Freq
			switch lastInsOp {
			case opMidi, opNote:
				freq = audio.MidiToFreq(int(getOpArg(args, 0, freq)))
			case opFreq:
				freq = getOpArg(args, 0, freq)
			}
			ins = &ir.Instruction{
				Freq: freq,
				Dur:  int(getOpArg(args, 1, float64(ins.Dur))),
				Vol:  getOpArg(args, 2, ins.Vol),
				Time: t,
				Info: "REPEAT" + raw,
			}
			t += lastDelay
			pr.Add(ins)
		}
	}

	for i, ins := range pr.Instructions() {
		fmt.Printf("%d - %s\n", i, ins)
	}

	return pr, nil
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
