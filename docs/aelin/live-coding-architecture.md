# Live-Coding Architecture

## Core Insight

The sequence tree is an **authoring abstraction** — how you write music.
At play time it collapses into a **flat timeline** — one dimension, left to right.
Like MIDI. Like IR.

Tree only exists at parse time. Playback never sees it.

---

## Two Phases

### Parse phase (tree)
```
bleed.toml
└── main
    ├── @seq1
    │   ├── @seq1_1
    │   └── @seq1_2
    └── @seq2
```
Bleeder resolves the tree into a flat IR using the buffer-drain algorithm.
Result: one sorted slice of Instructions with **relative Time values** (each phrase starts at t=0).

### Play phase (flat)
```
[seq1_1 instructions | seq1_2 instructions | seq2 instructions]
 ↑                    ↑                     ↑
 boundary             boundary              boundary
```
Renderer walks the flat IR left to right. No tree. No nesting. No cursors.
Boundaries are just metadata — slice positions marking where each phrase ends.

---

## Flat IR Structure

Each phrase is its own self-contained IR block.
Instructions are sorted and **delta-encoded** (MIDI-style): each `Time` = delta from previous instruction.
Renderer maintains its own `currentTime` cursor and accumulates deltas to derive absolute time.

```go
type Phrase struct {
    Name         string            // sequence name e.g. "seq1"
    Instructions []*ir.Instruction // sorted, delta Time (each = delta from previous instruction)
    Duration     int               // total ticks, used to advance renderer cursor
}

type FlatIR struct {
    Phrases []*Phrase  // top-level phrases in play order
}
```

Example:
```
FlatIR.Phrases:
  Phrase{seq1, [i0..i3], dur=48}
  Phrase{seq2, [i0..i7], dur=96}
```

Renderer:
```
cursor = 0
play seq1 → t=0; for ins in seq1: t += ins.Time; emit at cursor+t → cursor += seq1.Duration
play seq2 → t=0; for ins in seq2: t += ins.Time; emit at cursor+t → cursor += seq2.Duration
```

Hot-swap: replace `Phrase` in slice. No shifting, no Time recalculation. O(1).

---

## Hot-Swap

On file save → bleeder detects changed sequences.

1. Re-parse changed sequence → new instruction block (microseconds)
2. Wait for current phrase boundary to be reached by renderer
3. Replace old instruction block with new one in FlatIR
4. Continue playback — no restart, no gap

```
Renderer playing: [seq1_block | seq2_block | ...]
                              ↑
                        boundary reached
                        → swap seq2_block with updated version
                        → continue
```

### Cache invalidation
Each sequence cached after first parse.
On change: mark changed sequence as dirty + all sequences that reference it.
Only dirty sequences re-parse. Clean sequences reuse cached IR block.

---

## Listen Mode — User Experience

```
bleeder --listen bleed.toml
```

1. Starts up, parses everything, loads vibes. Silence.
2. User opens bleed.toml in editor.
3. User triggers play (via command or file-save hook).
4. Bleeder plays main sequence (or specific sequence by name).
5. User edits a sequence, saves.
6. Current phrase finishes → hot-swap → updated version plays.
7. Loop continues forever or until user stops.

### Trigger mechanisms
- File watcher — polls for changes, no editor plugin needed
- Unix socket — Neovim sends signal on save, instant reaction
- Both supported, socket preferred for live performance

---

## Phrase Boundary Rule

**Simple rule: finish current top-level phrase, then hot-swap.**

If playback is inside `@seq1` and user edits `@seq1_1` (nested inside):
→ `@seq1` finishes entirely (including all nested sequences)
→ bleeder re-parses `@seq1` (which picks up new `@seq1_1`)
→ updated `@seq1` plays next iteration

Why top-level only:
- No tree traversal at playback time
- One simple rule, easy to reason about
- Feels musical — changes land on phrase boundaries
- Keep phrases short (4-8 bars) → low latency

---

## Double-Buffer

```
Renderer goroutine:  playing FlatIR_current
Bleeder goroutine:   preparing FlatIR_next (re-parsing changed sequences)

boundary hit → atomic swap → renderer picks up FlatIR_next
               bleeder starts preparing FlatIR_next+1
```

Memory: two flat IRs in memory at once. Bounded, small.

---

## Performance

| Operation | Cost |
|-----------|------|
| Parse sequence → IR block | microseconds (cached after first) |
| Buffer-drain ordering | O(N log B), B = buffer size ≈ tiny |
| Hot-swap (slice replacement) | O(1) |
| Cache invalidation | O(dirty sequences) |
| Audio rendering | dominant cost, independent of IR structure |

Bottleneck is always audio rendering, never IR manipulation.

---

## What This Solves

- **Out-of-order instructions** — buffer-drain produces sorted flat IR per phrase
- **Hot-swap complexity** — Phrase replacement, O(1), no Time recalculation
- **Hot-swap duration change** — relative Time means longer/shorter phrase just updates Duration, nothing else shifts
- **Tree traversal at playback** — eliminated, play phase is always flat phrases
- **Memory** — bounded to two FlatIRs (current + next)
- **Latency** — at most one top-level phrase length
