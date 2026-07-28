package renderer

import (
	"bleeder/internal/audio"
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

func (wr *WAVRenderer) Render(r io.Reader, w io.Writer) error {
	return nil
}
