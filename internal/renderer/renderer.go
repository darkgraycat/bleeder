package renderer

import (
	"bleeder/internal/ir"
	"io"
)

type Renderer interface {
	Render(irp *ir.Program, w io.Writer) error
}
