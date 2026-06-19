package bleeder

import (
	"bleeder/internal/audio"
	"math"
	"strconv"
	"strings"
)

const (
	// DSL characters
	opPlay = ">"
	opLast = "<"
	opLink = "@"
	opVibe = "$"
	opRest = "_"
	opWith = "|"
	opArgs = ":"
)

var replacer = strings.NewReplacer(
	opPlay, " "+opPlay,
	opLast, " "+opLast,
	opLink, " "+opLink,
	opVibe, " "+opVibe,
	opRest, " "+opRest,
	opWith, " "+opWith,
)

// tokenize sequence raw content
func tokenizeContent(s string) [][]string {
	out := make([][]string, 0, 4)
	pre := strings.TrimSpace(replacer.Replace(s))
	for row := range strings.SplitSeq(pre, "\n") {
		if ts := strings.Fields(row); len(ts) > 0 {
			out = append(out, ts)
		}
	}
	return out
}

// parse sequence variables into map
func parseVars(s string, values []string) map[string]float64 {
	defs := strings.Fields(s)
	out := make(map[string]float64, len(defs))
	for i, d := range defs {
		k, v, _ := strings.Cut(d, ":")
		if i < len(values) {
			out[k] = evalArg(values[i])
			continue
		}
		pos := strings.IndexAny(v, "+-*/")
		lhs := v
		if pos >= 0 {
			lhs = v[:pos]
		}
		if ref, ok := out[lhs]; ok {
			if pos < 0 {
				out[k] = ref
				continue
			}
			v = strconv.FormatFloat(ref, 'f', -1, 64) + v[pos:]
		}
		out[k] = evalArg(v)
	}
	return out
}

// apply sequence variable to content
func applyVars(s string, vars map[string]float64) string {
	if len(vars) == 0 {
		return s
	}
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, k, strconv.FormatFloat(v, 'g', 2, 64))
	}
	return strings.NewReplacer(pairs...).Replace(s)
}

// evaluate argument with short arithmetic expression
func evalArg(s string) float64 {
	i := strings.IndexAny(s, "+-*/")
	if i < 0 {
		return parseTone(s)
	}
	lhs := parseTone(s[:i])
	rhs := parseTone(s[i+1:])
	switch s[i] {
	case '+':
		return lhs + rhs
	case '-':
		return lhs - rhs
	case '*':
		return lhs * rhs
	case '/':
		return lhs / rhs
	}
	return math.NaN()
}

// parse string which represents tone. can be midi or note
func parseTone(s string) float64 {
	if m := audio.NoteToMidi(s); m >= 0 {
		return float64(m)
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return math.NaN()
}

// Parse sequence content into IR Program
// func ParseContent(content string, context *ParserContext) (*ir.Program, error) {
// 	irp := ir.NewProgram()
// 	replaced := lcFormatter.Replace(content)
// 	if len(replaced) < 1 {
// 		return irp, nil
// 	}

// 	offset := 0
// 	lastDelay := 0
// 	lastInsOp := ""
// 	ins := &ir.Instruction{Info: "None"}

// 	for raw := range strings.SplitSeq(replaced[1:], lcSplit) {
// 		op := string(raw[0])
// 		args := strings.Fields(raw[1:])

// 		switch op {
// 		case lcMidi: // > midi operator
// 			ins = &ir.Instruction{
// 				Midi: audio.MidiToFreq(int(getOpArg(args, 0, 60))),
// 				Dur:  int(getOpArg(args, 1, 1)),
// 				Vol:  getOpArg(args, 2, 1.0),
// 				Time: offset,
// 				Info: raw,
// 			}
// 			lastInsOp = op
// 			lastDelay = 0
// 			irp.Add(ins)

// 		case lcNote: // : note operator
// 			ins = &ir.Instruction{
// 				Midi: audio.MidiToFreq(int(getOpNoteArg(args, 0, "c4"))),
// 				Dur:  int(getOpArg(args, 1, 1)),
// 				Vol:  getOpArg(args, 2, 1.0),
// 				Time: offset,
// 				Info: raw,
// 			}
// 			lastInsOp = op
// 			lastDelay = 0
// 			irp.Add(ins)

// 		case lcFreq: // ~ freq operator
// 			ins = &ir.Instruction{
// 				Midi: getOpArg(args, 0, audio.BaseToneFreq),
// 				Dur:  int(getOpArg(args, 1, 1)),
// 				Vol:  getOpArg(args, 2, 1.0),
// 				Time: offset,
// 				Info: raw,
// 			}
// 			lastInsOp = op
// 			lastDelay = 0
// 			irp.Add(ins)

// 		case lcLink: // @ link operator
// 			if context == nil || context.ResolveFunc == nil {
// 				return nil, fmt.Errorf("%s not supported without context", op)
// 			}
// 			if len(args) == 0 {
// 				return nil, fmt.Errorf("%s requires a sequence name", op)
// 			}
// 			irpNested, err := context.ResolveFunc(args[0], args[1:])
// 			if err != nil {
// 				return nil, err
// 			}
// 			irpNested = irpNested.Copy()
// 			irpNested.Shift(offset)
// 			lastInsOp = op
// 			irp.Merge(irpNested)

// 		case lcWait: // _ wait operator
// 			lastDelay = int(getOpArg(args, 0, float64(ins.Dur)))
// 			offset += lastDelay

// 		case lcLast: // | last operator
// 			freq := ins.Midi
// 			switch lastInsOp {
// 			case lcMidi, lcNote:
// 				freq = audio.MidiToFreq(int(getOpArg(args, 0, float64(audio.FreqToMidi(ins.Midi)))))
// 			case lcFreq:
// 				freq = getOpArg(args, 0, freq)
// 			case lcLink:
// 				return nil, fmt.Errorf("%s after %s is not implemented yet", lcLast, lastInsOp)
// 			}
// 			ins = &ir.Instruction{
// 				Midi: freq,
// 				Dur:  int(getOpArg(args, 1, float64(ins.Dur))),
// 				Vol:  getOpArg(args, 2, ins.Vol),
// 				Time: offset,
// 				Info: "REPEAT" + raw,
// 			}
// 			offset += lastDelay
// 			irp.Add(ins)
// 		}
// 	}

// 	for i, ins := range irp.Instructions() {
// 		fmt.Printf("%d - %s\n", i, ins)
// 	}

// 	return irp, nil
// }
