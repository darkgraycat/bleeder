package audio

import "math"

type Phase = float64

// WaveFunc returns signal value at normalized phase [0..1)
type WaveFunc func(p Phase) float64

// f(x) = sin(2πx)
func WaveSine(p Phase) float64 {
	return math.Sin(2 * math.Pi * p)
}

// f(x) =  1 if x < 0.5
//
//	-1 if x ≥ 0.5
func WaveSquare(p Phase) float64 {
	if p < 0.5 {
		return 1
	}
	return -1
}

// f(x) = 2x - 1
func WaveSaw(p Phase) float64 {
	return 2*p - 1
}

// f(x) = 4x - 1 if x < 0.5
//
//	3 - 4x if x ≥ 0.5
func WaveTriangle(p Phase) float64 {
	if p < 0.5 {
		return 4*p - 1
	}
	return 3 - 4*p
}
