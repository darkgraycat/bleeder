package bleeder

import (
	"bleeder/internal/ir"
	"fmt"
	"strings"
)

const (
	OpWait = iota
	OpPlay
	OpWave
	OpSeq
	OpRepeat
	OpRepeatLine
)

func ParseSequence(content string) (*ir.Program, error) {
	lines := strings.Split(content, "\n")
	lastOp := OpWait
	for _, line := range lines {
		fmt.Printf("Line %s'n", line)
		fmt.Printf("Last %s'n", lastOp)
		for part := range strings.FieldsSeq(line) {
			switch part[0] {
			case '>':

			}

		}
	}

	return nil, nil
}
