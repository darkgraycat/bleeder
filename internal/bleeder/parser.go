package bleeder

import (
	"bleeder/internal/audio"
	"math"
	"strconv"
	"strings"
)

const (
	chPlay = ">" // play note or midi
	chPrev = "<" // repeat last note or sequence
	chLink = "@" // play sequence
	chVibe = "$" // switch vibe
	chRest = "_" // delay next operation
	chWith = "|" // play operations in parallel
	chArgs = ":" // operation arguments separator
	chSkip = "#" // skip next operations in line
)

// helper replacer to format sequence content
var replacer = strings.NewReplacer(
	chPlay, " "+chPlay,
	chPrev, " "+chPrev,
	chLink, " "+chLink,
	chVibe, " "+chVibe,
	chRest, " "+chRest,
	chWith, " "+chWith,
	chSkip, " "+chSkip,
)

// tokenize sequence raw content
func tokenizeContent(s string) [][]string {
	out := make([][]string, 0, 4)
	pre := strings.TrimSpace(replacer.Replace(s))
	for row := range strings.SplitSeq(pre, "\n") {
		if idx := strings.IndexByte(row, chSkip[0]); idx >= 0 {
			row = row[:idx]
		}
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
		k, v, _ := strings.Cut(d, chArgs)
		if i < len(values) {
			out[k] = evalArg(getArg(values, i, v))
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
			v = strconv.FormatFloat(ref, 'g', 8, 64) + v[pos:]
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
		pairs = append(pairs, k, strconv.FormatFloat(v, 'g', 8, 64))
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

// get argument at position idx or fallback to prev value
func getArg(args []string, idx int, prev string) string {
	if idx >= len(args) {
		return prev
	}
	v := args[idx]
	if strings.IndexAny(v, "+-*/") == 0 {
		return prev + v
	}
	return v
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

// split raw arguments into slice
func splitArgs(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, chArgs)
}
