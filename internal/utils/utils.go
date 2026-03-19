package utils

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

