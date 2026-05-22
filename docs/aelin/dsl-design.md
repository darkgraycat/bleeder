# DSL Design: Lane, Riff, Vibe

## Lane-DSL (current, working)

Linear time-based. Each operator advances or places in time.

### Operators

| Char | Name | Args              | Example          |
|------|------|-------------------|------------------|
| `>`  | midi | midi, dur, vol    | `>60 2 0.8`      |
| `:`  | note | note, dur, vol    | `:e3 2 0.8`      |
| `~`  | freq | hz, dur, vol      | `~440 2 0.8`     |
| `_`  | wait | dur (opt)         | `_2` or `_`      |
| `\|` | last | mods              | `\|+7 \|*2`      |
| `@`  | link | name, vars...     | `@chord5 e3 2`   |
| `$`  | vibe | name              | `$bass`          |

### Arithmetic on args

All args support `+-*/` modifiers: `:e3+2 2*3 0.5/2`

### Var expansion

Sequence vars defined as `note:e3 d:2`, referenced in content as plain tokens.
Left-to-right resolution: `A:60 B:A*2` — B resolves to 120.
String vars (sequence names) stay as strings: `chord:chord5` → `@chord e3`

### Example

```toml
[lane.main]
content = '''
$bass
@chord5 a3 8 _
@chord5 f3 8 _
$lead
:e4 2 |+7 |+5 _2
'''

[lane.chord5]
vars = "note:e2 d:8 v:1.0"
content = '''
:note d v _ |+7 |+5
'''
```

---

## Riff-DSL (design, not implemented)

Grid-based. Time is implicit — position in the grid determines when.
One character = one time slot. Rows = voices. Spaces = bar separators.

### Characters

| Char | Meaning              |
|------|----------------------|
| `-`  | rest (silence)       |
| `>`  | sustain (hold prev)  |
| any  | var reference        |

### Slot duration

Determined by tempo + bar length. Default: 1 bar = 16 slots.
Override per riff: `beats = 8` in riff definition.

### Vars

Same as Lane vars. Vars map characters to notes, freqs, or lane sequences.

### Example

```toml
[riff.groove]
vars = "k:>36 s:>38 h:~8000 0.3"
content = '''
k-k- k-k-
-s-- -s--
hhhh hhhh
'''
```

Three voices: kick, snare, hihat. Aligned grid, read left to right.

`k` → plays MIDI 36 (kick). `-` → rest. `>` → sustain previous.

### Bar separator

Space in content = visual separator only, no timing effect.
`k-k- k-k-` = 8 slots, same as `k-k-k-k-`.

---

## Vibe-DSL (design, not implemented)

Defines sound shape. Controls synthesis, not timing.

### Structure

```toml
[vibe.bass]
args = "shape:saw attack:0.01 decay:0.1 sustain:0.8 release:0.2"

[vibe.kick]
args = "shape:sine attack:0.001 decay:0.3 sustain:0.0 release:0.05"

[vibe.piano]
args = "shape:sine attack:0.005 decay:0.4 sustain:0.6 release:0.5"

[vibe.pad]
args = "shape:triangle attack:0.3 decay:0.1 sustain:0.9 release:0.8"
```

### Shape values

- `sine` — pure tone
- `saw` — rich harmonics, buzzy
- `square` — hollow, reedy
- `triangle` — softer than saw

### ADSR

- `attack` — seconds from 0 to peak
- `decay` — seconds from peak to sustain level
- `sustain` — amplitude level (0.0–1.0) held during note
- `release` — seconds from sustain to 0 after note ends

### Usage in Lane

`$bass` in lane content switches current vibe. All subsequent instructions inherit it until next `$`.

```
$kick
>36 1 _
$snare
>38 1 _
$kick
>36 1 _
```

### Default vibe

If no `$` used, default vibe is `sine` with flat envelope (no ADSR, just on/off).
