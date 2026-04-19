package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// WAV file format interface
type WAV struct {
	sampleRate float64 // sample rate
	channels   int     // channels amount
	header     []byte  // premade header bytes
	samples    []int16 // samples bytes
}

// Create new WAV instance
func NewWAV(sampleRate, channels int) *WAV {
	return &WAV{
		sampleRate: float64(sampleRate),
		channels:   channels,
		header:     makeWAVHeader(sampleRate, channels, 2),
		samples:    []int16{},
	}
}

// Get sample rate
func (w *WAV) SampleRate() float64 { return w.sampleRate }

// Get channels
func (w *WAV) Channels() int { return w.channels }

// Get samples
func (w *WAV) Samples() []int16 { return  w.samples }

// Append samples into current WAV
func (w *WAV) Append(samples []int16) {
	w.samples = append(w.samples, samples...)
}

// Generate tone samples
func (w *WAV) GenerateSamples(freq, dur, vol float64, wave WaveFunc) []int16 {
	numSamples := int(dur * w.sampleRate)
	samples := make([]int16, numSamples)
	phase := 0.0
	step := freq / w.sampleRate
	amp := vol * math.MaxInt16
	for i := range samples {
		samples[i] = int16(wave(phase) * amp)
		phase += step
		if phase >= 1 {
			phase -= 1
		}
	}
	return samples
}

// Write WAV data
func (w *WAV) Write(wr io.Writer) error {
	dataSize := uint32(len(w.samples) * 2)
	fileSize := uint32(36 + dataSize)
	binary.LittleEndian.PutUint32(w.header[4:8], fileSize)
	binary.LittleEndian.PutUint32(w.header[40:44], dataSize)
	if _, err := wr.Write(w.header); err != nil {
		return err
	}
	return binary.Write(wr, binary.LittleEndian, w.samples)
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
