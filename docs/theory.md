# Understanding how sound can be synthesized
There a 3 main attributes:
- Amplitude
- Frequency
- Timbre/Shape

## ADSR
ADSR = how a sound changes over time after you press/release a note.

- A — Attack
How fast sound reaches full volume after pressing a key.
short attack = instant hit (drum, pluck)
long attack = fade in (pad, ambient)

- D — Decay
After peak volume, how fast it drops to sustain level.
Like: “initial punch disappears”.

- S — Sustain
Volume level while you keep holding the note.
Not time — level.
high sustain = organ-like constant sound
low sustain = plucky sound

- R — Release
How long sound fades out after releasing the key.
short release = abrupt stop
long release = lingering tail/reverb-ish feel

---

# WAV in Bleeder

WAV samples represents
`samples[i]` represents signal value at time `t = i / sampleRate`

WAV `phase` is
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

