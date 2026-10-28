package core

import (
	"math"
	"slices"
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
			// case "+", "-", "*", "/", "%", "^", "&", "|", ":":
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

func split(raw string) [][]string {
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

func plain(frames [][]string) [][]string {
	out := make([][]string, 0, 8)
	var base, next [][]string // rows before the current group / rows built with it
	var conn []string         // tokens between the previous item and the next one
	inGroup := false

	for _, f := range frames {
		b, e := f[0], f[len(f)-1]
		switch {
		case b == "]" && e == "[":
			base, next = next, nil
			conn = f[1 : len(f)-1]
		case e == "[":
			base, next = [][]string{nil}, nil
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
				next = append(next, slices.Concat(x, conn, f))
			}
		}
	}

	return out
}

// plain3 is plain, but rows are built once: items are only counted while reading,
// and each row (tail included) is created when its group closes.
func plain3(frames [][]string) [][]string {
	out := make([][]string, 0, len(frames))
	var base [][]string // rows before the current group
	var conn []string   // tokens between the previous item and the next one
	start := 0          // index of the current group's first item
	inGroup := false

	for i, f := range frames {
		b, e := f[0], f[len(f)-1]
		if b == "]" {
			items := frames[start:i]
			var tail []string
			if e != "[" {
				tail = f[1:]
			}
			rows := make([][]string, 0, len(base)*len(items))
			for _, item := range items {
				if item[0] == "&" {
					rows = append(rows, item)
					continue
				}
				for _, row := range base {
					if len(row) > 0 && row[0] == "&" {
						rows = append(rows, row)
						continue
					}
					rows = append(rows, slices.Concat(row, conn, item, tail))
				}
			}
			if e == "[" {
				base, conn, start = rows, f[1:len(f)-1], i+1
			} else {
				out = append(out, rows...)
				inGroup = false
			}
			continue
		}
		if e == "[" {
			base, conn, start, inGroup = [][]string{nil}, f[:len(f)-1], i+1, true
			continue
		}
		if !inGroup {
			out = append(out, f)
		}
	}

	return out
}
