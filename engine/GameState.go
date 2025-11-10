package engine

import "time"

type GameState struct {

	// time the state is set for
	GameStateID uint
	Timestamp   time.Time

	// state sets: array of unitIDs
	BenchUnits []uint
	FieldUnits []uint
	DeadUnits  []uint

	// team sets: index (unitID) ->  team index
	Team []uint

	// model set
	ModelNo []uint

	// core unit stats: index (unitID) ->  HP
	HP []int

	// telemetry
	TimesAttacked  []uint // index (unitID) -> val
	TimesDefended  []uint
	DamageDealt    []uint
	DamageReceived []uint

	// counter
	NextUnitID uint
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateGameState(t time.Time, GSID uint) *GameState {
	gm := &GameState{
		GameStateID:    GSID,
		Timestamp:      t,
		BenchUnits:     make([]uint, 0),
		FieldUnits:     make([]uint, 0),
		DeadUnits:      make([]uint, 0),
		Team:           make([]uint, 0),
		ModelNo:        make([]uint, 0),
		HP:             make([]int, 0),
		TimesAttacked:  make([]uint, 0),
		TimesDefended:  make([]uint, 0),
		DamageDealt:    make([]uint, 0),
		DamageReceived: make([]uint, 0),
		NextUnitID:     0,
	}
	return gm
}
