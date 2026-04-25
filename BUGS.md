# Known Bugs (found 2026-04-24)

1. **Mutex deadlock** (`Matchmaking.go:62,71`)
   `continue` inside the mutex lock skips `m.mutex.Unlock()`, deadlocking on the next join.

2. **Data race on `matchMap`** (`main.go:265`)
   Global map written by matchmaking goroutine, read by HTTP handlers, no synchronization.

3. **Data race on `match.ByUser`** (`Match.go`)
   HTTP handler writes it while EventScheduler goroutine reads it in Broadcast methods.

4. **Duplicate method check** (`main.go:41-49`)
   `r.Method != http.MethodPost` check is copy-pasted; second block is dead code.

5. **Duplicate ops in matchmaking** (`Matchmaking.go:100-113`)
   `Found <-`, `queue.Delete`, and `inGame[...]` each called twice per matched player.

6. **JSON tag mismatch** (`ClientMessageTypes.go:4`)
   `handMessage.Type` has tag `json:"Type"` (capital T) but outer type-sniff uses `json:"type"` (lowercase). `strictUnmarshal` with `DisallowUnknownFields` will always reject hand messages.

7. **Typo** (`Client.go:101,110`)
   "closing White client" should be "closing Write client".

8. **Debug print** (`EventSequencer.go:53`)
   `match.PQ.Print()` fires on every due event — leftover WIP logging.
