package state

type GameState struct {
	UpdateID int
	SimTime  int64 // current sim time ms

	Phase int // 0: pre-match, 1: draw, 2: board setup, 3: reveal, 4: play, 5: game end

	// full pause: everything stops (player pause, animation pause)
	IsFullPaused    bool
	FullPauseStart  int64 // wall ms when current full pause began
	FullPauseOffset int64 // total accumulated full-pause ms

	// sim pause: passive effects freeze while interactive chain runs
	IsSimPaused    bool
	SimPauseStart  int64 // chain time ms when current sim pause began
	SimPauseOffset int64 // total accumulated sim-pause ms (chain time units)

	// card states
	Cards []Card
	Deck  [][]int
	Hand  [][]int // front of arr is the top
	Field [][]int
	Grave [][]int

	// interactive chain: spells/traps/abilities, LIFO stack, sim paused while non-empty
	Chain []ChainItem

	// passive effects: summon windows, draws, recurring abilities; sim time PopTimes
	ReactionPQ PriorityQueue

	// other states
	Mana       []int       // per team
	Structures []Structure // per team
	Ready      []bool      // per team, whether the player has readied up
}

func DeepCopyGameState(prevState *GameState) *GameState {
	newState := &GameState{
		UpdateID:        prevState.UpdateID,
		SimTime:         prevState.SimTime,
		Phase:           prevState.Phase,
		IsFullPaused:    prevState.IsFullPaused,
		FullPauseStart:  prevState.FullPauseStart,
		FullPauseOffset: prevState.FullPauseOffset,
		IsSimPaused:     prevState.IsSimPaused,
		SimPauseStart:   prevState.SimPauseStart,
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

	// ReactionPQ (passive effects)
	newPQ := PriorityQueue{
		total: prevState.ReactionPQ.total,
		Items: make([]*EffectItem, len(prevState.ReactionPQ.Items)),
	}
	for i, item := range prevState.ReactionPQ.Items {
		newItem := &EffectItem{
			EffectIndex: item.EffectIndex,
			PopTime:     item.PopTime,
			index:       item.index,
			CardID:      item.CardID,
			Effects:     make([]Effect, len(item.Effects)),
		}
		copy(newItem.Effects, item.Effects)
		newPQ.Items[i] = newItem
	}
	newState.ReactionPQ = newPQ

	return newState
}
