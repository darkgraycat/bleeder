package audio

import "math"

type Phase = float64

const PI2 = 2 * math.Pi

// WaveFunc returns signal value at normalized phase [0..1)
type WaveFunc func(p Phase) float64

// WaveFunc lookup map
var WaveFuncs = map[string]WaveFunc{
	"sine":        WaveSine,
	"square":      WaveSquare,
	"saw":         WaveSaw,
	"triangle":    WaveTriangle,
	"abs-sine":    WaveAbsSine,
	"soft-square": WaveSoftSquare,
	"parabola":    WaveParabola,
	"cubic":       WaveCubic,
}

// f(p) = sin(2πp)
func WaveSine(p Phase) float64 {
	return math.Sin(PI2 * p)
}

// f(p) = 1 if p<0.5 or -1 if p≥0.5
func WaveSquare(p Phase) float64 {
	if p < 0.5 {
		return 1
	}
	return -1
}

// f(p) = 2p-1
func WaveSaw(p Phase) float64 {
	return 2*p - 1
}

// f(p) = 4p-1 if x<0.5 or 3-4p if p≥0.5
func WaveTriangle(p Phase) float64 {
	if p < 0.5 {
		return 4*p - 1
	}
	return 3 - 4*p
}

// f(p) = abs(sin(2πp))
func WaveAbsSine(p Phase) float64 {
	return math.Abs(math.Sin(PI2 * p))
}

// f(p) = tanh(5sin(2πp))
func WaveSoftSquare(p Phase) float64 {
	return math.Tanh(5 * math.Sin(PI2*p))
}

// f(p) = 1-2(2p-1)²
func WaveParabola(p Phase) float64 {
	x := 2*p - 1
	return 1 - 2*x*x
}

// f(p) = (2p-1)³
func WaveCubic(p Phase) float64 {
	x := 2*p - 1
	return x * x * x
}

// Mix multiple wave functions
func WaveFuncMix(waves ...WaveFunc) WaveFunc {
	switch len(waves) {
	case 0:
		return func(Phase) float64 { return 0 }
	case 1:
		return waves[0]
	}

	div := 1.0 / float64(len(waves))
	return func(p Phase) float64 {
		var sum float64
		for _, wave := range waves {
			sum += wave(p)
		}
		return sum * div
	}
}

// Get mixed wave function
func GetWaveFunc(names ...string) WaveFunc {
	waves := make([]WaveFunc, 0, len(names))
	for _, n := range names {
		if w, ok := WaveFuncs[n]; ok {
			waves = append(waves, w)
		}
	}
	return WaveFuncMix(waves...)
}
