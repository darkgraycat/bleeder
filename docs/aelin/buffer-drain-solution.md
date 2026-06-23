# Buffer-Drain Solution — Ordered IR Without Global Sort

## The Problem

Bleeder generates instructions out of time order when `@` sequences are involved.
`@first` and `@second` produce instructions interleaved by resolution order, not by time.
Global sort is expensive and gets worse at scale.

---

## Key Semantics

### `@sequence` is fire-and-forget
When the parser encounters `@first`, it fires @first at the current T and immediately continues parsing the main sequence.
`_N` after `@first` in the parent advances the **parent's** T — it does NOT wait for @first to finish.

This means @first's internal instructions can land at times BEYOND the next parent instruction — they interleave.
That's why ordering is needed.

### Time in IR: delta encoding (MIDI-style)
Each instruction's `Time` field = **delta from the previous instruction** in the final sorted output.
Not absolute. Not relative-to-phrase-start. Delta.

Buffer-drain produces sorted absolute-time instructions during parse.
Delta encoding is applied **after** sorting, as a final conversion step.

These are two separate concerns.

## The Insight

Two observations:
1. `_` is the **only thing that advances T**. T is monotonically increasing.
2. When `_` fires, you know nothing before currentT is coming from the main sequence anymore.

So `_` is both a **time advance** AND a **drain signal**.

---

## The Solution

Bleeder maintains a **min-heap buffer** (ordered by absolute T) alongside its main parse loop.

When `@sequence` is encountered:
- Pre-resolve the sequence into (absoluteT, Instruction) pairs
- absoluteT = currentT + instruction.relativeT
- Insert all pairs into the buffer

When `_` is encountered:
- Advance T
- Drain all buffer entries where T ≤ currentT → emit them in order
- Continue parsing

When main sequence ends:
- Flush all remaining buffer entries → emit them in order

---

## Trace Example

```
// main
~100 _1
@first _1
~200

// first
~20_2 ~40 ~60
```

`@first` fires at T=1 (fire-and-forget). Main's `_1` after `@first` advances main to T=2.
`_2` inside `@first` advances @first's internal time from T=1 to T=3.

| Step | Action | T | Buffer | Emitted (abs T) |
|------|--------|---|--------|-----------------|
| 1 | `~100` | 0 | [] | {100, T=0} |
| 2 | `_1` | 1 | [] | — |
| 3 | `@first` → insert ~20,~40,~60 | 1 | [(1,~20), (3,~40), (3,~60)] | — |
| 4 | drain T≤1 | 1 | [(3,~40), (3,~60)] | {20, T=1} |
| 5 | `_1` | 2 | [(3,~40), (3,~60)] | — |
| 6 | `~200` | 2 | [(3,~40), (3,~60)] | {200, T=2} |
| 7 | end — flush | 2 | [] | {40, T=3}, {60, T=3} |

**Final stream (absolute T, sorted):**
```
T=0: ~100
T=1: ~20
T=2: ~200
T=3: ~40
T=3: ~60
```

**After delta encoding:**
```
100 - dt=0
20  - dt=1
200 - dt=1
40  - dt=1
60  - dt=0  (simultaneous)
```

No global sort. Buffer held at most 2 items. ✓

---

## Complexity

- Each sequence pre-sorted once at parse time — small, trivial
- Buffer insert: O(log n) per instruction
- Buffer drain: O(k log n) where k = drained count
- Buffer size: bounded by sum of lookahead instructions across active sequences
- No global sort ever needed

---

## Properties

- **Time ordered output** — guaranteed by drain-on-`_` invariant
- **Memory bounded** — buffer holds only "future" instructions, drains as T advances
- **Streaming friendly** — emitted instructions go straight to Renderer, no accumulation
- **Naturally parallel** — multiple instructions at same T emitted together (chords)
- **Works with nested `@`** — recursive pre-resolve, same mechanism at every depth

## Delta encoding trade-off

The parse-time buffer uses **absolute T** internally — it's transient, built and discarded per-parse. No issue.

The final IR stores **delta T**. This means any `map[absoluteT][]Instruction` optimization is impossible — a delta value is meaningless without walking from the start of the phrase.

This is intentional. The renderer always walks instructions sequentially (front-to-back), and hot-swap replaces whole Phrases. There is no use case that needs random time access into the IR. Delta encoding wins: compact, MIDI-compatible, no hidden cost.

---

## Implementation sketch

```go
type bufferedEntry struct {
    t   int
    ins *ir.Instruction
}

// min-heap by t
type pendingBuffer []bufferedEntry

func (b *Bleeder) GenSeqIR(name string, args []string, parentT int) {
    seq := b.seqs[name]
    // pre-resolve sequence, offset all Times by parentT
    // insert into b.buffer
}

func (b *Bleeder) drainBuffer(upToT int, out chan *ir.Instruction) {
    for len(b.buffer) > 0 && b.buffer[0].t <= upToT {
        out <- heap.Pop(&b.buffer).(*ir.Instruction)
    }
}

// In main parse loop:
// case lcWait:
//     T += delay
//     b.drainBuffer(T, out)
//
// case lcLink:
//     b.GenSeqIR(name, args, T)  // pre-resolve into buffer
//
// end of sequence:
//     b.drainBuffer(MaxInt, out) // flush remaining
```
