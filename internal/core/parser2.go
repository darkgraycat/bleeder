package core

import (
	"math"
	"strconv"
	"strings"
)

var tokenizeReplacer = strings.NewReplacer(
	"[", " [ ", "]", " ] ", // for each group
	"(", " ( ", ")", " ) ", // positional arguments
	"{", " { ", "}", " } ", // named arguments
	"@", " @", ":", " : ", "&", " & ", "|", " | ", "_", " _ ",
	"+", " + ", "-", " -", "*", " * ", "/", " / ", "%", " % ", "^", " ^ ",
)

func split(content string) [][]string {
	out := make([][]string, 0, 8)
	mem := make([]string, 0, 16)
	src := strings.TrimSpace(tokenizeReplacer.Replace(content))
	for row := range strings.SplitSeq(src, "\n") {
		if i := strings.IndexByte(row, '#'); i >= 0 {
			row = row[:i] // skip rest of the line
		}
		prevIsValue := false // previos is raw value
		isJoinGroup := false // join into single group
		for token := range strings.FieldsSeq(row) {
			currIsValue := strings.IndexByte("+-*/%^", token[0]) < 0
			if prevIsValue && currIsValue && !isJoinGroup {
				out = append(out, append([]string(nil), mem...))
				mem = mem[:0]
			}
			switch token[0] {
			case '[', '(', '{', '@':
				isJoinGroup = true
			case ']', ')', '}':
				isJoinGroup = false
			}
			mem = append(mem, token)
			prevIsValue = currIsValue
		}
		if len(mem) > 0 {
			out = append(out, append([]string(nil), mem...))
			mem = mem[:0]
		}
	}
	return out
}

func expand(tokens []string) [][]string {
	out := make([][]string, 0, 8)
	template := make([]string, 0, 8)
	groups := make([][]string, 0, 4)
	gcoefs := make([]int, 0, 4) // mult of prev groups sizes
	gmarks := make([]int, 0, 4) // template indices to substitute
	combos := 1                 // total number of combinations
	for i := 0; i < len(tokens); i++ {
		if tokens[i] != "[" {
			template = append(template, tokens[i])
			continue
		}
		group := make([]string, 0, 4)
		template = append(template, "[]")
		for i++; tokens[i] != "]"; i++ {
			group = append(group, tokens[i])
		}
		groups = append(groups, group)
		gcoefs = append(gcoefs, combos)
		gmarks = append(gmarks, len(template)-1)
		combos *= len(group)
	}
	for i := range combos {
		exp := append([]string(nil), template...)
		for j, group := range groups {
			idx := (i / gcoefs[j]) % len(group)
			switch group[idx] {
			case "&":
				exp = exp[:0]
				exp = append(exp, "&")
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
