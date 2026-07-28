package renderer

import "io"

type Renderer interface {
	Render(r io.Reader, w io.Writer) error
}
