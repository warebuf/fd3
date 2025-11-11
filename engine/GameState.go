package engine

import (
	"time"

	copyslice "github.com/warebuf/fd3/utility/copySlice"
)

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

func CopyGameState(timestamp time.Time, prevState *GameState) *GameState {
	if prevState == nil {
		return nil
	}

	newState := &GameState{
		GameStateID: prevState.GameStateID + 1, // increment ID for new state
		Timestamp:   timestamp,
		NextUnitID:  prevState.NextUnitID,
	}

	newState.BenchUnits = copyslice.CopySliceUint(prevState.BenchUnits)
	newState.FieldUnits = copyslice.CopySliceUint(prevState.FieldUnits)
	newState.DeadUnits = copyslice.CopySliceUint(prevState.DeadUnits)
	newState.Team = copyslice.CopySliceUint(prevState.Team)
	newState.ModelNo = copyslice.CopySliceUint(prevState.ModelNo)
	newState.HP = copyslice.CopySliceInt(prevState.HP)
	newState.TimesAttacked = copyslice.CopySliceUint(prevState.TimesAttacked)
	newState.TimesDefended = copyslice.CopySliceUint(prevState.TimesDefended)
	newState.DamageDealt = copyslice.CopySliceUint(prevState.DamageDealt)
	newState.DamageReceived = copyslice.CopySliceUint(prevState.DamageReceived)

	return newState
}
