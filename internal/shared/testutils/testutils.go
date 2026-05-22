package testutils

import "testing"

func AssertErrNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("\nerror: `%v`", err)
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
	}}
