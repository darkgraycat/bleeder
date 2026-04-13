package player

import (
	"bleeder/internal/ir"
	"fmt"
)

type WAVPlayer struct{}

func NewWAVPlayer() *WAVPlayer {
	return &WAVPlayer{}
}

func (p *WAVPlayer) Play(ir *ir.Program, start, end int) error {
	insts := ir.GetInstructions()
	for i, inst := range insts {
		fmt.Printf("[%d] %f\t- dur: %f, vol: %f, time: %f, info: %s\n",
			i, inst.Freq, inst.Dur, inst.Vol, inst.Time, inst.Info)
	}
	return nil
}

func (p *WAVPlayer) Stop() error {
	return nil
}
