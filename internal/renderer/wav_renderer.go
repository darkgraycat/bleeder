package renderer

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"io"
)

type WAVRenderer struct {
	wav *audio.WAV_Old
}

func NewWAVRenderer(sampleRate, channels int) *WAVRenderer {
	return &WAVRenderer{
		wav: audio.NewWAV_Old(sampleRate, channels),
	}
}

func (r *WAVRenderer) Render(irp *ir.Program, w io.Writer) error {
	return nil
}
