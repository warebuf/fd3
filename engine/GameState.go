package engine

import (
	"time"

	"github.com/warebuf/fd3/utility/shallowCopy"
)

type GameState struct {
	// time the state is set for
	GameStateID int
	Timestamp   time.Time
	Round       int

	// team sets: index (unitID) ->  team index
	Team []int
	Zone []int // 0: Bench, 1: Field, 2: Dead

	// model set
	ModelNo []int

	// core unit stats: index (unitID) ->  HP
	HP []int

	// abilities
	Abilities       []int   // abilities are flattened for each unit
	AbilityStartIdx []int   // the starting index in Abilities[] for this unit
	AbilityCount    []int   // how many abilities belong to each unit
	AbilityCharge   []int64 // saved in millisconds
	AbilitySPD      []int64 // in milliseconds

	SummonCondition []SummonType
}

type Telemetry struct {
	TimesAttacked  []int64 // index (unitID) -> val
	TimesDefended  []int64
	DamageDealt    []int64
	DamageReceived []int64
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID:     GSID,
		Timestamp:       t,
		Round:           0,
		Team:            make([]int, 0),
		Zone:            make([]int, 0),
		ModelNo:         make([]int, 0),
		HP:              make([]int, 0),
		Abilities:       make([]int, 0),
		AbilityStartIdx: make([]int, 0),
		AbilityCount:    make([]int, 0),
		AbilityCharge:   make([]int64, 0),
		AbilitySPD:      make([]int64, 0),
		SummonCondition: make([]SummonType, 0),
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
		Round:       prevState.Round,
	}

	newState.Team = shallowCopy.IntSlice(prevState.Team)
	newState.Zone = shallowCopy.IntSlice(prevState.Zone)
	newState.ModelNo = shallowCopy.IntSlice(prevState.ModelNo)
	newState.HP = shallowCopy.IntSlice(prevState.HP)
	newState.Abilities = shallowCopy.IntSlice(prevState.Abilities)
	newState.AbilityStartIdx = shallowCopy.IntSlice(prevState.AbilityStartIdx)
	newState.AbilityCount = shallowCopy.IntSlice(prevState.AbilityCount)
	newState.AbilityCharge = shallowCopy.Int64Slice(prevState.AbilityCharge)
	newState.AbilitySPD = shallowCopy.Int64Slice(prevState.AbilitySPD)
	newState.SummonCondition = CopySummonTypeSlice(prevState.SummonCondition)

	return newState
}
