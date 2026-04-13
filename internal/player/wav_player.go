package player

import "bleeder/internal/ir"

type WAVPlayer struct{}

func NewWAVPlayer () *WAVPlayer {
	return &WAVPlayer{}
}

func (p *WAVPlayer) Play(ir *ir.Program, start, end int) error {
	return nil
}

func (p *WAVPlayer) Stop() error {
	return nil
}
