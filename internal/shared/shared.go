package shared

import (
	// "reflect"
	"strings"
)

func IfThenElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func Match[K comparable, V any](mapping map[K]V, key K, def V) V {
	if v, ok := mapping[key]; ok {
		return v
	}
	return def
}

// TODO: implement decoder to have shape and validation of it
func DecodeArgs(a []string, dst any) error {
	// v := reflect.ValueOf(dst)
	return nil
}

func ReplaceByMap(replacements map[string]string, lines ...string) []string {
	pairs := make([]string, 0, len(replacements)*2)
	for k, v := range replacements {
		pairs = append(pairs, k, v)
	}
	replacer := strings.NewReplacer(pairs...)
	replaced := make([]string, len(lines))
	for i, line := range lines {
		replaced[i] = replacer.Replace(line)
	}
	return replaced
}
