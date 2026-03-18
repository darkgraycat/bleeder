package audio

import (
	"bytes"
	"encoding/binary"
	"io"
)

type WAV struct {
	SampleRate int
	Channels   int
	Header     []byte
	Samples    []int16
}

func NewWAV(sampleRate, channels int) *WAV {
	return &WAV{
		SampleRate: sampleRate,
		Channels:   channels,
		Header:     makeWAVHeader(sampleRate, channels, 2),
		Samples:    []int16{},
	}
}

func (wav *WAV) AppendSamples(samples []int16) {
	wav.Samples = append(wav.Samples, samples...)
}

func (wav *WAV) Write(w io.Writer) error {
	dataSize := uint32(len(wav.Samples) * 2)
	fileSize := uint32(36 + dataSize)

	binary.LittleEndian.PutUint32(wav.Header[4:8], fileSize)
	binary.LittleEndian.PutUint32(wav.Header[40:44], dataSize)

	if _, err := w.Write(wav.Header); err != nil {
		return err
	}

	return binary.Write(w, binary.LittleEndian, wav.Samples)
}

func makeWAVHeader(sr, chs, bps int) []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte("RIFF"))                                  // FileTypeBlocID
	buf.Write(make([]byte, 4))                                 // FileSize
	buf.Write([]byte("WAVE"))                                  // FileFormatID
	buf.Write([]byte("fmt "))                                  // FormatBlocID
	binary.Write(buf, binary.LittleEndian, uint32(16))         // BlocSize
	binary.Write(buf, binary.LittleEndian, uint16(1))          // AudioFormat
	binary.Write(buf, binary.LittleEndian, uint16(chs))        // NbrChannels
	binary.Write(buf, binary.LittleEndian, uint32(sr))         // Frequency
	binary.Write(buf, binary.LittleEndian, uint32(sr*chs*bps)) // BytePerSec
	binary.Write(buf, binary.LittleEndian, uint16(chs*bps))    // BytePerBloc
	binary.Write(buf, binary.LittleEndian, uint16(bps*8))      // BitsPerSample
	buf.Write([]byte("data"))                                  // DataBlocID
	buf.Write(make([]byte, 4))                                 // DataSize

	return buf.Bytes()
}
