package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var formatter = strings.NewReplacer(
	"[", " [ ", "]", " ] ", "(", " ( ", ")", " ) ", "{", " { ", "}", " } ",
	"@", " @ ", "$", " $ ", "&", " & ", "|", " | ", ":", " : ", "_", " _ ",
	"+", " + ", "-", " - ", "*", " * ", "/", " / ", "%", " % ", "^", " ^ ",
)

func split(content string) [][]string {
	out := make([][]string, 0, 8)
	buf := make([]string, 0, 16)
	src := strings.TrimSpace(formatter.Replace(content))
	for row := range strings.SplitSeq(src, "\n") {
		if i := strings.IndexByte(row, '#'); i >= 0 {
			row = row[:i]
		}
		prevIsValue := false
		prevIsGroup := false
		for raw := range strings.FieldsSeq(row) {
			nextIsValue := true
			nextIsGroup := prevIsGroup
			switch raw[0] {
			case '+', '-', '*', '/', '%', '^', '&', '|', ':':
				nextIsValue = false
			case '[', '(', '{', '@', '$':
				nextIsGroup = true
			case ']', ')', '}':
				nextIsGroup = false
			}
			if prevIsValue && nextIsValue && !prevIsGroup {
				out = append(out, append([]string(nil), buf...))
				buf = buf[:0]
			}
			buf = append(buf, raw)
			prevIsValue = nextIsValue
			prevIsGroup = nextIsGroup
		}
		if len(buf) > 0 {
			out = append(out, append([]string(nil), buf...))
			buf = buf[:0]
		}
	}
	return out
}

func normalize(content string) ([][]string, [][]string) {
	out := make([][]string, 0, 8)
	sub := make([][]string, 0, 4)
	buf := make([]string, 0, 8)
	grp := make([]string, 0, 4)
	src := strings.TrimSpace(formatter.Replace(content))

	for row := range strings.SplitSeq(src, "\n") {
		if i := strings.IndexByte(row, '#'); i >= 0 {
			row = row[:i]
		}
		prevIsValue := false
		prevIsJoins := false
		inEachGroup := false
		for raw := range strings.FieldsSeq(row) {
			nextIsValue := true
			nextIsJoins := prevIsJoins
			switch raw[0] {
			case '+', '-', '*', '/', '%', '^', '&', '|', ':':
				nextIsValue = false
			case '(', '{', '@', '$':
				nextIsJoins = true
			case ')', '}':
				nextIsJoins = false
			case '[':
				inEachGroup = true
				continue
			case ']':
				sub = append(sub, append([]string(nil), grp...))
				grp = grp[:0]
				raw = fmt.Sprintf("§%d", len(sub)-1)
				nextIsValue = true
				inEachGroup = false
			}
			if inEachGroup {
				grp = append(grp, raw)
				continue
			}
			if prevIsValue && nextIsValue && !prevIsJoins {
				out = append(out, append([]string(nil), buf...))
				buf = buf[:0]
			}
			buf = append(buf, raw)
			prevIsValue = nextIsValue
			prevIsJoins = nextIsJoins
		}
		if len(buf) > 0 {
			out = append(out, append([]string(nil), buf...))
			buf = buf[:0]
		}
	}
	return out, sub
}

func expand(tokens []string) [][]string {
	out := make([][]string, 0, 8)
	template := make([]string, 0, 8)
	groups := make([][]string, 0, 4)
	gcoefs := make([]int, 0, 4) // mult of prev groups sizes
	gmarks := make([]int, 0, 4) // template indices to substitute
	combos := 1                 // total number of combinations
	fmt.Printf("START SPLITTING\n")
	for i := 0; i < len(tokens); i++ {
		if tokens[i] != "[" {
			template = append(template, tokens[i])
			continue
		}
		group := make([]string, 0, 4)
		template = append(template, "[]")
		for i++; i < len(tokens) && tokens[i] != "]"; i++ {
			group = append(group, tokens[i])
		}
		groups = append(groups, group)
		gcoefs = append(gcoefs, combos)
		gmarks = append(gmarks, len(template)-1)
		combos *= len(group)
	}
	fmt.Printf("T %v\n", template)
	fmt.Printf("G %v\n", groups)
	fmt.Printf("C %v\n", gcoefs)
	fmt.Printf("M %v\n", gmarks)
	for i := range combos {
		exp := append([]string(nil), template...)
		for j, group := range groups {
			idx := (i / gcoefs[j]) % len(group)
			switch group[idx] {
			case "&":
				exp = exp[:0]
				exp = append(exp, "&")
				goto flush
			case ":":
				goto flush
			case "|":
				if idx > 0 {
					exp[gmarks[j]] = group[idx-1]
				} else {
					exp[gmarks[j]] = "0"
				}
			default:
				exp[gmarks[j]] = group[idx]
			}
		}
	flush:
		out = append(out, exp)
	}
	return out
}

// NOT USED
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
