package renderer

import (
	"bleeder/internal/ir"
	"io"
)

// TODO: WAVRenderer, MIDIRenderer, TXTRenderer
type Renderer interface {
	Render(irp *ir.Program, w io.Writer) error
}
