# Bleeder Syntax Reference — Final

Quick reference for implementation. Updated after final design session.

---

## Operators

```go
const (
    opPlay = ">"   // play note/midi
    opLast = "<"   // reference last (with optional modifier)
    opLink = "@"   // link to sequence
    opVibe = "$"   // switch vibe/patch
    opRest = "_"   // rest / advance time
    opWith = "|"   // with (parallel)
    sepArgs = ":"  // argument separator
)
```

| Op | Lane | Riff | Description |
|----|------|------|-------------|
| `>` | play note | hold/extend | Lane: play note/midi. Riff: extend previous note duration |
| `<` | ✓ | ✓ | Reference last operation, optionally with modifier |
| `@` | ✓ | ✓ | Link to nested sequence |
| `$` | ✓ | ✓ | Switch vibe (sound patch) |
| `_` | ✓ | ✓ | Rest / advance time (silence) |
| `|` | ✓ | — | With (parallel) — Riff doesn't need it, rows are parallel |
| `:` | ✓ | ✓ | Argument separator |

---

## Lane vs Riff

| | Lane | Riff |
|---|---|---|
| **Newlines** | whitespace (collapsed) | row separator (parallel voice) |
| **Timeline** | single (sequential by default) | multi-track (one per row) |
| **Parallelism** | explicit `|` | implicit (rows) |
| **Use case** | composition, melodies, sequencing | drums, chords, multi-voice grids |
| **`>` operator** | play note | hold/extend previous |

---

## Timing Model (T/Ta State Machine)

**Variables:**
- `T` — current time cursor (ticks)
- `Ta` — time to advance before next operation

**Algorithm:**
```go
T := 0
Ta := 0
irp := ir.NewProgram()

for each token:
    if token == "|":
        Ta = 0  // cancel pending advance (parallel)
        continue
    
    T += Ta  // advance cursor
    
    switch token:
    case ">":
        ins := &ir.Instruction{
            Midi: note,
            Time: T,
            Dur: duration,
            Vol: volume,
        }
        irp.Add(ins)
        Ta = duration  // or 1 default
        
    case "@":
        nestedIR := GenSequence(seqName, args)
        shiftedIR := nestedIR.Copy()
        shiftedIR.Shift(T)
        irp.Merge(shiftedIR)
        Ta = nestedIR.Duration()
        
    case "_":
        Ta = restDuration  // don't emit instruction
        
    case "<":
        // repeat last instruction with optional modifier
        newIns := lastIns.Copy()
        newIns.Time = T
        if hasModifier:
            newIns.Midi += modifier
        irp.Add(newIns)
        Ta = newIns.Dur
        
    case "$":
        currentVibe = vibeName  // switch context
        Ta = 0  // no time advance
}

irp.Sort()  // sort by Time once at end
return irp
```

**Key rule:** advance time unless next token is `|`

---

## Arguments

**Format:** `operator:arg1:arg2:arg3`

Examples:
```
>c4:2:.8      // play c4, duration 2, volume 0.8
>e2:4         // play e2, duration 4 (volume defaults)
_4            // rest for 4 ticks
@bass:a3:2    // play sequence "bass" with args (a3, 2)
<+7           // repeat last, add 7 semitones
```

**Arithmetic in args:**
```
>c4+7         // c4 plus 7 semitones = g4
<-2           // last note minus 2 semitones
@seq:a3*2     // pass "a3*2" as arg (resolved in sequence context)
```

---

## Lane Examples

### Simple melody
```toml
[lane.melody]
content = ">c4 >e4 >g4 >c5"
```
Sequential: c4 at T=0, e4 at T=1, g4 at T=2, c5 at T=3

### Chord (parallel notes)
```toml
[lane.chord]
content = ">c4 |<+4 |<+7"
```
c4, e4, g4 all at T=0 (major chord)

### Melody with rests
```toml
[lane.groove]
content = ">c4 _2 >e4 _ >g4"
```
c4 at T=0, rest 2 ticks, e4 at T=3, rest 1 tick, g4 at T=5

### Using variables
```toml
[lane.arp]
vars = "root:c4"
content = ">root >root+4 >root+7 >root+12"
```
Arpeggio: c4, e4, g4, c5

### Referencing sequences
```toml
[lane.bass]
content = ">a2:4"

[lane.lead]  
content = ">c4 >e4 >g4"

[lane.verse]
content = "@bass | @lead"
```
Bass and lead play in parallel from T=0

### Parallel with offset
```toml
[lane.pattern]
content = "@long | _2 @short <<<
```
- long at T=0 (duration = 16)
- short at T=2, T=6, T=10, T=14 (in parallel with long)

### Vibe switching
```toml
[lane.section]
content = "$bass >a2:4 $lead >c4 >e4 >g4"
```
Plays a2 with bass vibe, then c4/e4/g4 with lead vibe

### Repeat patterns
```toml
[lane.riff]
content = ">e2 >g2 <+2 <+2"
```
e2, g2, a2, b2 (each builds on previous)

---

## Riff Examples

### Drum pattern
```toml
[riff.beat]
vars = "k:c1 s:d2 h:f#3"
content = """
k _ k _ k _ _ k
_ _ s _ _ _ s _
h h h h h h h h
"""
```
Three parallel voices (kick, snare, hihat), 8 ticks total

### Hold/extend notes
```toml
[riff.pad]
vars = "a:a2 c:c3"
content = """
a > > > > > > >
c > > c > > > >
"""
```
- Row 1: a2 held for 8 ticks
- Row 2: c3 held for 3, c3 again at T=3, held for 5

### With sequence references
```toml
[riff.groove]
vars = "k:c1 s:d2"
content = """
k _ k _
@snare_fill
"""
```
Kick pattern on row 1, sequence @snare_fill on row 2 (parallel)

### Repeat in grid
```toml
[riff.pattern]
vars = "a:e2 b:c2"
content = """
a _ < _
b b _ <
"""
```
- Row 1: a at T=0, rest, repeat a at T=2, rest
- Row 2: b at T=0, b at T=1, rest, repeat b at T=3

### Vibe switching in Riff
```toml
[riff.multi]
vars = "a:a2 b:c2 k:c1"
content = """
$bass
a > > >
b _ b _
$drums
k _ k _
"""
```
First two rows use bass vibe, last row uses drums vibe

---

## Variables

**Definition:** `vars = "name:value name:value"`

**Types:** numbers only (notes resolve to midi, YAGNIY on sequence refs)

**Resolution:**
```go
parseVars(vars, values) → map[string]string
```

1. Split vars by space
2. For each `name:defaultValue`:
   - If override provided in values, use it
   - Else parse defaultValue via `evalArg`:
     - Try `NoteToMidi(value)` (e.g., "e2" → 40)
     - Try `ParseFloat(value)` (e.g., "60" → 60.0)
     - Try arithmetic (e.g., "60+7" → 67.0)
     - If all fail, store as-is (future: string support)
3. Return map for string replacement in content

**Example:**
```toml
[lane.melody]
vars = "root:c4 dur:2"
content = ">root:dur >root+4:dur >root+7:dur"
```

Call: `GenSequence("melody", ["e4", "1"])`
- root = e4 (override)
- dur = 1 (override)
- Result: `>e4:1 >e4+4:1 >e4+7:1`

---

## Composition Patterns

### Layering sequences (parallel)
```toml
[lane.full]
content = "@bass | @drums | @lead"
```
All three sequences start at T=0, play in parallel

### Sequential arrangement
```toml
[lane.song]
content = "@intro @verse @chorus @verse @outro"
```
Each section plays after the previous ends

### Nested with args
```toml
[lane.chord]
vars = "root:60"
content = ">root |<+4 |<+7"

[lane.progression]
content = "@chord:c4 @chord:f4 @chord:g4 @chord:c4"
```
Plays chord progression: C-F-G-C (each is a major chord)

### Multi-track composition
```toml
[lane.bass]
content = ">a2:4 >f2:4"

[lane.lead]
content = "_2 >c4 >e4 >g4"

[lane.arrangement]
content = "@bass | @lead"
```
Bass and lead in parallel, lead starts 2 ticks in

---

## Common Pitfalls

### ❌ Inline multi-track doesn't work
```toml
# WRONG: lead starts at T=20, not T=0
content = ">a2:4 _4 >a2:4 _4 | >c4 >e4 >g4"
```

### ✓ Use composition for parallel tracks
```toml
[lane.bass]
content = ">a2:4 _4 >a2:4 _4"

[lane.lead]
content = ">c4 >e4 >g4"

[lane.together]
content = "@bass | @lead"
```

### ❌ Can't nest `|` without accumulating time
```toml
# WRONG: second | doesn't reset to T=0
content = ">c4 _4 | >e4 _4 | >g4"
```

### ✓ Use Riff for multi-voice grids
```toml
[riff.chords]
content = """
c4 > > >
e4 > > >
g4 > > >
"""
```

---

## Implementation Checklist

- [ ] T/Ta state machine parser
- [ ] `evalArg` helper (note + arithmetic)
- [ ] `parseVars` with override support
- [ ] `GenSequence` with recursion for `@`
- [ ] IR sort after collection
- [ ] Cache (optional, YAGNIY)
- [ ] Lane mode (newlines = whitespace)
- [ ] Riff mode (newlines = rows)
- [ ] Vibe/patch tracking during parse
- [ ] Operator constants
- [ ] Error handling for invalid syntax

---

## Parse Flow Summary

```
Input (Lane or Riff) → Normalize (handle newlines) →
Parse with T/Ta (collect instructions) →
Sort by Time →
Return *ir.Program
```

**No streaming during parse** — nested `@` can contribute at any time, must parse full tree first.

**Phrase-level streaming** — `GenMainIR` loops, calling `GenSequence` per phrase, streams result to renderer.

---

**End of reference. Good luck implementing! ☕**
