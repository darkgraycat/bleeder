# Bleeder Syntax & Timing Model v2

## Core Philosophy

**`@` is the superpower** — composition through named abstraction. Everything builds from sequences calling sequences.

**One language, two modes** — Lane and Riff are the same syntax with different newline handling.

**Auto-advance + explicit parallelism** — time flows naturally, `&` opts into parallel execution.

---

## Timing Model

### Auto-Advance by Default

Each operator advances time automatically (by duration or 1 tick):

```
>c4 >e4 >g4    // plays at T=0, T=1, T=2 (melody)
```

### `&` for Parallelism

`&` prevents time advance, making the next operation parallel with the previous:

```
>c4 &|+4 &|+7    // chord: c4, e4, g4 all at T=0
```

**Rule:** advance time unless next token is `&`

### Implementation (T/Ta state)

```go
T := 0   // current time cursor
Ta := 0  // time to advance before next op

for each token:
    if token == "&":
        Ta = 0  // cancel pending advance
        continue
    
    T += Ta  // advance cursor
    
    switch token:
    case "@sequence":
        add instruction at T
        Ta = sequence.Duration()
    case ">note":
        add instruction at T
        Ta = note.Dur  // or 1 default
    case "_":
        Ta = wait.Dur  // rest, don't emit instruction
    case "|":
        repeat last instruction at T
        Ta = last.Dur
    }
```

No fork, no lookahead — just two variables.

---

## Operators

| Op | Lane | Riff | Description |
|----|------|------|-------------|
| `>` | ✓ | ✓ | Play note/midi (float64) |
| `@` | ✓ | ✓ | Play nested sequence |
| `_` | ✓ | ✓ | Rest / advance time (silence) |
| `|` | ✓ | ✓ | Repeat last operation |
| `&` | ✓ | — | Play in parallel (Riff rows already parallel) |
| `$` | ✓ | ✓ | Switch vibe (sound patch) |

**Whitespace:** always meaningful separator (space, tab, newline in Lane)

**Arguments:** space-separated after operator: `>c4 2 .8` (note, duration, volume)

**Arithmetic:** inline expressions: `>c4+7`, `|+4`, `e2*2` (parsed via `evalArg`)

---

## Lane Mode

**Newlines = whitespace** (collapsed, no semantic meaning)

**Use `&` for explicit parallelism**

### Examples

**Melody:**
```
>c4 >e4 >g4 >c5
```

**Chord:**
```
>c4 &|+4 &|+7
```

**Parallel sequences:**
```
@bass &@drums
```

**Parallel with offset:**
```
@long &_2 @short |||
```
= long at T=0, shorts at T=2, 6, 10, 14 (in parallel)

**Multi-line (same as inline):**
```
>c4
>e4
>g4
```
= T=0, T=1, T=2 (sequential)

---

## Riff Mode

**Newlines = row separator** (each row is a parallel voice)

**Rows are parallel by default** (no `&` needed)

**Columns = ticks** (visual alignment, space-separated tokens)

### Examples

**Drum pattern:**
```
[riff.drums]
vars = "k=c2 s=d2 h=f#2"
content = """
k _ k _
_ s _ s
h h h h
"""
```

Three voices in parallel:
- kick: T=0, T=2
- snare: T=1, T=3
- hihat: T=0, T=1, T=2, T=3

**Sequence in grid:**
```
k _ k _
@bass _ _ _
```

**Vibe switching:**
```
$bass
a _ a a
$drum
k k _ k
```

---

## Lane vs Riff

| | Lane | Riff |
|---|---|---|
| **Newlines** | whitespace (ignored) | row separator (parallel voice) |
| **Parallelism** | explicit `&` | implicit (rows) |
| **Use case** | melodies, sequencing, composition | drums, chords, multi-voice grids |
| **Mental model** | linear flow, time advances | grid, columns = ticks |

**Implementation:** same parser, mode flag controls newline handling

---

## Variables

**Vars are numbers only** (no sequence refs in vars, YAGNIY)

**Defined in TOML:**
```toml
[lane.melody]
vars = "e=e2 dur=2"
content = ">e dur >e+7 dur"
```

**Resolution:** `parseVars(vars, values) → map[string]string`
- Try parse as number (note or float) via `evalArg`
- If NaN, keep as string (for future features)
- Apply to content via string replacement

---

## Composition Examples

**Bass + drums in parallel:**
```toml
[lane.verse]
content = "@bass &@drums"
```

**Arpeggio pattern:**
```toml
[lane.arp]
vars = "root=c4"
content = ">root >root+4 >root+7 >root+12"
```

**Long sequence with short repeating:**
```toml
[lane.texture]
content = "@pad &@click |||"
```
= pad plays once, click repeats 4 times in parallel

**Riff as multi-voice:**
```toml
[riff.chord_prog]
vars = "a=a2 c=c3 e=e3"
content = """
a a a a
c c e e
e e g g
"""
```

---

## Parse → Sort → Stream

**Flow:**
1. Parse sequence content (recursively resolve `@`)
2. Collect all instructions with **absolute time** T
3. **Sort once** by T
4. Cache result (optional, YAGNIY)
5. Renderer streams or consumes as needed

**Phrase-level streaming:**
```go
GenMainIR loop:
  invalidate dirty cache
  program := GenSequence(main, [])
  for _, ins := range program.Instructions():
    send to renderer → ffplay
  (loop repeats)
```

**Why not stream during parse?** Nested `@sequences` can contribute instructions at any absolute time (including the past). Must parse full tree before emitting in order.

---

## Design Decisions Summary

✓ **Auto-advance by default** — optimizes for common case (melodies), less `_` noise  
✓ **`&` for explicit parallelism** — simple lookahead rule, clean mental model  
✓ **`@` as composition primitive** — named abstraction > inline nesting  
✓ **Numbers-only vars** — defer sequence-ref vars (YAGNIY)  
✓ **Token-based (space-separated)** — loses char-grid compactness, gains expressiveness  
✓ **One language, two modes** — Lane/Riff differ only in newline semantics  
✓ **T/Ta state machine** — no fork/stack, simple and fast  
✓ **Parse whole phrase first** — sort once, stream at phrase level  

---

## Next Steps

- Implement T/Ta parser for Lane mode
- Add Riff mode (newline as row separator)
- Test with real dark-wave patterns
- Measure parse speed (decide if cache is needed)
- Build text renderer for IR visualization
