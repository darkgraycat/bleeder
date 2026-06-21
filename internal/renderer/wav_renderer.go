package renderer

import (
	"bleeder/internal/audio"
	"bleeder/internal/ir"
	"io"
)

type WAVRenderer struct {
	wav *audio.WAV
}

func NewWAVRenderer(sampleRate, channels int) *WAVRenderer {
	return &WAVRenderer{
		wav: audio.NewWAV(sampleRate, channels),
	}
}

func (wr *WAVRenderer) Render(irp *ir.Program, w io.Writer) {

}
