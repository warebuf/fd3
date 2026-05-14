package state

type GameState struct {
	UpdateID      int
	WallTimestamp int64 // wall clock time when this state was last updated

	Phase int // 0: pre-match, 1: draw, 2: board setup, 3: reveal, 4: play, 5: game end

	PauseStack      [2]int64 // the wall ms when each pause started (3 pause phases: indefinite pause, sim pause, simulating)
	FullPauseOffset int64    // total accumulated full-pause ms relative to wall ms
	SimPauseOffset  int64    // total accumulated sim-pause ms relative to wall ms

	// card states
	Cards []Card
	Deck  [][]int
	Hand  [][]int // front of arr is the top
	Field [][]int
	Grave [][]int

	Chain        []ChainItem  // interactive chain: spells/traps/abilities, LIFO stack, sim paused while non-empty
	SimulationPQ SimulationPQ // passive effects: summon windows, draws, recurring abilities; sim time PopTimes

	// other states
	Mana       []int       // per team
	Structures []Structure // per team
	Ready      []bool      // per team, whether the player has readied up
}

func (gs *GameState) GetPauseMode() int {
	if gs.PauseStack[0] != 0 { // indefinite pause
		return 0
	}
	if gs.PauseStack[1] != 0 { // sim pause
		return 1
	}
	return 2 // currently simulating
}

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		UpdateID:        prevState.UpdateID,
		WallTimestamp:   prevState.WallTimestamp,
		Phase:           prevState.Phase,
		PauseStack:      prevState.PauseStack,
		FullPauseOffset: prevState.FullPauseOffset,
		SimPauseOffset:  prevState.SimPauseOffset,
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

	// Mana
	newState.Mana = make([]int, len(prevState.Mana))
	copy(newState.Mana, prevState.Mana)

	// Structures
	newState.Structures = make([]Structure, len(prevState.Structures))
	copy(newState.Structures, prevState.Structures)

	// Ready
	newState.Ready = make([]bool, len(prevState.Ready))
	copy(newState.Ready, prevState.Ready)

	// Chain (interactive stack)
	newState.Chain = make([]ChainItem, len(prevState.Chain))
	for i, item := range prevState.Chain {
		newItem := ChainItem{
			Type:      item.Type,
			CardID:    item.CardID,
			StartTime: item.StartTime,
			PopTime:   item.PopTime,
			OccurTime: item.OccurTime,
			Effects:   make([]Effect, len(item.Effects)),
		}
		copy(newItem.Effects, item.Effects)
		newState.Chain[i] = newItem
	}

	// SimulationPQ (passive effects)
	newPQ := SimulationPQ{
		total: prevState.SimulationPQ.total,
		Items: make([]*SimulationItem, len(prevState.SimulationPQ.Items)),
	}
	for i, item := range prevState.SimulationPQ.Items {
		newItem := &SimulationItem{
			SeqID:   item.SeqID,
			PopTime: item.PopTime,
			index:   item.index,
			CardID:  item.CardID,
			Effects: make([]Effect, len(item.Effects)),
		}
		copy(newItem.Effects, item.Effects)
		newPQ.Items[i] = newItem
	}
	newState.SimulationPQ = newPQ

	return newState
}
