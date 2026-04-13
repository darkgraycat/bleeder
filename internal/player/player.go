package player

import "bleeder/internal/ir"

type Player interface {
	Play(ir *ir.Program, start, end int) error
	Stop() error
}
