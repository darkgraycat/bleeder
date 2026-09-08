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

// Scan sequence raw content and return templated frames with substitution groups
func scan(raw string) (frames [][]string, groups [][]string) {
	frames = make([][]string, 0, 8)
	groups = make([][]string, 0, 4)
	fBuffer := make([]string, 0, 8)
	gBuffer := make([]string, 0, 4)

	formatted := strings.TrimSpace(formatter.Replace(raw))

	for row := range strings.SplitSeq(formatted, "\n") {
		if i := strings.IndexByte(row, '#'); i >= 0 {
			row = row[:i]
		}

		prevIsValue := false
		prevIsJoins := false
		inEachGroup := false

		for tok := range strings.FieldsSeq(row) {
			nextIsValue := true
			nextIsJoins := prevIsJoins

			switch tok {
			case "+", "-", "*", "/", "%", "^", "|", ":":
				nextIsValue = false
			case "(", "{", "@", "$":
				nextIsJoins = true
			case ")", "}":
				nextIsJoins = false
			case "[":
				inEachGroup = true
				continue
			case "]":
				groups = append(groups, append([]string(nil), gBuffer...))
				gBuffer = gBuffer[:0]
				tok = "§"
				nextIsValue = true
				inEachGroup = false
			}

			if inEachGroup {
				gBuffer = append(gBuffer, tok)
				continue
			}

			if prevIsValue && nextIsValue && !prevIsJoins {
				frames = append(frames, append([]string(nil), fBuffer...))
				fBuffer = fBuffer[:0]
			}

			fBuffer = append(fBuffer, tok)
			prevIsValue = nextIsValue
			prevIsJoins = nextIsJoins
		}

		if len(fBuffer) > 0 {
			frames = append(frames, append([]string(nil), fBuffer...))
			fBuffer = fBuffer[:0]
		}
	}

	return frames, groups
}

// Flat templated frames into expressions using substitution groups
func flat(frames [][]string, groups [][]string) (expressions [][]string) {
	if len(groups) == 0 {
		return frames
	}

	expressions = make([][]string, 0, len(frames)*4)
	groupOffset := 0

	for _, frame := range frames {
		subsTotal := 0
		for _, tok := range frame {
			if tok == "§" {
				subsTotal++
			}
		}

		if subsTotal == 0 {
			expressions = append(expressions, frame)
			continue
		}

		subsIndices := make([]int, subsTotal)
		appendFlags := make([]bool, subsTotal)
		template := make([]string, 0, len(frame))

	build:
		template = template[:0]
		subIndex := 0

		for _, tok := range frame {
			if tok != "§" {
				template = append(template, tok)
				continue
			}

			val := groups[groupOffset+subIndex][subsIndices[subIndex]]
			switch val {
			case "&":
				expressions = append(expressions, []string{"&"})
				appendFlags[subIndex] = false
				goto next
			case "+", "-", "*", "/", "%", "^", "|", ":":
				last := len(expressions) - 1
				expressions[last] = append(expressions[last], val)
				appendFlags[subIndex] = true
				goto next
			default:
				if appendFlags[subIndex] {
					last := len(expressions) - 1
					expressions[last] = append(expressions[last], val)
					appendFlags[subIndex] = false
					goto next
				}
				template = append(template, val)
				subIndex++
			}
		}
		expressions = append(expressions, append([]string(nil), template...))

	next:
		for subIndex := range subsTotal {
			subsIndices[subIndex]++
			if subsIndices[subIndex] < len(groups[groupOffset+subIndex]) {
				goto build
			}
			subsIndices[subIndex] = 0
		}
		groupOffset += subsTotal
	}

	return expressions
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
