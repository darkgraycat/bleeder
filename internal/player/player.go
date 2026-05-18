package player

import "bleeder/internal/ir"

type Player interface {
	Play(irp *ir.Program, start, end int) error
	Stop() error
}
