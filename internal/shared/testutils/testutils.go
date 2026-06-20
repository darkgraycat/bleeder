package testutils

import (
	"strings"
	"testing"
)

func AssertErr(t *testing.T, err error, expMsg string) {
	t.Helper()
	if expMsg == "" {
		if err != nil {
			t.Fatalf("\nexpected no error, got: %v", err)
		}
	} else {
		if err == nil || err.Error() != expMsg {
			t.Fatalf("\nexpected error: %q, got: %v", expMsg, err)
		}
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
		t.Fatalf("\nexpected: `%f`\nactual:   `%f`", exp, act)
	}
}

func AssertStrings(t *testing.T, exp, act string) {
	t.Helper()
	if exp != act {
		t.Fatalf("\nexpected: `%s`\nactual:   `%s`", exp, act)
	}
}

func AssertSlices[T comparable](t *testing.T, exp, act []T) {
	t.Helper()
	if len(exp) != len(act) {
		t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
	}
	for i, v := range act {
		if v != exp[i] {
			t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
		}
	}
}

func AssertMaps[T comparable, K comparable](t *testing.T, exp, act map[K]T) {
	t.Helper()
	if len(exp) != len(act) {
		t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
	}
	for k, v := range act {
		if v != exp[k] {
			t.Fatalf("\nexpected: `%v`\nactual:   `%v`", exp, act)
		}
	}
}

func CheckFlags(t *testing.T) {
	t.Helper()
	parts := strings.Split(t.Name(), "/")
	name := parts[len(parts)-1]
	if len(name) < 1 {
		return
	}
	switch name[0] {
	case '-':
		t.SkipNow()
	case '!':
		t.FailNow()
	case '|':
		t.Parallel()
	}
}
