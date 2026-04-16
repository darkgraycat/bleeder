package player

import "bleeder/internal/ir"

type Player interface {
	Play(pr *ir.Program, start, end int) error
	Stop() error
}
