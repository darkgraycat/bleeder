package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var tokenizeReplacer = strings.NewReplacer(
	"[", " [ ", "]", " ] ", // for each group
	"(", " ( ", ")", " ) ", // positional arguments
	"{", " { ", "}", " } ", // named arguments
	"@", " @", ":", " : ", "&", " & ", "|", " | ", "_", " _ ",
	"+", " + ", "-", " -", "*", " * ", "/", " / ", "%", " % ",
)

func tokenize(content string) [][]string {
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
			currIsValue := strings.IndexByte("+-*/%&", token[0]) < 0
			if prevIsValue && currIsValue && !isJoinGroup {
				out = append(out, append([]string(nil), mem...))
				mem = mem[:0]
			}
			switch token[0] {
			case '@', '[', '(', '{':
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

// TODO: do tokenize4 - where we could split into expressions BEFORE .Fields()

// currently doing: expansion of []
func devSyntax(tokens []string) []string {
	out := make([]string, 0, 16)
	mem := make([]string, 0, 8)
	isExpr := true

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		fmt.Printf("[%d] %q\n", i, token)

		switch token {
		case "[":
			i++
			for i < len(tokens) && tokens[i] != "]" {
				fmt.Printf("[%d] GROUP %q\n", i, tokens[i])
				product := append(mem, tokens[i])
				out = append(out, product...) // bug - it flushes
				i++
			}
		case "+", "-", "*", "/", "%":
			isExpr = true
			mem = append(mem, token)
		default:
			if isExpr {
				mem = append(mem, token)
			} else {
				out = append(out, token)
				mem = mem[:0] // clear mem
			}
			isExpr = false
		}
	}

	return out
}

func devSyntax2(tokens []string) []string {
	out := make([]string, 0, 16) // final output

	var groups [][]string
	var template []string

	// build template and collect substitution groups
	for i := 0; i < len(tokens); i++ {
		if tokens[i] != "[" {
			template = append(template, tokens[i])
			continue
		}
		start := i + 1
		for i++; tokens[i] != "]" && i < len(tokens); i++ {
		}
		groups = append(groups, tokens[start:i])
		template = append(template, fmt.Sprintf("$%d", len(groups)))

	}

	// in case no [] group were found
	if len(groups) == 0 {
		return tokens
	}

	total := 1
	for _, items := range groups {
		total *= len(items)
	}

	// for comboIdx := 0; comboIdx < totalCombos; comboIdx++ {
	// 	remainder := comboIdx
	// 	var combo [][]string
	//
	// 	for _, items := range splitGroups {
	// 		itemIdx := remainder % len(items)
	// 		combo = append(combo, items[itemIdx])
	// 		remainder /= len(items)
	// 	}
	//
	// 	combos = append(combos, combo)
	// }

	// for i := len(groups) - 1; i >= 0; i-- {
	// }

	fmt.Printf("TOK \t%v\n", strings.Join(tokens, " "))
	fmt.Printf("GR \t%v\n", groups)
	fmt.Printf("TEMP \t%v\n", strings.Join(template, " "))

	fmt.Printf("TOT COMB \t%v\n", total)
	return out
}

// func devSyntax3(tokens []string) []string {
// 	out := make([]string, 0, 16)  // final output
// 	mem := make([][]string, 0, 8) // collected groups
// 	isExpr := false
//
// 	for i := 0; i < len(tokens); i++ {
// 		token := tokens[i]
// 		switch token {
// 		case "+", "-", "*", "/", "%":
// 		default:
// 			if isExpr {
//
// 			} else {
// 				isExpr = true // start new expression
// 				mem = append(mem, []string{token})
// 			}
// 		}
// 	}
//
// 	return out
// }

/*
Question:
what if we can write "expand" (or another name) function that accepts a dictionary of variables
to substitute them in "tokens" and to expand groups?
Idea is simple - do everything possible while it is just a text.
Then produce simplified output for IR generator
*/

// NOT USED
// evaluate arithmetic expression with variables map
func evalVars2(s string, vars map[string]string) float64 {
	i := strings.LastIndexAny(s, "+-*/%")
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
