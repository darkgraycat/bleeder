# Streaming Architecture — Design Notes

## What problems we tried to solve

Current WAVPlayer works like this:
1. Generate full IR (all instructions, all time)
2. Allocate full sample buffer upfront
3. Write every instruction into buffer at `ins.Time * sampleRate`
4. Write WAV file, play via afplay

This works for "render once, play once". But for live-coding you need streaming —
emit audio continuously, swap content without stopping.

The root problems:
- **Full buffer requires knowing total duration upfront** — can't stream
- **Instructions are unordered in IR** — `@first` and `@second` produce instructions
  interleaved by resolution order, not by time. You can't stream unordered events.
- **`afplay` is fire-and-forget** — no control, no swap, no continuity

---

## Solutions we discussed and why they sucked

### Sort IR before rendering
Sort instructions by `Time` before streaming. Simple, works.
**Problem:** 100K instructions, sorting on every eval, every loop iteration.
Not premature — this bites you hard when you hit scale.

### Channel of Instructions (lazy pipeline)
Replace IR storage with `chan *Instruction`. Bleeder pushes, Renderer pulls.
Zero storage, pure flow.
**Problem:** Bleeder generates instructions out of time order due to nested `@` sequences.
Channel doesn't fix ordering — just moves the problem to Renderer.

### Barriers / time window signals
Bleeder pushes special "barrier" messages: "I'm done with everything up to t=3".
Renderer holds until barrier, then emits that window safely.
**Problem:** Bleeder can't know safe flush points. Nested sequences produce
instructions far into the future before Bleeder knows current offset is safe.

### Clock-based IR scheduler
IR has its own tick clock. Bleeder pushes instructions, clock emits them when
their time arrives.
**Problem:** Same race condition. If Bleeder is slow and clock reaches t=7
before Bleeder pushed t=4 — you lose instructions.

### SequenceRef — fire and forget
Instead of resolving `@first` into instructions, Bleeder pushes
`SequenceRef{name, startTime}`. Renderer resolves sequences just-in-time.
**Problem:** Renderer becomes aware of sequences. Renderer should be dumb —
it renders audio, not DSL.

### Goroutines per sequence
Each `@` spawns a goroutine that resolves the sequence concurrently,
pushes instructions into shared channel.
**Problem:** Multiple goroutines pushing into same channel = unordered again.
Renderer still needs a mixing window. Same problem, more complexity.

---

## What live-coding actually is

**The Eval model** (SonicPi): play on demand, stop when done. You change
something, hit play, new version plays from start.

**The Loop model** (Tidal/Strudel): continuous loop, always playing.
You edit inside the loop. Changes take effect at the next cycle boundary.
Seamless, no restart. You can perform live concerts. This is what bleeder wants.

---

## The architecture that emerged

### Key insight: the phrase is the atomic unit

A long loop (1 minute) means waiting up to 1 minute for changes to land.
That kills the live feel.

The solution: **sub-cycle hot-swap at phrase boundaries**.

Each top-level `@sequence` call in main IS a phrase. When you save the file,
the change takes effect when the currently-playing sequence ends and the
next one begins. Latency = at most one sequence length (seconds, not minutes).
Changes always land on a musical boundary. Feels responsive. Feels right.

> Note from Cat: nope, each @sequence is "phrase" and going to be re-played once changed

### The model

```
[meta]
main = "main"

[lane.main]
content = '''
@intro
@verse
@chorus
@outro
'''
```

Main loop runs forever. Each sequence = one hot-swappable phrase.

```
Renderer plays: @intro → @verse → @chorus → @outro → @intro → ...
                                   ↑
                        you edit @chorus, save file
                        Bleeder re-parses @chorus
                        slots updated version in at next @chorus boundary
```

> Note from Cat: nope, if @chorus is currently played and I updated it (and saved file) - it gets re-played one more time (in case @outro isnt started yet).

### Why this solves everything

**Out-of-order instructions?** — Bleeder resolves one sequence at a time.
Each sequence is small and finite. Sort microseconds, not millions of instructions.

**Memory?** — Bounded. One sequence in Renderer, one being prepared in Bleeder.
Not the whole loop.

**Race conditions?** — Gone. You have the entire current sequence's duration
to resolve the next one. Always enough time.

**Latency?** — At most one sequence length. Keep sequences short = responsive feel.

**Live-coding?** — On file save, re-parse only changed sequences.
Slot them in at next boundary. No restart, no gap.

### The double-buffer

```
Renderer:  playing @verse (current buffer)
Bleeder:   resolving @chorus (next buffer) ← re-parsed from saved file

@verse ends → swap buffers → Renderer plays @chorus
              Bleeder starts resolving @outro (or re-parsed @verse if changed again)
```

### IR stays simple

IR doesn't need to be a stream, a channel, or a scheduler.
It's just one sequence worth of instructions — small, finite, sortable trivially.
Double-buffered between Bleeder and Renderer.

`indexesCache` on Program — drop it. Unused overhead.
`timeScale` on Program — drop it. Unused.

Instruction gets `Vibe *Vibe` field for instrument/sound shape.
Everything else stays.

---

## What's next (when you wake up)

1. Update architecture doc with phrase/sequence hot-swap model
2. Decide: does `ir.Program` stay as-is (sorted slice) or become something leaner?
3. Design the double-buffer swap mechanism between Bleeder and Renderer
4. Think about file-watch hook for triggering re-parse on save

# More notes from Cat
About 4. :
We going to have a daemon running, and daemon starts with original file to perform a preparsing.
Then when Vim saves the file it can send a signal, I guess.
The main problem - I need to know locations of all edits.

But what if we could do full stdin/stdout interface?
```sh
# play bleed
cat bleed.toml | bleeder wav | ffplay

# save using wav renderer
cat bleed.toml | bleeder wav > song.wav

# save using txt renderer
cat bleed.toml | bleeder txt > tabs.txt

# play raw
echo ">e3_ >a3" | bleeder wav | ffplay

# start in daemon mode (is it possible?)
# bleeder preparses all included bleeds (they are readonly for now)
cat bleed.toml | bleeder wav --listen > ffplay

# send to daemon on file-save
cat bleed.toml | bleeder --send

# send raw to daemon
# @some gets resolved from bleed.toml parsed on startup
echo ">e3_ >a3_ @some" | bleeder --send

### Or what the point of having stdin? So what if:
# play
bleeder bleed.toml --wav | ffplay
# daemon (outputs as TUI renderer of playback)
bleeder bleed.toml --wav

```
> Btw, is it possible to have readable and workable DSL using | and > as in shell? Dunno. Like piping notes into synth? `e2 a2 | ??? > bass`? What can be "???" then?

As for 1. 2. and 3. :
I think we need completely new model, which includes:
file-diff - to know which sequence needs to be replayed, in case of multiple sequences is updated we need to jump into first one that appears on playback, not just position in file
...and much more



