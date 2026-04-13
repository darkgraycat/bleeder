package shared

import (
	"strings"
)

// Deprecated: use flag package instead
// TODO: write own parser using 'reflect'
type Args struct {
	Positional []string
	Flags      map[string]string
}

// Deprecated: use flag package instead
func NewArgs(args []string) *Args {
	pos := make([]string, 0)
	flg := make(map[string]string)
	k := ""
	for _, v := range args {
		if strings.Contains(v, "--") {
			k = strings.Replace(v, "--", "", 1)
			flg[k] = ""
		} else if len(k) > 0 {
			flg[k] = v
			k = ""
		} else {
			pos = append(pos, v)
		}
	}
	return &Args{Positional: pos, Flags: flg}
}

func (a *Args) At(position int) string {
	return a.Positional[position]
}

func (a *Args) Get(flag string) string {
	return a.Flags[flag]
}

func (a *Args) Has(flag string) bool {
	return a.Flags[flag] == ""
}

func (a *Args) Length() int {
	return len(a.Positional) + len(a.Flags)
}
