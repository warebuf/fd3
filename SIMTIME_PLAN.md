# SimTime Refactor

## Goal

Replace wall-clock time (`time.Now()`) with a simulation clock (`gs.SimTime`) so chain reactions resolve immediately rather than waiting for real time to pass.

## The Core Idea

Right now, chain PopTimes are absolute Unix ms timestamps, and `ManageStackPQ` only processes a chain when `time.Now()` has caught up to its `PopTime`. This means the server literally waits 4 real seconds (REACTION_WINDOW) before resolving a card play.

With SimTime, `gs.SimTime` is the only clock. The server advances it explicitly, and `ManageStackPQ` processes everything with `PopTime <= gs.SimTime` immediately.

---

## Changes

### 1. `engine/state/GameState.go`

Add `SimTime int64` to the `GameState` struct:

```go
type GameState struct {
    SimTime   int64      // current simulation time in Unix ms
    UpdateID  int
    Timestamp time.Time
    // ... rest unchanged
}
```

Also add `SimTime` to `DeepCopyGameState`:
```go
newState := &GameState{
    SimTime:   prevState.SimTime,
    UpdateID:  prevState.UpdateID,
    // ... rest unchanged
}
```

---

### 2. `engine/randomFunctions/ManageStackPQ.go`

Change the due-check from `time.Now()` to `gs.SimTime`:

```go
// Before:
if earliestStack.PopTime > time.Now().UnixMilli() {
    return
}

// After:
if earliestStack.PopTime > gs.SimTime {
    return
}
```

---

### 3. `engine/randomFunctions/ScheduleDeclareCard.go`

Change `scheduleDeclareCardEffects` to base PopTimes on `gs.SimTime` instead of the `popTime` parameter:

```go
// Before (in scheduleDeclareCardEffects):
PopTime: popTime + constants.LEAD + constants.REACTION_WINDOW - constants.LEAD,

// After:
PopTime: gs.SimTime + constants.REACTION_WINDOW,
```

Same for the ChainStep's `StartTime`, `PopTime`, `OccurTime` — base them on `gs.SimTime`.

`ScheduleDeclareCard` needs to accept `gs *state.GameState` if it doesn't already, so it can read `gs.SimTime`.

---

### 4. `match/EventSequencer.go` — Effect event handler

Set `gs.SimTime` from wall clock before calling `ManageStackPQ`. This keeps recurring abilities working via the real-time timer:

```go
case "Effect":
    oldLogsSize := len(match.EventLog)
    match.GameState.SimTime = time.Now().UnixMilli()  // ADD THIS
    randomFunctions.ManageStackPQ(match.GameState, &match.EventLog, &match.NextUpdateID)
    // ... rest unchanged
```

---

### 5. `match/EventSequencer.go` — Phase 4 card play handler

After `ScheduleDeclareCard`, advance SimTime by REACTION_WINDOW and call `ManageStackPQ` inline so the chain resolves immediately (no background wait):

```go
case 4: // play
    if ie.Type != "hand" {
        continue
    }
    oldLogsSize := len(match.EventLog)

    match.GameState.SimTime = ie.T.UnixMilli()  // set sim clock to "now"
    randomFunctions.ScheduleDeclareCard(match.GameState, &match.EventLog, &match.NextUpdateID, ie.TeamIdx, ie.CardIndex, ie.T.UnixMilli())

    if len(match.EventLog) > oldLogsSize {
        match.NextUpdateID += 1

        match.GameState.SimTime = ie.T.UnixMilli() + constants.REACTION_WINDOW  // skip ahead — resolve immediately
        randomFunctions.ManageStackPQ(match.GameState, &match.EventLog, &match.NextUpdateID)

        match.GameState.Timestamp = next.Event.Timestamp().Add(constants.LEAD * time.Millisecond)
        match.GameState.UpdateID = match.NextUpdateID
        match.NextUpdateID += 1
        match.PublishedState = state.DeepCopyGameState(match.GameState)
        match.BroadcastLogs(oldLogsSize)
    }
```

---

## What stays the same

- The Effect timer + `scheduleStackPQ` mechanism still exists and handles **recurring abilities** (they fire in real-time via the timer, which sets `SimTime = time.Now()`)
- Log timestamps are unchanged — the client still animates events spread across the reaction window
- The server just no longer waits for real time to pass for card-play resolutions
