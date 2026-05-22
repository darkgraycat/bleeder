package bleeder

import "strings"

// helper character
const splitChar = "\\"

// Lane-DSL operators
const (
	lcNone = ""
	lcMidi = ">"
	lcNote = ":"
	lcFreq = "~"
	lcWait = "_"
	lcLast = "|"
	lcLink = "@"
	lcVibe = "$"
)

// cached Lane-DSL replacer
var lcFormatter = strings.NewReplacer(
	"\n", " ", // trim newline
	"\t", " ", // trim tabchar
	lcMidi, splitChar+lcMidi,
	lcNote, splitChar+lcNote,
	lcFreq, splitChar+lcFreq,
	lcWait, splitChar+lcWait,
	lcLast, splitChar+lcLast,
	lcLink, splitChar+lcLink,
	lcVibe, splitChar+lcVibe,
)

const (
	rcNone = ""
	rcRest = "-"
	rcFill = ">"
)

// tokenize Lane-DSL
func tokenizeLaneContent(s string) [][]string {
	replaced := strings.TrimSpace(lcFormatter.Replace(s))
	if len(replaced) < 1 {
		return nil
	}
	out := make([][]string, 0, len(replaced))
	for raw := range strings.SplitSeq(replaced[1:], splitChar) {
		out = append(out, strings.Fields(raw))
	}
	return out
}

// normalize lane-DSL into separate lines
func normalizeLaneContent(s string) []string {
	formatted := strings.TrimSpace(lcFormatter.Replace(s))
	if len(formatted) < 1 {
		return nil
	}
	return strings.Split(formatted[1:], splitChar)
}

// normalize riff-DSL into separate lines
func normalizeRiffContent(s string) []string {
	return strings.Fields(s) // TODO
}

// normalize vibe-DSL into separate lines
func normalizeVibeContent(s string) []string {
	return strings.Fields(s) // TODO
}

func extractSequenceVariables(s string) {
	// vars := make(map[string]string)
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
