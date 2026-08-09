# Bleeder Live-Coding Architecture Plan

## Overview

Build MVP live-coding functionality as a monolith. Bleeder will handle parsing, IR generation, streaming, and WAV rendering in a single binary. Future optimization may split into separate tools (bleeder → bleeder-wav pipeline).

## Core Concept

**Chunk = instructions with same timestamp**

Instructions that start at the same time are grouped and sent together. This naturally handles parallel operations (`|` in DSL) and provides clean batching boundaries.

---

## Architecture

```
┌─────────────────────────────┐
│  bleeder live song.bleed    │
│  ┌───────────────────────┐  │
│  │   TCP Server          │  │
│  │   - play <seq>        │  │
│  │   - stop              │  │
│  │   - sync              │  │
│  │   - info              │  │
│  └──────────┬────────────┘  │
│             ↓               │
│  ┌───────────────────────┐  │
│  │   BleedCtx            │  │
│  │   - Manages state     │  │
│  │   - Generates IR      │  │
│  │   - Streams chunks    │  │
│  │   - Loops infinitely  │  │
│  └──────────┬────────────┘  │
│             ↓               │
│  Serialized IR (stdout)     │
└─────────────┬───────────────┘
              ↓
    Timestamped instructions
              ↓
         ffplay -
```

---

## Components

### 1. BleedCtx (core/bleed_ctx.go)

**Responsibility:** High-level API for Bleeder. Business layer between Bleeder/Bleed and Cmd.

**State:**
- `bleed *Bleed` - current bleed file data
- `bleeder *Bleeder` - DSL processor
- `irp *ir.Program` - current IR
- `position float64` - current playback position (time)
- `isPlaying bool` - playback state
- `currentSeq string` - which sequence is playing
- `mu sync.Mutex` - thread safety

**API Methods:**

```go
// Create new context
func NewBleedCtx(bleed *Bleed) *BleedCtx

// Start playing/looping a sequence
func (ctx *BleedCtx) Play(seqName string) error

// Stop playback
func (ctx *BleedCtx) Stop() error

// Reload bleed file, regenerate IR, swap from current position
func (ctx *BleedCtx) Sync() error

// Get current state (playing, position, sequence)
func (ctx *BleedCtx) Info() string

// Stream IR chunks to writer (blocking, runs loop)
func (ctx *BleedCtx) Stream(w io.Writer) error
```

**Notes:**
- BleedCtx does NOT watch files (external tools send sync command)
- Handles infinite looping internally
- Thread-safe (TCP commands + streaming run concurrently)

---

### 2. Streaming Model

**Chunk Definition:**
All instructions with the same timestamp (`Time` field).

**Example IR:**
```
Instruction{Time: 0.0, Midi: 60, Dur: 8.0}  ┐
Instruction{Time: 0.0, Midi: 64, Dur: 1.0}  ├─ Chunk 1 (t=0.0)
                                            ┘
Instruction{Time: 1.0, Midi: 67, Dur: 1.0}  ← Chunk 2 (t=1.0)
Instruction{Time: 2.0, Midi: 62, Dur: 1.0}  ← Chunk 3 (t=2.0)
```

**Serialized Output Format:**
```
play t=0.0 m=60 d=8.0 v=1.0
play t=0.0 m=64 d=1.0 v=1.0
play t=1.0 m=67 d=1.0 v=1.0
play t=2.0 m=62 d=1.0 v=1.0
```

**Streaming Pseudocode:**
```go
func (ctx *BleedCtx) Stream(w io.Writer) error {
    for ctx.isPlaying {
        // Get unique timestamps from IR
        times := ctx.irp.UniqueTimes()
        
        for _, t := range times {
            // Get all instructions at this time (chunk)
            chunk := ctx.irp.InstructionsAtTime(t)
            
            // Serialize and write chunk
            ctx.writeChunk(chunk, w)
            
            // Update position
            ctx.position = t
            
            // Check for sync/stop commands
            if ctx.checkSwap() {
                break // Restart loop with new IR
            }
        }
        
        // Loop: restart from beginning
        ctx.position = 0.0
    }
}
```

---

### 3. Swap/Sync Behavior

**On `sync` command:**

1. Generate new IR (in background, while old IR still playing)
2. When new IR ready, swap at next chunk boundary
3. **Let old notes finish** - don't stop currently playing notes
4. Start streaming new IR from current position

**Example:**

```
Old IR:
  t=0.0 m=60 d=8.0  (long note)
  t=1.0 m=67 d=1.0

New IR (after sync):
  t=0.0 m=64 d=8.0  (different note)
  t=1.0 m=69 d=1.0  (different note)

Timeline:
  t=0.0  Old m=60 starts (8 beat duration)
  t=1.0  SYNC happens
         - Old m=60 keeps playing until t=8.0
         - New IR: skip t=0.0 (in past), send t=1.0 m=69
  t=1.0+ New m=69 plays, old m=60 finishes naturally
```

**Why this approach:**
- Simple (no state tracking of active notes)
- No abrupt stops (old notes finish gracefully)
- Acceptable overlap during transition
- Good enough for live-coding MVP

---

### 4. TCP Server (cmd/server.go)

**Part of `bleeder live` command.**

**Commands:**

#### `play <seqName>`
Start playing/looping the named sequence infinitely.

Example:
```
play main
play myRiff
```

Response:
```
OK playing <seqName>
```

#### `stop`
Stop playback.

Response:
```
OK stopped
```

#### `sync`
Reload bleed file, regenerate IR, swap from current position.

Response:
```
OK synced
```

#### `info`
Get current state.

Response:
```
playing: main
position: 4.5
loop: 3
```

**Protocol:**
- Text-based commands (one per line)
- Responses prefixed with `OK` or `ERR`
- Multiple clients can connect (commands queued)

---

### 5. Looping

**Infinite Loop:**
When `play <seq>` is called, BleedCtx loops the sequence infinitely.

```go
for ctx.isPlaying {
    // Stream full IR
    for _, time := range times {
        chunk := getChunk(time)
        write(chunk)
    }
    // Restart from beginning
    position = 0.0
}
```

**No built-in loops in sequences** - looping is external (controlled by `play` command).

---

## Implementation Steps

### Phase 1: BleedCtx Foundation
- [ ] Create `core/bleed_ctx.go`
- [ ] Implement `NewBleedCtx(bleed)`
- [ ] Implement basic state management (Play, Stop, Info)
- [ ] Add mutex for thread safety

### Phase 2: IR Streaming
- [ ] Implement `Stream(w io.Writer)` method
- [ ] Add `UniqueTimes()` to `ir.Program`
- [ ] Add `InstructionsAtTime(t)` to `ir.Program`
- [ ] Implement chunk serialization (text format)
- [ ] Implement infinite looping

### Phase 3: Sync/Swap
- [ ] Implement `Sync()` method
- [ ] Reload bleed file
- [ ] Regenerate IR
- [ ] Swap from current position
- [ ] Test overlap behavior

### Phase 4: TCP Server
- [ ] Create `cmd/server.go`
- [ ] Implement TCP listener
- [ ] Parse commands (play, stop, sync, info)
- [ ] Wire to BleedCtx methods
- [ ] Handle concurrent connections

### Phase 5: CLI Integration
- [ ] Update `CmdLive` to use BleedCtx
- [ ] Start TCP server
- [ ] Start streaming to stdout
- [ ] Handle graceful shutdown

### Phase 6: Testing
- [ ] Test single sequence playback
- [ ] Test looping
- [ ] Test sync/swap during playback
- [ ] Test TCP commands
- [ ] Test with `ffplay -`

---

## Future Enhancements (Post-MVP)

### Renderer Split
Extract rendering into separate binaries:
```bash
bleeder live song.bleed | bleeder-wav | ffplay -
bleeder live song.bleed | bleeder-irp | tui-viz
```

### Multiple Renderers
Run multiple renderers simultaneously:
```bash
bleeder live song.bleed | tee >(bleeder-wav | ffplay -) >(bleeder-irp | tui-viz)
```

### Advanced Swap Strategies
- Quantized swap (wait for beat/measure boundary)
- Crossfade between old/new IR
- Smart note continuation

### T Constant
Add `T` (absolute time) variable for pattern-based content:
```toml
content = ">T*4 _1 >T*4+2 _1"
```

### Vibe Streaming
Include vibe/patch changes in serialized IR.

---

## Notes

- **Monolith first**: Keep everything in one binary for MVP
- **Simple swap**: Let old notes finish, no state tracking
- **Text format**: Human-readable, debuggable
- **External watching**: No file watching in bleeder, external tools send `sync`
- **Chunk = timestamp**: Natural batching boundary

---

## Questions / Decisions Pending

- [ ] Exact text format for serialized instructions
- [ ] Error handling for malformed bleed files during sync
- [ ] Max buffer size for TCP commands
- [ ] Graceful shutdown behavior (finish current chunk? stop immediately?)
