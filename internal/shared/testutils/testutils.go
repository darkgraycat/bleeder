package testutils

import "testing"

func AssertErrNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func AssertInts(t *testing.T, exp, act int) {
	t.Helper()
	if exp != act {
		t.Fatalf("expected: %d\nactual:   %d", exp, act)
	}
}

func AssertFloats(t *testing.T, exp, act float64) {
	t.Helper()
	if exp != act {
		t.Fatalf("expected: %f\nactual:   %f", exp, act)
	}
}

func AssertStrings(t *testing.T, exp, act string) {
	t.Helper()
	if exp != act {
		t.Fatalf("expected: %s\nactual:   %s", exp, act)
	}
}
