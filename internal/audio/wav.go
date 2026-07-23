package audio

import (
	"bytes"
	"encoding/binary"
	"io"
)

// WAV file format interface
type WAV struct {
	sr int // sample rate
	ch int // channels count
}

// Create new WAV instance
func NewWAV(sampleRate, channels int) *WAV {
	return &WAV{
		sr: sampleRate,
		ch: channels,
	}
}

// Write header with known size
func (w *WAV) WriteHeader(wr io.Writer, numSamples int) error {
	bps := 2
	dataSize := numSamples * bps
	fileSize := 36 + dataSize
	buf := new(bytes.Buffer)

	buf.Write([]byte("RIFF"))                                     // FileTypeBlocID
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))      // FileSize
	buf.Write([]byte("WAVE"))                                     // FileFormatID
	buf.Write([]byte("fmt "))                                     // FormatBlocID
	binary.Write(buf, binary.LittleEndian, uint32(16))            // BlocSize
	binary.Write(buf, binary.LittleEndian, uint16(1))             // AudioFormat
	binary.Write(buf, binary.LittleEndian, uint16(w.ch))          // NbrChannels
	binary.Write(buf, binary.LittleEndian, uint32(w.sr))          // Frequency
	binary.Write(buf, binary.LittleEndian, uint32(w.ch*w.sr*bps)) // BytePerSec
	binary.Write(buf, binary.LittleEndian, uint16(w.ch*bps))      // BytePerBloc
	binary.Write(buf, binary.LittleEndian, uint16(8*bps))         // BitsPerSample
	buf.Write([]byte("data"))                                     // DataBlocID
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))      // DataSize

	return binary.Write(wr, binary.LittleEndian, buf.Bytes())
}

// Write raw samples data
func (w *WAV) WriteSamples(wr io.Writer, samples []int16) error {
	return binary.Write(wr, binary.LittleEndian, samples)
}

// Get sample rate
func (w *WAV) SampleRate() int {
	return w.sr
}

// Get channels
func (w *WAV) Channels() int {
	return w.ch
}
