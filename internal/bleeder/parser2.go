package bleeder

import (
	"bleeder/internal/audio"
	"math"
	"strconv"
	"strings"
)

const (
	// Lane-DSL characters
	lcMidi = ">"
	lcNote = ":"
	lcFreq = "~"
	lcWait = "_"
	lcLast = "|"
	lcLink = "@"
	lcVibe = "$"
	// Riff-DSL characters
	rcRest = "-"
	rcFill = ">"
	// helper characters
	lcSplit = "\\"
	rcSplit = "\n"
)

// Lane-DSL cached replacer
var lcFormatter = strings.NewReplacer(
	" ", "", // trim space ch
	"\n", "", // trim newline
	"\t", "", // trim tabchar
	",", " ",
	lcMidi, lcSplit+lcMidi,
	lcNote, lcSplit+lcNote,
	lcFreq, lcSplit+lcFreq,
	lcWait, lcSplit+lcWait,
	lcLast, lcSplit+lcLast,
	lcLink, lcSplit+lcLink,
	lcVibe, lcSplit+lcVibe,
)

// Riff-DSL cached replacer
var rcFormatter = strings.NewReplacer(
	"\t", "", // trim tabchar
	" ", "", // trim space ch
)

// normalize Lane-DSL into separate lines
func normalizeLaneContent(s string) []string {
	pre := lcFormatter.Replace(s)
	out := strings.Split(pre, lcSplit)
	if len(out) < 1 {
		return nil
	}
	return out[1:]
}

// normalize Riff-DSL into separate lines
func normalizeRiffContent(s string) []string {
	pre := rcFormatter.Replace(s)
	out := strings.Split(strings.TrimSpace(pre), rcSplit)
	if len(out) < 1 {
		return nil
	}
	return out
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
	return -1
}

// parse sequence variables into map
func parseVars(s string, values []string) map[string]float64 {
	defs := strings.Fields(s)
	out := make(map[string]float64, len(defs))
	for i, d := range defs {
		k, v, _ := strings.Cut(d, "=")
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
		pairs = append(pairs, k, strconv.FormatFloat(v, 'f', -1, 64))
	}
	return strings.NewReplacer(pairs...).Replace(s)
}
