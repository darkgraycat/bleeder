package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// WAV_Old file format interface
type WAV_Old struct {
	sampleRate int     // sample rate
	channels   int     // channels amount
	header     []byte  // premade header bytes
	samples    []int16 // samples bytes
}

// Create new WAV instance
func NewWAV_Old(sampleRate, channels int) *WAV_Old {
	return &WAV_Old{
		sampleRate: sampleRate,
		channels:   channels,
		header:     makeWAVHeader(sampleRate, channels, 2),
		samples:    []int16{},
	}
}

// Write header with known size
func (w *WAV_Old) WriteHeader(numSamples int, wr io.Writer) error {
	dataSize := uint32(numSamples * 2) // samples → bytes
	fileSize := uint32(36 + dataSize)
	header := makeWAVHeader(w.sampleRate, w.channels, 2)
	binary.LittleEndian.PutUint32(header[4:8], fileSize)
	binary.LittleEndian.PutUint32(header[40:44], dataSize)
	_, err := wr.Write(header)
	return err
}

// Write raw samples data
func (w *WAV_Old) WriteSamples(samples []int16, wr io.Writer) error {
	return binary.Write(wr, binary.LittleEndian, samples)
}

// Get sample rate
func (w *WAV_Old) SampleRate() int { return w.sampleRate }

// Get channels
func (w *WAV_Old) Channels() int { return w.channels }

// Get samples
func (w *WAV_Old) Samples() []int16 { return w.samples }

// Deprecated: samples must be generated outside
// Append samples into current WAV
func (w *WAV_Old) Append(samples []int16) {
	w.samples = append(w.samples, samples...)
}

// Deprecated: samples must be generated outside
// Generate tone samples
func (w *WAV_Old) GenerateSamples(freq, dur, vol float64, wave WaveFunc) []int16 {
	numSamples := int(dur * float64(w.sampleRate))
	samples := make([]int16, numSamples)
	phase := 0.0
	step := freq / float64(w.sampleRate)
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

// Deprecated: samples must be generated outside
// Generate tone samples using attack and release
func (w *WAV_Old) GenerateSamplesEnvelope(freq, dur, vol, attackCoef, releaseCoef float64, wave WaveFunc) []int16 {
	numSamples := int(dur * float64(w.sampleRate))
	samples := make([]int16, numSamples)
	phase := 0.0
	step := freq / float64(w.sampleRate)
	amp := vol * math.MaxInt16
	attack := int(float64(w.sampleRate) * attackCoef)
	release := int(float64(w.sampleRate) * releaseCoef)
	attackStep := 1.0 / float64(attack)
	releaseStep := 1.0 / float64(release)
	for i := range samples {
		envelope := 1.0
		if i < attack {
			envelope = attackStep * float64(i)
		} else if i >= numSamples-release {
			envelope = releaseStep * float64(numSamples-i)
		}
		samples[i] = int16(wave(phase) * amp * envelope)
		phase += step
		if phase >= 1 {
			phase -= 1
		}
	}
	return samples
}

// Deprecated: use WriteSamples and WriteHeader instead
// Write WAV data
func (w *WAV_Old) Write(wr io.Writer) error {
	dataSize := uint32(len(w.samples) * 2) // *2 = int16 bytes
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
