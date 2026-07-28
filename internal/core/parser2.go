package core

import "strings"

// chars:
// [] - group tones or sequences
// () - send positional arguments
// {} - send named arguments
// _  - rest
// |  - prev
// @  - link
// &  - with
// #  - skip

var tokenizeReplacer = strings.NewReplacer(
	"[", " [", "]", "] ",
	"(", " (", ")", ") ",
	"{", " {", "}", "} ",
	"@", " @", "&", " & ",
	"|", " |", "_", " _ ",

	// chPlay, " "+chPlay,
	// chPrev, " "+chPrev,
	// chLink, " "+chLink,
	// chVibe, " "+chVibe,
	// chRest, " "+chRest,
	// chWith, " "+chWith,
	// chSkip, " "+chSkip,
)

func tokenize(s string) [][]string {
	out := make([][]string, 0, 4)
	src := strings.TrimSpace(tokenizeReplacer.Replace(s))
	for row := range strings.SplitSeq(src, "\n") {
		if i := strings.IndexByte(row, '#'); i >= 0 {
			row = row[:i]
		}
		if fields := strings.Fields(row); len(fields) > 0 {
			out = append(out, fields)
		}
	}
	return out
}
