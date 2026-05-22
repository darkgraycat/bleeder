# Bleeder Architecture Proposal

## Pipeline

```
Input (file / stdin / unix socket)
    ↓
LoadBleed()          — bleed.go, pure I/O, resolves includes
    ↓
NewBleeder(bleed)    — bleeder.go, populates lanes/riffs/vibes
    ↓
GenMainIR()          — walks sequences, expands vars, calls parsers
    ↓
ir.Program           — flat array of Instructions with absolute Time
    ↓
Renderer             — WAVRenderer / MIDIRenderer / TextRenderer
    ↓
Output (file / stdout / pipe)
```

---

## Package Responsibilities

### `internal/bleeder`
Owns the DSL → IR transformation. No I/O, no audio.

- `bleed.go` — data structs: Bleed, Meta, Sequence, SequenceType, Vibe
- `bleeder.go` — Bleeder struct: GenMainIR, GenSeqIR, GenIR, var expansion, caching
- `parser.go` — Lane-DSL: ParseContent (operator loop)
- `parser_riff.go` — Riff-DSL: ParseRiff (grid loop) — to be written
- `parser_utils.go` — shared helpers: splitOpArgs, modOpArg, tokenizeLaneContent, extractSequenceVars

### `internal/ir`
Dumb data. No logic beyond structural operations.

- `Program`: flat instruction array + index cache
- `Instruction`: Time, Dur, Vol, Freq, Vibe *Vibe
- Methods: Add, Merge, Copy, Shift, Cut, First, Last, Duration

### `internal/audio`
Pure math. No I/O, no IR knowledge.

- `notes.go` — note/midi/freq conversion
- `wav.go` — WAV sample generation, ADSR envelope
- `waves.go` — waveform functions (sine, saw, square, triangle)
- `synth.go` — Vibe → sample generation (to be written)

### `internal/render`
Consumes IR, produces output.

- `renderer.go` — Renderer interface
- `wav.go` — WAVRenderer: IR → WAV file → stdout or pipe
- `midi.go` — MIDIRenderer: IR → MIDI events (future)
- `text.go` — TextRenderer: IR → tab notation (future)

### `internal/daemon`
Unix socket server for live-coding from Neovim.

- `daemon.go` — listens on socket, receives raw content, calls Bleeder.GenIR, pipes to renderer

### `cmd`
Thin CLI wiring. No DSL logic.

- `cmd.go` — commands: play, listen, send
- `cfg.go` — config loading

---

## IR: Instruction with Vibe

Add Vibe to Instruction:

```go
type Instruction struct {
    Time int
    Dur  int
    Vol  float64
    Freq float64
    Vibe *Vibe   // nil = default sine
}
```

Vibe lives in `internal/ir` as a simple data struct:

```go
type Vibe struct {
    Shape   string  // sine, saw, square, triangle
    Attack  float64 // seconds
    Decay   float64 // seconds
    Sustain float64 // 0.0-1.0 amplitude
    Release float64 // seconds
}
```

Parser tracks `currentVibe *Vibe`, stamps each new instruction. `$bass` switches it.

---

## Renderer Interface

```go
type Renderer interface {
    Render(irp *ir.Program) error
}
```

WAVRenderer writes to `io.Writer` (stdout by default), caller pipes to ffplay or file.
No more hardcoded `afplay`. Unix philosophy: renderer writes bytes, shell handles playback.

```
bleeder play file.bleed | ffplay -i pipe:0
bleeder play file.bleed > out.wav
```

---

## Live-coding Daemon

Unix socket at `/tmp/bleeder.sock`. Neovim sends raw Lane-DSL content, daemon calls `bleeder.GenIR(content)`, renders, pipes to ffplay process kept alive.

```
nvim → :BleederSend → unix socket → daemon → GenIR → WAVRenderer → ffplay
```

No file round-trip. Instant eval.

---

## What to Delete

- `cmd/bleeder.go` — old DSL parser, superseded by `internal/bleeder`
- `cmd/bleed.go` — old data structs, superseded by `internal/bleeder/bleed.go`
- `internal/ir/ir.go` timeScale field — unused
- `cmd/cfg.go` mapping section — operators are now hardcoded in parser2.go

---

## Migration Path

1. Finish GenSeqIR in bleeder.go (var expansion + ParseContent call)
2. Wire GenIR to use b.context (already stubbed)
3. Add Vibe to Instruction, implement $vibe in parser
4. Add Vibe section to bleed.go + LoadBleed
5. Write WAVRenderer using io.Writer instead of afplay
6. Write Renderer interface, wire to cmd/
7. Delete old cmd/bleeder.go and cmd/bleed.go
8. Write Riff parser
9. Write daemon
