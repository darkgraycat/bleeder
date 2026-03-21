package player

import "bleeder/internal/ir"

type Player interface {
	Play(ir *ir.Program, command int)
	Stop()
}

type WAVPlayer struct{}

func (p *WAVPlayer) Play(ir *ir.Program, command int) {}
func (p *WAVPlayer) Stop()                            {}
