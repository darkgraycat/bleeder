# Bleeder Roadmap

Live-coding music sequencer/synthesizer - implementation plan

---

## Current Status

**DONE:**
- ✅ Lane DSL parser (sequential syntax)
- ✅ Riff DSL parser (grid-based syntax)
- ✅ Variable system with boundary-based replacement
- ✅ IR (Intermediate Representation) generation
- ✅ Error handling with structured messages
- ✅ Performance optimization (sub-500ns per operation)
- ✅ Line continuation with `|` operator
- ✅ TOML file loading with includes
- ✅ Sequence linking with `@` operator
- ✅ Basic WAV renderer (prototype)

**NEXT:** Build the live-coding experience

---

## Phase 1: Get The Loop Working ⚡

**Goal:** write code → hear music → tweak → hear changes

**Time estimate:** ~6 hours  
**Priority:** CRITICAL

### 1. Basic Tempo/BPM System
**Time:** 2-3 hours

Convert IR abstract time units → real seconds based on BPM

**Tasks:**
- [ ] Add `TimeScale()` method to calculate seconds-per-tick from tempo
- [ ] Update WAV player to use tempo-based time conversion
- [ ] Remove `forDebugTimeTempVariableAtAll` hack from WAV player
- [ ] Test with different tempos (60, 120, 140 BPM)

**Formula:**
```
secondsPerTick = 60.0 / tempo / 4.0
realTime = irTime * secondsPerTick
```

---

### 2. Simple File Watcher
**Time:** 2-3 hours

Auto-reload and replay on file save (no daemon yet - just full reload)

**Tasks:**
- [ ] Add file watcher (fsnotify or similar)
- [ ] Implement watch loop: load → parse → render → play → wait for change
- [ ] Handle parse errors without crashing watcher
- [ ] Add debouncing (don't reload on every keystroke)

**Simple approach:** Full reload on every change. Hot-swap can come later if needed.

---

### 3. Basic CLI
**Time:** 1 hour

Make it usable from terminal

**Tasks:**
- [ ] Implement `bleeder play <file>` - one-shot playback
- [ ] Implement `bleeder watch <file>` - live-coding mode
- [ ] Add `-h/--help` flag
- [ ] Add version flag

**Commands:**
```bash
bleeder play song.bleed    # render and play once
bleeder watch song.bleed   # watch file, auto-reload on save
```

---

## Phase 2: Make It Sound Good 🔊

**Goal:** Enjoy composing, not just prototyping

**Time estimate:** 4-6 hours  
**Priority:** IMPORTANT

### 4. Better WAV Renderer

**Current issues:**
- Fixed ADSR envelope (0.03, 0.06 hardcoded)
- Limited waveforms (only soft square + parabola mix)
- Basic soft-clipping normalization

**Tasks:**
- [ ] Configurable ADSR envelopes (attack, decay, sustain, release)
- [ ] More waveforms: sine, saw, triangle, noise
- [ ] Better volume normalization/limiting
- [ ] Maybe: simple low-pass filter for warmth
- [ ] Maybe: cross-platform audio playback (not just `afplay`)

**Keep it simple!** Don't build a full synthesizer yet - just good-enough sound quality.

---

## Phase 3: Make It Robust 🛡️

**Goal:** Live-coding without crashes or frustration

**Time estimate:** 6-8 hours  
**Priority:** IMPORTANT

### 5. Cycle Detection
**Time:** 2 hours

Prevent infinite recursion in sequence links

**Tasks:**
- [ ] Track call stack during `GenSeqIR`
- [ ] Detect cycles (e.g., `main → verse → chorus → verse`)
- [ ] Return clear error: `cycle detected: main → verse → chorus → verse`

**Simple approach:** Pass call stack slice through recursive calls, check membership before recursing.

---

### 6. Daemon + Unix Socket (OPTIONAL)
**Time:** 4-6 hours

Hot-swap IR without stopping playback

**Only do this if file watcher feels too slow!**

**Tasks:**
- [ ] Daemon process that stays running
- [ ] Unix socket for communication
- [ ] Protocol: send new IR, daemon swaps on next cycle
- [ ] Neovim integration (send buffer to daemon)
- [ ] Error recovery (don't crash on parse errors)

**Defer until Phase 1 file watcher proves too slow.**

---

## Phase 4: Expressiveness 🎨

**Goal:** More creative control

**Time estimate:** 10-12 hours  
**Priority:** NICE TO HAVE

### 7. Vibe System
**Time:** 6-8 hours

Audio patches and effects

**Current state:**
- Vibe struct exists in TOML
- `$vibeName` operator parses but returns "not implemented"

**Tasks:**
- [ ] Define vibe effects (filter, reverb, distortion, etc.)
- [ ] Add vibe reference to IR instructions (or global patch?)
- [ ] Implement effect chain in WAV renderer
- [ ] Test with different vibes per sequence

---

### 8. MIDI Export
**Time:** 3-4 hours

Export compositions for DAW integration

**Tasks:**
- [ ] IR → MIDI file converter
- [ ] Tempo mapping (BPM → MIDI tempo)
- [ ] Track/channel assignment
- [ ] CLI: `bleeder export song.bleed -o out.mid`

---

## Polish & Future

**Lower priority - defer until core experience is solid**

### Other Features
- [ ] Config file (`~/.config/bleeder/config.toml`)
- [ ] Better error messages for musicians (not just devs)
- [ ] Tab export (guitar tabs from note sequences)
- [ ] Documentation (DSL reference, examples, architecture)
- [ ] More operators (maybe randomness, conditionals?)

---

## Development Strategy

### Why This Order?

**Vertical Slice First:**  
Get the EXPERIENCE working (live-coding loop) before polishing layers.

**Fast To Usable:**  
Phase 1 = ~6 hours → working live-coding instrument!

**Musical First:**  
Tempo + file watcher = actually usable for composition.

**Defer Complexity:**  
Daemon/socket adds complexity - only needed if file reload feels slow.

### Next Steps

1. **Start with Tempo/BPM** - foundation for everything else
2. **Add file watcher** - unlock live-coding experience  
3. **Build CLI** - make it usable
4. **= ~6 hours to live-coding instrument!**

Then decide based on needs:
- Sounds good enough? → Add vibe system
- Sounds bad? → Improve WAV renderer
- File reload too slow? → Add daemon

---

## Notes

**Architecture:**
```
.bleed file (TOML)
    ↓
Bleed struct (parsed)
    ↓
Bleeder processor
    ↓
IR Program (flat instruction array with time/midi/dur/vol)
    ↓
Player (WAV / MIDI / tabs)
```

**Key Design Principles:**
- Immutable after init
- Pre-cache everything
- Simple flat structures
- No over-engineering
- YAGNI - defer complexity until needed

**Performance Targets:**
- Parser: <500ns per operation ✅
- File reload: <100ms for typical song
- Audio latency: <50ms (if daemon mode)
