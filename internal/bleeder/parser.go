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

// Get IR from raw sequence content
func ParseContent(content string, t int) (*ir.Program, error) {
	pr := ir.NewProgram()
	replaced := replacer.Replace(content)
	if len(replaced) < 1 {
		return pr, nil
	}

	lastOp := opWait
	lastDelay := 0
	in := &ir.Instruction{Info: "Noop"}

	for raw := range strings.SplitSeq(replaced[1:], opSplitter) {
		op := string(raw[0])
		args := strings.Fields(raw[1:])
		fmt.Printf("Line %s - %v\n", op, args) // TODO: remove log

		switch op {
		// >
		case opMidi:
			in = &ir.Instruction{
				Freq: audio.MidiToFreq(int(getArg(args, 0, 60))),
				Dur:  int(getArg(args, 1, 1)),
				Vol:  getArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastOp = op
			lastDelay = 0
			pr.Add(in)
		// :
		case opNote:
			in = &ir.Instruction{
				Freq: audio.MidiToFreq(int(getNoteArg(args, 0, "c4"))),
				Dur:  int(getArg(args, 1, 1)),
				Vol:  getArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastOp = op
			lastDelay = 0
			pr.Add(in)
		// ~
		case opFreq:
			in = &ir.Instruction{
				Freq: getArg(args, 0, audio.C4freq),
				Dur:  int(getArg(args, 1, 1)),
				Vol:  getArg(args, 2, 1.0),
				Time: t,
				Info: raw,
			}
			lastOp = op
			lastDelay = 0
			pr.Add(in)
		// @
		case opLink:
			lastOp = op
			return nil, fmt.Errorf("not implemented yet: %s", op)
		// _
		case opWait:
			lastDelay = int(getArg(args, 0, float64(in.Dur)))
			t += lastDelay
		// |
		case opLast:
			freq := in.Freq
			switch lastOp {
			case opMidi, opNote:
				freq = audio.MidiToFreq(int(getArg(args, 0, freq)))
			case opFreq:
				freq = getArg(args, 0, freq)
			}
			in = &ir.Instruction{
				Freq: freq,
				Dur:  int(getArg(args, 1, float64(in.Dur))),
				Vol:  getArg(args, 2, in.Vol),
				Time: t,
				Info: "REPEAT" + raw,
			}
			t += lastDelay
			pr.Add(in)
		}
	}

	// TODO: remove logs
	for i, in := range pr.Instructions() {
		fmt.Printf("%d - %f hz\to: %v\td: %v\t %s\n", i, in.Freq, in.Time, in.Dur, in.Info)
	}

	return pr, nil
}

func getArg(args []string, idx int, def float64) float64 {
	if idx >= len(args) {
		return def
	}
	lhs, op, rhs := splitOpArgs(args[idx])
	return getModArg(
		shared.Str2Float(lhs, def),
		shared.Str2Float(rhs, 0.0),
		op,
	)
}

func getNoteArg(args []string, idx int, def string) float64 {
	d := float64(audio.NoteToMidi(def))
	if idx >= len(args) {
		return d
	}
	lhs, op, rhs := splitOpArgs(args[idx])
	midi := audio.NoteToMidi(lhs)
	return getModArg(
		shared.Str2Float(strconv.Itoa(midi), d),
		shared.Str2Float(rhs, 0.0),
		op,
	)
}

func getModArg(a, b float64, op string) float64 {
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

func splitOpArgs(s string) (lhs, op, rhs string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '+', '-', '*', '/':
			return s[:i], s[i : i+1], s[i+1:]
		}
	}
	return s, "", ""
}
