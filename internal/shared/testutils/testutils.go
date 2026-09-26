package testutils

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

func AssertErr(t *testing.T, err error, expMsg string) {
	t.Helper()
	if expMsg == "" {
		if err != nil {
			t.Fatalf("\nexpected no error, got: %v", err)
		}
	} else if err == nil || err.Error() != expMsg {
		t.Fatalf("\nexpected error: %q, got: %v", expMsg, err)
	}
}

func AssertBools(t *testing.T, exp, act bool) {
	t.Helper()
	if exp != act {
		t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
	}
}

func AssertInts(t *testing.T, exp, act int) {
	t.Helper()
	if exp != act {
		t.Fatalf("\nexpected: `%d`\nactual:   `%d`", exp, act)
	}
}

func AssertFloats(t *testing.T, exp, act float64) {
	t.Helper()
	if exp != act {
		t.Fatalf("\nexpected: `%g`\nactual:   `%g`", exp, act)
	}
}

func AssertStrings(t *testing.T, exp, act string) {
	t.Helper()
	if exp != act {
		t.Fatalf("\nexpected: %q (%d)\nactual:   %q (%d)", exp, len(exp), act, len(act))
	}
}

func AssertSlices[T comparable](t *testing.T, exp, act []T) {
	t.Helper()
	for i := range min(len(exp), len(act)) {
		if exp[i] != act[i] {
			t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
		}
	}
	if len(exp) != len(act) {
		t.Fatalf("\nexpected: `%v` (%d)\nactual:   `%v` (%d)", exp, len(exp), act, len(act))
	}
}

func AssertNested[T comparable](t *testing.T, exp, act [][]T) {
	t.Helper()
	for i := range min(len(exp), len(act)) {
		if !slices.Equal(exp[i], act[i]) {
			t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp[i], act[i])
		}
	}
	if len(exp) != len(act) {
		t.Fatalf("\nexpected: `%v` (%d)\nactual:   `%v` (%d)", exp, len(exp), act, len(act))
	}
}

func AssertMaps[K, V comparable](t *testing.T, exp, act map[K]V) {
	t.Helper()
	if !maps.Equal(exp, act) {
		t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
	}
}

func UseFlags(t *testing.T) {
	t.Helper()
	parts := strings.Split(t.Name(), "/")
	name := parts[len(parts)-1]
	if name == "" {
		return
	}
	switch name[0] {
	case '-':
		t.Skip("flagged as skip")
	case '!':
		t.Fatal("flagged as fail")
	case '|':
		t.Parallel()
	}
}
