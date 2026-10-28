package core

import (
	"math"
	"strconv"
	"strings"
)

var formatter = strings.NewReplacer(
	"[", " [ ", "]", " ] ", "(", " ( ", ")", " ) ", "{", " { ", "}", " } ",
	"@", " @ ", "$", " $ ", "&", " & ", "|", " | ", ":", " : ", "_", " _ ",
	"+", " + ", "-", " - ", "*", " * ", "/", " / ", "%", " % ", "^", " ^ ",
)

// Scan sequence raw content for rows of lexems
func scan(raw string) [][]string {
	out := make([][]string, 0, 8)
	buf := make([]string, 0, 8)
	normal := strings.TrimSpace(formatter.Replace(raw))

	for line := range strings.SplitSeq(normal, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		var inSeq, inLink, inEach bool
		for tok := range strings.FieldsSeq(line) {
			seq, link := true, inLink
			switch tok {
			case "+", "-", "*", "/", "%", "^", "|", ":":
				seq = false
			case "&":
				seq = inEach
			case "[":
				inEach = true
			case "]":
				inEach = false
			case "@", "$", "(", "{":
				link = true
			case ")", "}":
				link = false
			}
			if seq && inSeq && !inLink {
				out = append(out, append([]string(nil), buf...))
				buf = buf[:0]
			}
			inSeq, inLink = seq, link
			buf = append(buf, tok)
		}
		if len(buf) > 0 {
			out = append(out, append([]string(nil), buf...))
			buf = buf[:0]
		}
	}

	return out
}

// Flat rows of lexems by expanding each groups
func flat(frames [][]string) [][]string {
	out := make([][]string, 0, 8)
	base := make([][]string, 0, 4)
	next := make([][]string, 0, 4)
	var conn []string
	inGroup := false

	for _, f := range frames {
		b, e := f[0], f[len(f)-1]
		switch {
		case b == "]" && e == "[":
			base, next = next, base[:0]
			conn = f[1 : len(f)-1]
		case e == "[":
			base, next = append(base[:0], nil), next[:0]
			conn = f[:len(f)-1]
			inGroup = true
		case b == "]":
			for _, x := range next {
				if len(x) > 0 && x[0] == "&" {
					out = append(out, x)
					continue
				}
				out = append(out, append(x, f[1:]...))
			}
			inGroup = false
		case !inGroup:
			out = append(out, f)
		case b == "&":
			next = append(next, f)
		default:
			for _, x := range base {
				if len(x) > 0 && x[0] == "&" {
					next = append(next, x)
					continue
				}
				row := make([]string, 0, len(x)+len(conn)+len(f)+4)
				row = append(append(append(row, x...), conn...), f...)
				next = append(next, row)
			}
		}
	}

	return out
}

// TODO:
// To complete the list of functions parser should have we need:
func vars(raw string) (vars map[string]string) {
	// to parse args from sequence
	return vars
}

// NOT USED
// TODO: write "eval function" that actualy evaluates whole
// evaluate arithmetic expression with variables map
func evalVars2(s string, vars map[string]string) float64 {
	i := strings.LastIndexAny(s, "+-*/%^")
	if i > 0 {
		lhs := evalVars2(s[:i], vars)
		rhs := evalVars2(s[i+1:], vars)
		switch s[i] {
		case '+':
			return lhs + rhs
		case '-':
			return lhs - rhs
		case '*':
			return lhs * rhs
		case '/':
			return lhs / rhs
		case '%':
			return math.Mod(lhs, rhs)
		case '^':
			return math.Pow(lhs, rhs)
		}
		return math.NaN()
	}
	if ref, ok := vars[s]; ok {
		val, err := strconv.ParseFloat(ref, 64)
		if err != nil {
			return math.NaN()
		}
		return val
	}
	return parseTone(s)
}
