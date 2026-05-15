# fd3

Real-time card game server in Go.

## Package Overview

```
fd3/
├── main.go
├── matchmaking/          lobby and guest management
├── networking/           IP/network utilities
├── match/                game loop, event scheduling, input handling
│   └── eventpq/          EventPQ types and heap implementation
└── engine/
    ├── constants/        shared numeric constants (LEAD, REACTION_WINDOW, etc.)
    ├── state/            pure data types (GameState, Card, Chain, SimulationPQ, etc.)
    ├── defs/             card/structure definitions and effect templates
    │   ├── effects/
    │   └── primitives/
    ├── logs/             log/event struct types
    └── randomFunctions/  game logic functions (ScheduleDeclareCard, ScheduleDrawCard, etc.)
```

## Import Flow

```
match  ──────────────────────────────►  randomFunctions
  │                                           │
  │                                           ▼
  ├──────────────────────────────────►  state, defs, logs, constants
  │
  └──────────────────────────────────►  eventpq
```

`engine/` is the pure logic layer — it has no knowledge of `match`, clients, or networking.
`match/` is the orchestration layer; it owns the `EventPQ`, timers, the `Match` struct, and
client connections.

This means `randomFunctions` can update `GameState` but cannot directly call anything on
`Match` (e.g. `TransitionPauseState`). When game logic needs to trigger a match-level side
effect (like entering sim pause after starting a chain), it signals intent via a return value
and the `match` layer acts on it.
