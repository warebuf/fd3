package engine

import (
	"time"

	"github.com/warebuf/fd3/utility/shallowCopy"
)

type GameState struct {

	//
	GameStateID int

	// time the state is set for
	Timestamp time.Time

	// additional info on the state
	Round int

	// structure data
	StructureHP   []int
	StructureTeam []int

	// team sets: index (unitID) ->  team index
	Team []int
	Zone []int // 0: Bench, 1: Field, 2: Dead, 3: QueuedSummon

	// model set
	ModelNo []int

	// core unit stats: index (unitID) ->  HP
	HP  []int
	EXP []int

	// summon resolution
	SummonCharge []int // in milliseconds

	// abilities
	Abilities       []Ability // abilities are flattened for each unit
	AbilityCount    []int     // how many functions belong to each unit
	AbilityStartIdx []int     // the starting index in Functions[] for this unit

}

type Ability struct {
	AbilityNo int
	Type      int // 0: auto, 1: manual, 2: passive
	UnitIndex int

	Charge int // in milliseconds
	SPD    int // in milliseconds
	Target [5]int16
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID:      GSID,
		Timestamp:        t,
		Round:            0,
		StructureHP:      make([]int, 0),
		StructureTeam:    make([]int, 0),
		Team:             make([]int, 0),
		Zone:             make([]int, 0),
		ModelNo:          make([]int, 0),
		HP:               make([]int, 0),
		EXP:              make([]int, 0),
		SummonCharge:     make([]int, 0),
		Functions:        make([]int, 0),
		FunctionType:     make([]int, 0),
		FunctionStartIdx: make([]int, 0),
		FunctionCount:    make([]int, 0),
		FunctionCharge:   make([]int, 0),
		FunctionSPD:      make([]int, 0),
		FunctionTarget:   make([]int, 0),
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

	newState.StructureHP = shallowCopy.IntSlice(prevState.StructureHP)
	newState.StructureTeam = shallowCopy.IntSlice(prevState.StructureTeam)

	newState.Team = shallowCopy.IntSlice(prevState.Team)
	newState.Zone = shallowCopy.IntSlice(prevState.Zone)

	//
	newState.ModelNo = shallowCopy.IntSlice(prevState.ModelNo)

	//
	newState.HP = shallowCopy.IntSlice(prevState.HP)
	newState.EXP = shallowCopy.IntSlice(prevState.EXP)

	//
	newState.SummonCharge = shallowCopy.IntSlice(prevState.SummonCharge)

	//
	newState.Functions = shallowCopy.IntSlice(prevState.Functions)
	newState.FunctionType = shallowCopy.IntSlice(prevState.FunctionType)
	newState.FunctionStartIdx = shallowCopy.IntSlice(prevState.FunctionStartIdx)
	newState.FunctionCount = shallowCopy.IntSlice(prevState.FunctionCount)
	newState.FunctionCharge = shallowCopy.IntSlice(prevState.FunctionCharge)
	newState.FunctionSPD = shallowCopy.IntSlice(prevState.FunctionSPD)
	newState.FunctionTarget = shallowCopy.IntSlice(prevState.FunctionTarget)

	return newState
}
