package state

import "github.com/warebuf/fd3/engine/constants"

const (
	PhasePreMatch = 0
	PhaseDraw     = 1
	PhasePlay     = 2
	PhaseGameOver = 3
)

const PauseTransitionUnpause = -1
const (
	PausePhaseFullPause  = 0
	PausePhaseSimPause   = 1
	PausePhaseSimulating = 2
)

type GameState struct {
	UpdateID      int
	WallTimestamp int64 // wall clock time when this state was last updated

	Phase                int   // 0: pre-match, 1: draw, 2: play, 3: game over
	PhaseEndInActiveTime int64 // active time (ms) when the current phase ends; 0 if no fixed end

	// pause related
	PausePhase         int     // 0: indefinite pause, 1: sim pause, 2: simulating
	FullPauseStartedAt int64   // denominated in wall time
	SimPauseStartedAt  int64   // denominated in wall time
	FullPauseOffset    int64   // total accumulated full-pause ms relative to wall ms
	SimPauseOffset     int64   // total accumulated sim-pause ms relative to wall ms
	TeamPaused         int     // which team paused the game; -1 means no team is paused
	TimeBank           []int64 // remaining pause ms per team

	// card states
	Cards []Card
	Deck  [][]int // front of arr is the top
	Hand  [][]int
	Field [][]int
	Grave [][]int
	Traps [][]int // per team, traps that have been set but not yet activated
	Equip [][]int // per team, equip cards attached to a host card

	Chain []ChainItem // LIFO, sim paused while non-empty

	// other states
	Mana       []int // per team
	Structures [constants.NUM_TEAMS][constants.STRUCTURES_PER_TEAM]Structure
	Ready      []bool // per team, whether the player has readied up
	SkipChain  []bool // per team, waived reacting to the current chain top; cleared whenever the chain changes

	// Standoff-style blind mana auction; see match.handleBid. -1/[-1,-1] when no auction is open.
	// CensorForTeam hides an opponent's still-sealed bid before this ever reaches their client.
	StandoffCardIndex int
	PendingBidAmount  []int // per team; -1 means not yet sealed for the current auction

	Winner int // -1: no winner yet, 0+: winning team index, -2: tie
}

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		UpdateID:             prevState.UpdateID,
		WallTimestamp:        prevState.WallTimestamp,
		Phase:                prevState.Phase,
		PhaseEndInActiveTime: prevState.PhaseEndInActiveTime,
		PausePhase:           prevState.PausePhase,
		FullPauseStartedAt:   prevState.FullPauseStartedAt,
		SimPauseStartedAt:    prevState.SimPauseStartedAt,
		FullPauseOffset:      prevState.FullPauseOffset,
		SimPauseOffset:       prevState.SimPauseOffset,
		TeamPaused:           prevState.TeamPaused,
		Winner:               prevState.Winner,
		Structures:           prevState.Structures, // a fixed-size array of values, so this copy is already deep
		StandoffCardIndex:    prevState.StandoffCardIndex,
	}

	// Cards
	newState.Cards = make([]Card, len(prevState.Cards))
	copy(newState.Cards, prevState.Cards)

	// Deck
	newState.Deck = make([][]int, len(prevState.Deck))
	for i := range prevState.Deck {
		newState.Deck[i] = make([]int, len(prevState.Deck[i]))
		copy(newState.Deck[i], prevState.Deck[i])
	}

	// Hand
	newState.Hand = make([][]int, len(prevState.Hand))
	for i := range prevState.Hand {
		newState.Hand[i] = make([]int, len(prevState.Hand[i]))
		copy(newState.Hand[i], prevState.Hand[i])
	}

	// Field
	newState.Field = make([][]int, len(prevState.Field))
	for i := range prevState.Field {
		newState.Field[i] = make([]int, len(prevState.Field[i]))
		copy(newState.Field[i], prevState.Field[i])
	}

	// Grave
	newState.Grave = make([][]int, len(prevState.Grave))
	for i := range prevState.Grave {
		newState.Grave[i] = make([]int, len(prevState.Grave[i]))
		copy(newState.Grave[i], prevState.Grave[i])
	}

	// Traps
	newState.Traps = make([][]int, len(prevState.Traps))
	for i := range prevState.Traps {
		newState.Traps[i] = make([]int, len(prevState.Traps[i]))
		copy(newState.Traps[i], prevState.Traps[i])
	}

	// Equip
	newState.Equip = make([][]int, len(prevState.Equip))
	for i := range prevState.Equip {
		newState.Equip[i] = make([]int, len(prevState.Equip[i]))
		copy(newState.Equip[i], prevState.Equip[i])
	}

	// Mana
	newState.Mana = make([]int, len(prevState.Mana))
	copy(newState.Mana, prevState.Mana)

	// Ready
	newState.Ready = make([]bool, len(prevState.Ready))
	copy(newState.Ready, prevState.Ready)

	// SkipChain
	newState.SkipChain = make([]bool, len(prevState.SkipChain))
	copy(newState.SkipChain, prevState.SkipChain)

	// TimeBank
	newState.TimeBank = make([]int64, len(prevState.TimeBank))
	copy(newState.TimeBank, prevState.TimeBank)

	// PendingBidAmount
	newState.PendingBidAmount = make([]int, len(prevState.PendingBidAmount))
	copy(newState.PendingBidAmount, prevState.PendingBidAmount)

	// Chain
	newState.Chain = make([]ChainItem, len(prevState.Chain))
	for i, item := range prevState.Chain {
		newItem := ChainItem{
			Type:                item.Type,
			CardIndex:           item.CardIndex,
			ActiveIndex:         item.ActiveIndex,
			TeamIndex:           item.TeamIndex,
			StructureIndex:      item.StructureIndex,
			StartAtActiveTime:   item.StartAtActiveTime,
			PopAtActiveTime:     item.PopAtActiveTime,
			ExecuteAtActiveTime: item.ExecuteAtActiveTime,
			Effects:             make([]Effect, len(item.Effects)),
		}
		copy(newItem.Effects, item.Effects)
		newState.Chain[i] = newItem
	}

	return newState
}
