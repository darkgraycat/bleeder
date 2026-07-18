package core

import (
	"bleeder/internal/audio"
	"math"
	"sort"
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
func parseVars(s string, vals []string) map[string]float64 {
	defs := strings.Fields(s)
	out := make(map[string]float64, len(defs))
	for i, def := range defs {
		k, v, _ := strings.Cut(def, chArgs)
		if i < len(vals) && vals[i] != "" {
			switch vals[i][0] {
			case '+', '-', '*', '/':
				v += vals[i]
			default:
				v = vals[i]
			}
		}
		out[k] = evalVars(v, out)
	}
	return out
}

// evaluate arithmetic expression with variables map
func evalVars(s string, vars map[string]float64) float64 {
	i := strings.LastIndexAny(s, "+-*/")
	if i > 0 {
		lhs := evalVars(s[:i], vars)
		rhs := evalVars(s[i+1:], vars)
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
	if ref, ok := vars[s]; ok {
		return ref
	}
	return parseTone(s)
}

// apply sequence variable to content
func applyVars(s string, vars map[string]float64) string {
	if len(vars) == 0 {
		return s
	}
	pairs := make([][2]string, 0, len(vars))
	for k, v := range vars {
		vStr := strconv.FormatFloat(v, 'g', 8, 64)
		if v < 0 { // convert -1 to absolute value 0-1
			vStr = "0" + vStr
		}
		pairs = append(pairs, [2]string{k, vStr})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i][0]) > len(pairs[j][0])
	})

	var result strings.Builder
	result.Grow(len(s))
outer:
	for i := 0; i < len(s); {
		for _, pair := range pairs {
			k, v := pair[0], pair[1]
			if strings.HasPrefix(s[i:], k) {
				nextPos := i + len(k)
				if (i == 0 || !isAlphaNum(s[i-1])) &&
					(nextPos >= len(s) || !isAlphaNum(s[nextPos])) {
					result.WriteString(v)
					i = nextPos
					continue outer
				}

			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

// get argument at position idx or use fallback value
func getArg(args []string, idx int, fallback string) string {
	if idx >= len(args) || args[idx] == "" {
		return fallback
	}
	v := args[idx]
	switch v[0] {
	case '+', '-', '*', '/':
		return fallback + v
	}
	return v
}

// split raw arguments into slice using special character
func splitArgs(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, chArgs)
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

// checks if character is alphanumeric
func isAlphaNum(c byte) bool {
	return (c|0x20 >= 'a' && c|0x20 <= 'z') ||
		(c >= '0' && c <= '9')
}
