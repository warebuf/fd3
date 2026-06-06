package state

const (
	PhasePreMatch   = 0
	PhaseDraw       = 1
	PhaseBoardSetup = 2
	PhaseReveal     = 3
	PhasePlay       = 4
	PhaseGameOver   = 5
)

const (
	PausePhaseFullPause  = 0
	PausePhaseSimPause   = 1
	PausePhaseSimulating = 2
)

type GameState struct {
	UpdateID      int
	WallTimestamp int64 // wall clock time when this state was last updated

	Phase int // 0: pre-match, 1: draw, 2: board setup, 3: reveal, 4: play, 5: game over

	// pause related
	PausePhase      int      // 0: indefinite pause, 1: sim pause, 2: simulating
	PauseStack      [2]int64 // the wall ms when each pause started
	FullPauseOffset int64    // total accumulated full-pause ms relative to wall ms
	SimPauseOffset  int64    // total accumulated sim-pause ms relative to wall ms
	TeamPaused      int      // which team paused the game; -1 means no team is paused
	TimeBank        []int64  // remaining pause ms per team

	// card states
	Cards []Card
	Deck  [][]int // front of arr is the top
	Hand  [][]int
	Field [][]int
	Grave [][]int

	Stack []StackItem // interactive stack: spells/traps/abilities, LIFO, sim paused while non-empty

	// other states
	Mana       []int       // per team
	Structures []Structure // per team
	Ready      []bool      // per team, whether the player has readied up

	Winner int // -1: no winner yet, 0+: winning team index, -2: tie
}

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		UpdateID:        prevState.UpdateID,
		WallTimestamp:   prevState.WallTimestamp,
		Phase:           prevState.Phase,
		PausePhase:      prevState.PausePhase,
		PauseStack:      prevState.PauseStack,
		FullPauseOffset: prevState.FullPauseOffset,
		SimPauseOffset:  prevState.SimPauseOffset,
		TeamPaused:      prevState.TeamPaused,
		Winner:          prevState.Winner,
	}

	// Cards
	newState.Cards = make([]Card, len(prevState.Cards))
	for i, card := range prevState.Cards {
		newState.Cards[i] = card
		if len(card.AbilityTimers) > 0 {
			newState.Cards[i].AbilityTimers = make([]AbilityTimer, len(card.AbilityTimers))
			copy(newState.Cards[i].AbilityTimers, card.AbilityTimers)
		}
	}

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

	// Mana
	newState.Mana = make([]int, len(prevState.Mana))
	copy(newState.Mana, prevState.Mana)

	// Structures
	newState.Structures = make([]Structure, len(prevState.Structures))
	copy(newState.Structures, prevState.Structures)

	// Ready
	newState.Ready = make([]bool, len(prevState.Ready))
	copy(newState.Ready, prevState.Ready)

	// TimeBank
	newState.TimeBank = make([]int64, len(prevState.TimeBank))
	copy(newState.TimeBank, prevState.TimeBank)

	// Stack
	newState.Stack = make([]StackItem, len(prevState.Stack))
	for i, item := range prevState.Stack {
		newItem := StackItem{
			Type:                item.Type,
			CardID:              item.CardID,
			StartAtActiveTime:   item.StartAtActiveTime,
			PopAtActiveTime:     item.PopAtActiveTime,
			ExecuteAtActiveTime: item.ExecuteAtActiveTime,
			Effects:             make([]Effect, len(item.Effects)),
		}
		copy(newItem.Effects, item.Effects)
		newState.Stack[i] = newItem
	}

	return newState
}
