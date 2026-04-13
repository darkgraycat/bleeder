package shared

import (
	"strconv"
	"strings"
)

// An analog of ternary expression
func IfThenElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// Return first match or default value
func Match[K comparable, V any](mapping map[K]V, key K, def V) V {
	if v, ok := mapping[key]; ok {
		return v
	}
	return def
}

// Replace substrings by values from map
func ReplaceByMap(m map[string]string, lines ...string) []string {
	pairs := make([]string, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, k, v)
	}
	r := strings.NewReplacer(pairs...)

	out := make([]string, len(lines))
	for i, s := range lines {
		out[i] = r.Replace(s)
	}
	return out
}

func ParseFloat64(v []string, i int, def float64) float64 {
    if i >= len(v) {
        return def
    }
    f, err := strconv.ParseFloat(v[i], 64)
    if err != nil {
        return def
    }
    return f
}
