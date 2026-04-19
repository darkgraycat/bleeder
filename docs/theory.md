## Understanding how sound can be synthesized
There a 3 attributes:
- Amplitude
- Frequency
- Timbre/Shape

## WAV samples represents
`samples[i]` represents signal value at time `t = i / sampleRate`

## WAV `phase` is
In my normalized version - `phase ∈ [0, 1)`
*horizontal position on the wave*
Not time itself, but position inside repeating shape.
```go
step := freq / sr
phase += step
```
freq = how many cycles per second (Hz)
sr = how many samples per second
step = cycles per sample

In old radians version - `phase ∈ [0 .. 2π)`
radians = phase * 2π

So. Phase is:
phase = fractional progress through one cycle of the wave
