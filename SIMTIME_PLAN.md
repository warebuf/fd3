# Pause / SimTime Refactor

## Goal

Allow the simulation to be paused — for a player taking a break, or to hold open a reaction window after a card play. When paused, no events resolve. When unpaused, the simulation resumes exactly where it left off, and the reaction window extends by however long the pause lasted.

---

## The Core Idea

**SimTime = `wallTime - PauseOffset`**

All PopTimes and log timestamps are in SimTime space, not wall-clock time.
The server checks `PopTime <= simTime` instead of `PopTime <= time.Now()`.
The frontend checks `PopTime <= frontendSimTime` instead of `PopTime <= wallTime`.

Converting SimTime → WallTime: `wallTime = simTime + PauseOffset`

SimTime advances in lockstep with wallTime, lagging behind by PauseOffset.
When paused, PauseOffset grows at the same rate as wallTime — so SimTime stays frozen.
When unpaused, PauseOffset stops growing — so SimTime resumes advancing from where it froze.
Everything scheduled against SimTime naturally waits out any pause.

---

## State

### `GameState`

```go
IsPaused    bool   // whether the simulation is currently paused
PauseStart  int64  // wall-clock ms when the current pause began
PauseOffset int64  // total accumulated paused ms (updated on each unpause)
```

All three live on `GameState`. Clients use `IsPaused` and `PauseOffset` to maintain their own SimTime

Also add all three to `DeepCopyGameState`.

### `SimTime()` — free function

```go
func SimTime(gs *state.GameState) int64 {
    if gs.IsPaused {
        return gs.PauseStart - gs.PauseOffset  // frozen
    }
    return time.Now().UnixMilli() - gs.PauseOffset
}
```

---

## Pause / Unpause

```go
func (m *Match) Pause() {
    if m.GameState.IsPaused {
        return
    }
    m.GameState.IsPaused = true
    m.GameState.PauseStart = time.Now().UnixMilli()
    m.stopPQTimer()
    // append PauseLog, broadcast
}

func (m *Match) Unpause() {
    if !m.GameState.IsPaused {
        return
    }
    m.GameState.PauseOffset += time.Now().UnixMilli() - m.GameState.PauseStart
    m.GameState.IsPaused = false
    scheduleStackPQ(m)
    m.setPQTimer()
    // append ResumeLog, broadcast
}
```

---

## Server Changes

### `ManageStackPQ` — add simTime parameter

```go
// Before:
func ManageStackPQ(gs *state.GameState, logList *[]any, UpdateID *int) {
    if earliestStack.PopTime > time.Now().UnixMilli() {

// After:
func ManageStackPQ(gs *state.GameState, logList *[]any, UpdateID *int, simTime int64) {
    if earliestStack.PopTime > simTime {
```

SimTime is not stored on `GameState` — computed at the call site and passed in.

### Scheduling call sites — use SimTime instead of wallTime

Anywhere a PopTime or log timestamp is currently set from `time.Now()` or a raw wall-clock `popTime`, switch to `SimTime(gs)`:

```go
// Before (Phase 4 handler):
randomFunctions.ScheduleDeclareCard(..., ie.T.UnixMilli())

// After:
randomFunctions.ScheduleDeclareCard(..., SimTime(match.GameState))
```

Same applies inside `scheduleDeclareCardEffects`, `scheduleRecurringAbilities`, and any other function that stamps a PopTime or log timestamp.

### Effect handler in `EventSequencer`

Compute SimTime and pass it in. Skip if paused.

```go
case "Effect":
    if match.GameState.IsPaused {
        continue
    }
    randomFunctions.ManageStackPQ(match.GameState, &match.EventLog, &match.NextUpdateID, SimTime(match.GameState))
    // ... rest unchanged
```

### `scheduleStackPQ` — convert SimTime PopTime to WallTime for the timer

The Effect timer fires in wall time, but PopTimes are now SimTime. To convert:

```
wallTime = simTime + PauseOffset
```

So the timer should fire at `PopTime + PauseOffset`:

```go
// Before:
earliest := time.UnixMilli(earliestStack.PopTime)

// After:
earliest := time.UnixMilli(earliestStack.PopTime + match.GameState.PauseOffset)
```

Example: card played when PauseOffset = 3000ms (3s of prior pauses). PopTime = SimTime + REACTION_WINDOW = (wallTime - 3000) + 4000. Timer fires at PopTime + 3000 = wallTime + 4000 — exactly 4 seconds from now in wall time. ✓

On unpause, `scheduleStackPQ` is called again with the updated PauseOffset, so the timer always reflects the current accumulated pause.

---

## Frontend

The frontend maintains its own SimTime:

```
frontendSimTime = wallTime - totalPausedTime
```

- **PauseLog received**: record `pauseStart = wallTime`
- **ResumeLog received**: `totalPausedTime += wallTime - pauseStart`

All log timestamps are SimTime-based. To schedule an animation the frontend converts:

```
wallAnimationTime = logTimestamp + totalPausedTime
```

When paused, `totalPausedTime` grows, so `wallAnimationTime` shifts forward automatically. No timer-shifting needed.

---

## What stays the same

- The REACTION_WINDOW constant and all scheduling logic are unchanged
- Recurring abilities still fire in SimTime, extended by any pause duration
- Phase 3 auto-summon staggering is unaffected

## What gets wired up later

- A "confirm / pass priority" client message that calls `Unpause()` early
- Manual pause/unpause from a client UI
