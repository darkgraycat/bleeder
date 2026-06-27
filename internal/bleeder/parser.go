package bleeder

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
func parseVars(s string, values []string) map[string]float64 {
	defs := strings.Fields(s)
	out := make(map[string]float64, len(defs))
	for i, def := range defs {
		k, v, _ := strings.Cut(def, chArgs)
		if i < len(values) && values[i] != "" {
			if isModCh(values[i][0]) {
				v += values[i]
			} else {
				v = values[i]
			}
		}
		pos := strings.IndexAny(v, "+-*/")
		if pos < 0 {
			if ref, ok := out[v]; ok {
				v = strconv.FormatFloat(ref, 'g', -1, 64)
			}
		} else {
			if ref, ok := out[v[:pos]]; ok {
				v = strconv.FormatFloat(ref, 'g', -1, 64) + v[pos:]
			}
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
	pairs := make([][2]string, 0, len(vars))
	for k, v := range vars {
		pairs = append(pairs, [2]string{k, strconv.FormatFloat(v, 'g', 8, 64)})
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

// evaluate argument with short arithmetic expression
func evalArg(s string) float64 {
	i := strings.LastIndexAny(s, "+-*/")
	if i < 0 {
		return parseTone(s)
	}
	lhs := evalArg(s[:i])
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
	if idx >= len(args) || args[idx] == "" {
		return prev
	}
	v := args[idx]
	if isModCh(v[0]) {
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

// checks if character is one of +-*/
func isModCh(c byte) bool {
	return c == '+' || c == '-' ||
		c == '*' || c == '/'
}

// checks if character is alphanumeric
func isAlphaNum(c byte) bool {
	return (c|0x20 >= 'a' && c|0x20 <= 'z') ||
		(c >= '0' && c <= '9')
}
