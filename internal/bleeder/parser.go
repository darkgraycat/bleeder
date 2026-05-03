package bleeder

import (
	"bleeder/internal/ir"
	"fmt"
	"strings"
)

const (
	opPlay = ">"
	opNote = ":"
	opWave = "~"
	opLink = "@"
	opWait = "_"
	opLast = "|"
)

const opSplitter = "\\"

var replacer = strings.NewReplacer(
	"\n", "", // trim newline
	opPlay, opSplitter+opPlay,
	opNote, opSplitter+opNote,
	opWave, opSplitter+opWave,
	opLink, opSplitter+opLink,
	opWait, opSplitter+opWait,
	opLast, opSplitter+opLast,
)

func ParseRaw(content string, t int) (*ir.Program, error) {
	pr := ir.NewProgram()
	lastOp := opWait
	in := &ir.Instruction{Info: "Start"}
	// create iterator without first empty line with opSplitter
	iterator := strings.SplitSeq(replacer.Replace(content)[2:], opSplitter)
	for raw := range iterator {
		op := string(raw[0])
		args := strings.Fields(raw[1:])
		fmt.Printf("Line %s - %v\n", op, args) // TODO: remove log
		switch op {
		case opPlay:
			in = &ir.Instruction{
				Freq: 0.0,
				Dur:  0.0,
				Vol:  0.0,
				Time: t,
				Info: raw, // for debug
			}
			pr.Add(in)
		case opNote:
			// return nil, fmt.Errorf("not implemented yet: %s", op)
		case opWave:
			// return nil, fmt.Errorf("not implemented yet: %s", op)
		case opLink:
			// return nil, fmt.Errorf("not implemented yet: %s", op)
		case opWait:
			// return nil, fmt.Errorf("not implemented yet: %s", op)
		case opLast:
			// return nil, fmt.Errorf("not implemented yet: %s", op)
		}
		lastOp = op
	}

	fmt.Println(lastOp, in)
	return pr, nil
}
