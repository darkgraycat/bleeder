package shared

import (
	"strconv"
	"strings"
)

// A ternary expression analog
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

// Map over each element
func Map[T, R any](items []T, fn func(T) R) []R {
	result := make([]R, len(items))

	for i, item := range items {
		result[i] = fn(item)
	}

	return result
}

// Reduce over each element
func Reduce[T, R any](items []T, initial R, fn func(R, T) R) R {
	result := initial

	for _, item := range items {
		result = fn(result, item)
	}

	return result
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

// Convert string to float with default fallback value
func Str2Float(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

// Convert string to int with default fallback value
func Str2Int(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return i
}

// Convert string to bool with default fallback value
func Str2Bool(s string, def bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}

// Convert float to string
func Float2Str(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Convert int to string
func Int2Str(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Convert bool to string
func Bool2Str(b bool) string {
	return strconv.FormatBool(b)
}
