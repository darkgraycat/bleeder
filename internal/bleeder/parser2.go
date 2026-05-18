package bleeder

import "strings"

// format lane-DSL into separate lines
func formatLaneContent(content string) []string {
	formatted := replacer.Replace(content)
	if len(formatted) < 1 {
		return nil
	}
	return strings.Split(formatted[1:], opSplitter)
}

// format riff-DSL into separate lines
func formatRiffContent(content string) []string {
	return strings.Fields(content) // TODO
}
