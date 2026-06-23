# Parser Design — Final

## Bleeder's Single Responsibility

Given a sequence name and args, return a correctly ordered `*ir.Program`. Fast.
Everything else (streaming, hot-swap, rendering) is the caller's problem.

```go
func (b *Bleeder) GenSequence(name string, args []string) (*ir.Program, error)
```

---

## Parse Algorithm

No buffer. No heap. No incremental drain.

1. Parse sequence content recursively
2. Collect all instructions with **absolute time** into a flat slice
3. Track `currentVibe` — stamp onto each instruction at emit time
4. When `@sequence` encountered — recurse, get back flat slice, shift all times by current offset, append
5. `_` just advances local time offset
6. **Sort once at the end** by absolute time — O(N log N), fast for any realistic sequence size
7. Convert to **delta encoding** (each instruction's time = delta from previous)

```go
// during parse
ins := &ir.Instruction{
    Time: currentOffset,  // absolute, pre-sort
    Vibe: currentVibe,
    ...
}

// post-sort conversion
for i := 1; i < len(instructions); i++ {
    instructions[i].Time -= instructions[i-1].Time
}
```

---

## Cache

Two-level map — name → args → program:

```go
cache map[string]map[string]*ir.Program
```

- Cache key: `(name, args)`
- Cache hit: return cached program directly — delta encoding means no time shifting needed, result is always valid regardless of call site
- Cache miss: parse, store, return

### Invalidation

When sequence `"foo"` changes:

```go
delete(cache, "foo")  // wipes all arg variants in one shot
```

Cascade: also delete any sequence that `@references` foo (transitively). Only dirty sequences re-parse. Clean sequences are instant cache hits.

---

## Hot-Swap

No explicit swap mechanism needed.

1. File watcher / unix socket detects file change
2. Identify changed sequences, delete from cache (+ cascade)
3. `GenMainIR` loops forever — next iteration hits cache miss → re-parses → updated version plays
4. Phrase boundary = hot-swap point, falls out of the loop naturally

---

## Streaming

Streaming happens at the **phrase level**, not instruction level.

Within a single sequence, you cannot stream — nested `@sequences` can contribute instructions at any absolute time, so you must parse the whole tree before emitting anything.

```go
func (b *Bleeder) GenMainIR(out chan<- *ir.Instruction) error {
    for {
        b.invalidateDirty()
        program, err := b.GenSequence(b.main, nil)
        if err != nil {
            return err
        }
        for _, ins := range program.Instructions() {
            out <- ins
        }
    }
}
```

Renderer sees a continuous stream of correctly ordered, delta-encoded instructions.

---

## Vibe

`$vibe` operator switches `currentVibe` during parsing.
Vibe is baked into each instruction at parse time — renderer never resolves vibes, just reads the field.

---

## Summary

| Concern | Solution |
|---------|----------|
| Ordering | Collect with abs time → sort once → delta encode |
| Performance | Two-level cache, copy on hit, no time shifting needed |
| Invalidation | `delete(cache, name)` — wipes all arg variants |
| Hot-swap | Cache invalidation + natural loop iteration |
| Streaming | Phrase-level only via `GenMainIR` loop |
| Vibe switching | `currentVibe` tracked during parse, stamped on instruction |
