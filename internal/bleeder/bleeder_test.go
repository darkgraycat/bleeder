package bleeder

import (
	"bleeder/internal/shared/logs"
	"bleeder/internal/shared/testutils"
	"fmt"
	"testing"
)

func TestGenLaneIR(t *testing.T) {
	tests := []struct {
		name     string
		given    [][]string
		expected []string
		errMsg   string
	}{
		{
			name:  "simple sequential melody",
			given: [][]string{{">40", ">80", ">c4"}},
			expected: []string{
				"m40.0 v1.0 d1 t0",
				"m80.0 v1.0 d1 t1",
				"m60.0 v1.0 d1 t2",
			},
		},
		{
			name: "chord after chord",
			given: [][]string{
				{">40", "|", ">80"},
				{">c4", "|", ">d4"},
			},
			expected: []string{
				"m40.0 v1.0 d1 t0",
				"m80.0 v1.0 d1 t0",
				"m60.0 v1.0 d1 t1",
				"m62.0 v1.0 d1 t1",
			},
		},
		{
			name: "time manipulations",
			given: [][]string{
				{">40", ">60:2", "|", "_", ">80"},
			},
			expected: []string{
				"m40.0 v1.0 d1 t0",
				"m60.0 v1.0 d2 t1",
				"m80.0 v1.0 d1 t2",
			},
		},
		{
			name: "repeat previos",
			given: [][]string{
				{">40:3", "<+8:-1"},
				{">40", "|", "<60+8"},
			},
			expected: []string{
				"m40.0 v1.0 d3 t0",
				"m48.0 v1.0 d2 t3",
				"m40.0 v1.0 d1 t5",
				"m68.0 v1.0 d1 t5",
			},
		},
	}

	b := NewBleeder(&Bleed{})
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutils.CheckFlags(t)
			irp, err := b.genLaneIR(tc.given)

			testutils.AssertErr(t, err, tc.errMsg)
			testutils.AssertInts(t, len(tc.expected), irp.Length())

			for i, ins := range irp.Instructions() {
				act := fmt.Sprintf("m%.1f v%.1f d%d t%d", ins.Midi, ins.Vol, ins.Dur, ins.Time)
				exp := tc.expected[i]
				testutils.AssertStrings(t, exp, act)
			}
		})
	}
}

func BenchmarkGenLaneIR(b *testing.B) {
	tests := []struct {
		given [][]string
	}{
		{
			given: [][]string{{">40", ">80", ">c4"}},
		},
		{
			given: [][]string{
				{">40", "|", ">80"},
				{">c4", "|", ">d4"},
			},
		},
		{
			given: [][]string{
				{">40", ">60:2", "|", "_", ">80"},
			},
		},
		{
			given: [][]string{
				{">40:3", "<+8:-1"},
				{">40", "|", "<60+8"},
			},
		},
	}

	logLevel := logs.GetLogLevel()
	logs.SetLogLevel(logs.DISABLED)
	bl := NewBleeder(&Bleed{})
	for i, tc := range tests {
		b.Run(fmt.Sprintf("case%d", i), func(b *testing.B) {
			for b.Loop() {
				bl.genLaneIR(tc.given)
			}
		})
	}
	logs.SetLogLevel(logLevel)
}
