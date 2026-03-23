package states

import (
	"time"
)

type GameState struct {

	//
	GameStateID int

	// time the state is set for
	Timestamp time.Time

	// additional info on the state
	Round int

	//
}

type Card struct {
	CID   int // card ID
	DefNo int // the card's serial number
	Team  int
	Type  int // structure, unit, trap, spell
	Zone  int // deck, hand, field, graveyard
	Armed bool
}

type Permanents struct {
	PID        int
	SourceCard int
	Type       int // structure, unit, continuous trap, continuous spell
	HP         int
	Team       int

	// abilities?
}

type Ability struct {
	EffectNo  int
	Type      int // 0: auto, 1: manual, 2: passive
	UnitIndex int

	Charge  int // in milliseconds
	SPD     int // in milliseconds
	Pending bool
	Target  [5]int16
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID:     GSID,
		Timestamp:       t,
		Round:           0,
		StructureHP:     make([]int, 0),
		StructureTeam:   make([]int, 0),
		Team:            make([]int, 0),
		Zone:            make([]int, 0),
		ModelNo:         make([]int, 0),
		HP:              make([]int, 0),
		EXP:             make([]int, 0),
		SummonCharge:    make([]int, 0),
		Abilities:       make([]Ability, 0),
		AbilityCount:    make([]int, 0),
		AbilityStartIdx: make([]int, 0),
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

	newState.StructureHP = copyIntSlice(prevState.StructureHP)
	newState.StructureTeam = copyIntSlice(prevState.StructureTeam)

	newState.Team = copyIntSlice(prevState.Team)
	newState.Zone = copyIntSlice(prevState.Zone)

	//
	newState.ModelNo = copyIntSlice(prevState.ModelNo)

	//
	newState.HP = copyIntSlice(prevState.HP)
	newState.EXP = copyIntSlice(prevState.EXP)

	//
	newState.SummonCharge = copyIntSlice(prevState.SummonCharge)

	//
	newState.Abilities = copyAbilitySlice(prevState.Abilities)
	newState.AbilityCount = copyIntSlice(prevState.AbilityCount)
	newState.AbilityStartIdx = copyIntSlice(prevState.AbilityStartIdx)

	return newState
}

// helpers

func copyAbilitySlice(src []Ability) []Ability {
	if src == nil {
		return nil
	}
	dst := make([]Ability, len(src))
	copy(dst, src)
	return dst
}

func copyIntSlice(src []int) []int {
	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}
