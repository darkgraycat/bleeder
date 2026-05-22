# Audio Generation — Notes

## What is a sample?

A sample is just a number — one moment of audio amplitude, typically `int16` (-32768 to 32767).
A buffer of samples at a given rate (44100/sec) = audio you can hear.

Pre-recorded samples (WAV files) are just buffers of these numbers saved to disk.
FLStudio, Tidal, SonicPi all use pre-recorded WAV files as sound sources.
No real-time synthesis — the sound was generated once, saved, replayed at the right time.

---

## Sound sources

Two ways to produce a `[]int16` buffer:

**Generated tone** — math produces waveform samples in real-time:
- Sine — pure tone, no harmonics
- Saw — rich harmonics, buzzy (good for bass/lead)
- Square — hollow, reedy
- Triangle — softer than saw

**Sample** — load a WAV file, decode into `[]int16`, play it back.
Pitch shifting = play back faster or slower.

Both produce the same thing: `[]int16`. Downstream processing is identical.

---

## ADSR

Shapes volume over time. Applied by multiplying each sample value by the envelope curve.

```
amplitude
    1 │   ╱╲
      │  ╱  ╲____
      │ ╱        ╲
    0 │╱          ╲
      └─────────────── time
        A  D  S  R
```

- **Attack** — time from 0 to peak (seconds)
- **Decay** — time from peak to sustain level (seconds)
- **Sustain** — amplitude held during note (0.0–1.0)
- **Release** — time from sustain to 0 after note ends (seconds)

ADSR works on both generated tones and samples — it's just math on `[]int16`.

---

## Filters

Applied after waveform generation. Cuts or boosts frequency ranges.

- **Low-pass** — lets low frequencies through, cuts highs. Warm/muffled sound.
- **High-pass** — lets highs through, cuts lows. Thin/bright sound.
- **Band-pass** — lets a frequency band through. Nasal/telephone sound.

Implementation: convolution with a filter kernel. Future feature for bleeder.

---

## Synthesis pipeline

```
Sound source (generated or sample)
    ↓ []int16
ADSR envelope applied
    ↓ []int16
Filter applied (future)
    ↓ []int16
Mix with other instructions
    ↓ []int16
Soft-clip (tanh to prevent overflow)
    ↓ []int16
Output (WAV / ffplay stdin)
```

---

## Vibe — sound shape definition

Vibe separates *what the sound is* from *when it plays*.
IR carries Freq, Time, Dur, Vol — the musical intent.
Vibe carries Sound + ADSR + Filter — the timbral character.

```go
type Sound interface {
    Generate(freq float64, dur int) []int16
}

type ToneSound struct{ Shape WaveFunc }   // generated waveform
type SampleSound struct{ data []int16 }   // pre-loaded WAV

type Vibe struct {
    Sound   Sound
    Attack  float64
    Decay   float64
    Sustain float64
    Release float64
    // Filter future
}
```

Renderer: `vibe.Apply(sound.Generate(freq, dur))` → `[]int16` ready for mixing.

---

## DSL operators (updated)

`~` (raw frequency) — dropped. Redundant.
`>` handles both integer and fractional MIDI: `>60` = C4, `>60.5` = quarter tone above C4.
Microtonal via fractional MIDI is cleaner than raw Hz values.

Final operators:
- `>` midi (integer or fractional)
- `:` note name
- `_` wait
- `|` repeat/modify last
- `@` link sequence
- `$` switch vibe
