package core

import "iter"

// lex yields tokens as substrings of raw, with "\n" marking each line end.
// Comments (# to end of line) are skipped.
func lex(raw string) iter.Seq[string] {
	return func(yield func(string) bool) {
		start := -1 // start of the current word, -1 = not inside one
		flush := func(end int) bool {
			if start < 0 {
				return true
			}
			ok := yield(raw[start:end])
			start = -1
			return ok
		}

		for i := 0; i < len(raw); i++ {
			switch raw[i] {
			case ' ', '\t', '\r':
				if !flush(i) {
					return
				}
			case '\n':
				if !flush(i) || !yield("\n") {
					return
				}
			case '#':
				if !flush(i) {
					return
				}
				for i < len(raw) && raw[i] != '\n' {
					i++
				}
				i-- // let the loop see the '\n'
			case '[', ']', '(', ')', '{', '}', '@', '$', '&', '|', ':', '_', '+', '-', '*', '/', '%', '^':
				if !flush(i) || !yield(raw[i:i+1]) {
					return
				}
			default:
				if start < 0 {
					start = i
				}
			}
		}
		flush(len(raw))
	}
}

// arenaDiv estimates raw bytes per token to size the arenas; too small means regrowth, too big wastes memory.
const arenaDiv = 3

// scanNew is scan with a single-pass lexer and shared token arenas:
// frames and groups are capped subslices of two backing slices, not copies.
func scanNew(raw string) (frames [][]string, groups [][]string) {
	frames = make([][]string, 0, 8)
	groups = make([][]string, 0, 4)
	fArena := make([]string, 0, len(raw)/arenaDiv+4)
	gArena := make([]string, 0, len(raw)/arenaDiv+4)
	fStart, gStart := 0, 0

	prevIsValue := false
	prevIsJoins := false
	inEachGroup := false

	for tok := range lex(raw) {
		if tok == "\n" {
			if len(fArena) > fStart {
				frames = append(frames, fArena[fStart:len(fArena):len(fArena)])
				fStart = len(fArena)
			}
			prevIsValue, prevIsJoins, inEachGroup = false, false, false
			continue
		}

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
			groups = append(groups, gArena[gStart:len(gArena):len(gArena)])
			gStart = len(gArena)
			tok = "§"
			nextIsValue = true
			inEachGroup = false
		}

		if inEachGroup {
			gArena = append(gArena, tok)
			continue
		}

		if prevIsValue && nextIsValue && !prevIsJoins {
			frames = append(frames, fArena[fStart:len(fArena):len(fArena)])
			fStart = len(fArena)
		}

		fArena = append(fArena, tok)
		prevIsValue = nextIsValue
		prevIsJoins = nextIsJoins
	}

	if len(fArena) > fStart {
		frames = append(frames, fArena[fStart:len(fArena):len(fArena)])
	}

	return frames, groups
}
