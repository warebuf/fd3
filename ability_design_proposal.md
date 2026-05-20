# Ability Design Proposal

## Problem

`SimulationPQ` currently lives in `GameState`, which means it gets sent to clients on full state refresh, and every `AddSimItemLog`/`RemoveSimItemLog` leaks internal scheduling details to the frontend — including things like draw timers that the frontend has no use for.

## Proposed Design

### Pull `SimulationPQ` out of shared state

`SimulationPQ` becomes a pure backend scheduling mechanism. It is no longer part of the synced game state. `AddSimItemLog` and `RemoveSimItemLog` are removed from the log protocol for card abilities.

Non-card sim items (draws, `CardID == -1`) stay as pure PQ internals with no frontend-visible state and no logs needed.

### `AbilityTimer` slice on `Card`

Each card carries a fixed-size array of `AbilityTimer` entries — one per currently active/pending ability instance. Fixed size (capped at `MaxAbilitiesPerCard`) for memory continuity; empty slots marked by sentinel (e.g. `Active bool`).

```go
type AbilityTimer struct {
    Active     bool
    SeqID      int   // unique instance ID — matches SimulationItem.SeqID
    AbilityIdx int   // index into card def's ability list (what to display)
    StartTime  int64
    OccurTime  int64
}
```

- `SeqID` is the stable instance handle. Two instances of the same ability (e.g. granted twice) are distinguished by `SeqID`, not `AbilityIdx`.
- When pushing to SimulationPQ, also write an `AbilityTimer` entry on the card.
- When the item pops and fires, match on `SeqID` to clear the right slot.

### Ability types

- **Recurring**: gets an `AbilityTimer` entry as soon as scheduled; frontend shows countdown.
- **On-summon / one-shot**: gets an `AbilityTimer` entry when scheduled, cleared on fire.
- **Triggered (on-death, when X happens)**: no `AbilityTimer` until the trigger condition is met; countdown only appears when the ability is actually pending.

### Frontend sync

Full state refresh naturally includes all timing info since it lives on the card. No special log messages needed for ability timers — card state updates cover it.
