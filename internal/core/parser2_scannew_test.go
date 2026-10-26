package core

import (
	"strings"
	"testing"
)

var (
	sinkFrames [][]string
	sinkGroups [][]string
)

const scanCmpMixed = `
[a b] + [c d]
e2 2 + [a b c] + [0 1] d3
[x y] * 2 a3
`

const scanCmpComments = `
# eeee aaaa ff ff
e2:4 a2:4 f2:2|2
# run bass intro
$bass(0.77 12)
@intro{vol 0.5 tone g3}
`

// Compare: go test ./internal/core -run '^$' -bench ScanCompare -benchmem -count=10 | tee scan.txt
// then: benchstat -col /impl scan.txt
func BenchmarkScanCompare(b *testing.B) {
	cases := []struct {
		name  string
		given string
	}{
		{"mixed", scanCmpMixed},
		{"comments", scanCmpComments},
		{"large", strings.Repeat(scanCmpMixed+scanCmpComments, 100)},
	}

	for _, c := range cases {
		b.Run("case="+c.name+"/impl=old", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkFrames, sinkGroups = scan(c.given)
			}
		})
		b.Run("case="+c.name+"/impl=new", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkFrames, sinkGroups = scanNew(c.given)
			}
		})
	}
}
