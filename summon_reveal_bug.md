# Summon Bug in Reveal Phase — Analysis

## Summary

Unit card summons during the reveal phase (Phase 3) never resolve. Three bugs combine to cause this.

---

## Bug 1 — `scheduleEarliestSimItem` is never called after a unit is declared

**File**: `match/HandlersReveal.go:18-24`

```go
chainStarted := ScheduleDeclareCard(...)
if chainStarted {
    scheduleChainResolution(match)
} else {
    scheduleReveal(match, wallTime+800)
    // ← nothing here
}
```

When a unit is declared, `ScheduleDeclareCard` calls `scheduleSummon` which pushes a
`SimItemSummonResolve` into the SimulationPQ. But `scheduleEarliestSimItem` is never called,
so no `EffectEvent` is ever registered in the EventPQ. The summon timer never fires at all.

---

## Bug 2 — `ManageSimulationPQ` bails out during Phase 3

**File**: `engine/sim/ManageSimulationPQ.go:22`

```go
if gs.PausePhase != 2 {
    return
}
```

Phase 3 (reveal) runs with `PausePhase == 1` (sim pause, inherited from Phase 2's
`TransitionPauseState(1, ...)`). This guard means the SimPQ is never processed during the
entire reveal phase. Even if Bug 1 were fixed and an `EffectEvent` did fire, it would hit
this check and do nothing.

---

## Bug 3 — Phase 4 starts before summon timers can fire

**File**: `match/HandlersReveal.go:9-11`

```go
if len(match.RevealQueue) == 0 {
    PushPhase(match, wallTime+2000, 4)
    return
}
```

`REACTION_WINDOW = 4000ms`. A unit summon timer takes 4000ms to resolve. But Phase 4 is
pushed only 2000ms after the last reveal. Phase 4 always fires first.

When Phase 4 starts, `TransitionPauseState(-1, ...)` is called, which updates `SimPauseOffset`
by the full duration of the Phase 2+3 pause (potentially tens of seconds). After that update,
`scheduleEarliestSimItem` recalculates the EffectEvent wall time as:

```
new_wall_time = PopAtSimTime + new_SimPauseOffset
```

`PopAtSimTime` was computed with the old `SimPauseOffset = 0` as `wallTimeAtPop + 4000`.
After the update, `new_SimPauseOffset` could be 90,000+ ms (the board setup + reveal duration).
The result is an EffectEvent scheduled ~90 seconds too late.

**Concrete example**: Suppose Phase 2 started at T=5000ms and the last reveal fired at
T=10000ms. `SimPauseOffset = 0` throughout Phases 2-3. The summon item has `PopAtSimTime = 14000`.
Phase 4 starts at T=12000ms. `SimPauseOffset` becomes `12000 - 5000 = 7000`. The EffectEvent
is rescheduled to `14000 + 7000 = T=21000ms` — 9 seconds late and wrong.

---

## Why the math works for Phase 3

Before fixing, it's worth confirming the timing math is valid during Phase 3:

- `SimPauseOffset = 0` throughout Phases 2 and 3 (set at game start, never updated until Phase 4)
- In `scheduleSummon`: `simTime = wallTimeAtPop - 0 = wallTimeAtPop`
- `PopAtSimTime = wallTimeAtPop + REACTION_WINDOW`
- In `scheduleEarliestSimItem`: `effectEventWallTime = PopAtSimTime + SimPauseOffset = wallTimeAtPop + 4000`

So the EffectEvent would correctly fire 4 seconds after the reveal — the timing is right, the
only problems are structural (bugs 1, 2, 3).

---

## Proposed Fixes

### Fix 1 — Call `scheduleEarliestSimItem` after unit declaration

**File**: `match/HandlersReveal.go`

```go
} else {
    scheduleReveal(match, wallTime+800)
    scheduleEarliestSimItem(match)  // add this
}
```

`scheduleEarliestSimItem` is a no-op if the SimPQ is empty, so calling it unconditionally
in the else branch is safe.

---

### Fix 2 — Allow `ManageSimulationPQ` to run during Phase 3

**File**: `engine/sim/ManageSimulationPQ.go`

Change:
```go
if gs.PausePhase != 2 {
    return
}
```
To:
```go
if gs.PausePhase != 2 && !(gs.PausePhase == 1 && gs.Phase == 3) {
    return
}
```

The `currentSimTime` calculation (`time.Now().UnixMilli() - gs.SimPauseOffset`) remains
correct during Phase 3 because `SimPauseOffset = 0`, making sim time equal to wall time.

---

### Fix 3 — Delay Phase 4 past the last summon timer

**File**: `match/HandlersReveal.go`

Change:
```go
PushPhase(match, wallTime+2000, 4)
```
To:
```go
PushPhase(match, wallTime+2000+constants.REACTION_WINDOW, 4)
```
(requires importing `github.com/warebuf/fd3/engine/constants`)

This adds a 4000ms buffer so all outstanding summon timers (at most `REACTION_WINDOW` ms each)
complete before Phase 4 starts. Once Phase 4 begins, the SimPQ will be empty and the
`SimPauseOffset` update is harmless.

---

## Full trace after all three fixes

Given RevealQueue = [unit A, unit B] and Phase 3 starting at T=0:

| Time     | Event |
|----------|-------|
| T+500    | Reveal fires, declares unit A. SimPQ gets item (popAt=T+4500). EffectEvent scheduled for T+4500. Next reveal at T+1300. |
| T+1300   | Reveal fires, declares unit B. SimPQ gets item (popAt=T+5300). EffectEvent stays at T+4500 (A is still earliest). Next reveal at T+2100. |
| T+2100   | Reveal fires, queue empty. Phase 4 pushed for T+2100+6000 = T+8100. |
| T+4500   | EffectEvent fires. ManageSimulationPQ allowed (PausePhase=1, Phase=3). Unit A summon resolves, joins Field. SummonResolutionLog emitted. EffectEvent updated to T+5300. |
| T+5300   | EffectEvent fires. Unit B summon resolves, joins Field. SimPQ now empty. |
| T+8100   | Phase 4 starts. SimPQ is empty. SimPauseOffset update is harmless. |

---

## Notes / open questions

- The `PauseState.go` file has two `// do we have to remove any reveal items here?` comments
  in `TransitionPauseState`. After these fixes, the answer is: **no** for the normal path
  (Phase 4 starts after summons complete), but **yes** if a player triggers an indefinite pause
  (PausePhase 0) during Phase 3. That path would also break the RevealItem in the EventPQ.
  That's a separate issue not addressed here.

- `REACTION_WINDOW = 4000ms` feels long for a reveal-phase summon where there is no
  counter opportunity. Consider whether a shorter reveal summon window makes sense, which would
  also shorten the Phase 4 delay added by Fix 3.
