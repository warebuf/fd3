package matchmaking

import (
	"container/heap"
	"time"

	"github.com/google/uuid"

	"github.com/warebuf/fd3/engine/state"
	"github.com/warebuf/fd3/match"
	"github.com/warebuf/fd3/match/eventpq"
)

func CreateMatch(tickets []*Ticket, MatchMap map[string]*match.Match) *match.Match {

	// create matchID
	id := uuid.NewString()
	for {
		if _, exists := MatchMap[id]; !exists {
			break
		}
		id = uuid.NewString()
	}

	// create whitelist userIDs
	userIDs := make(map[string]uint)
	for i := 0; i < len(tickets); i++ {
		userIDs[tickets[i].TicketID] = uint(i)
	}

	// creating the first state, instantiating the card objects
	var deckDefs [][]int
	var structureDefNos []int
	for i := 0; i < len(tickets); i++ {
		deckDefs = append(deckDefs, tickets[i].Deck)
		structureDefNos = append(structureDefNos, tickets[i].StructureDefNo)
	}
	gs := DeckSetup(deckDefs, structureDefNos)

	// shuffle the decks, then arrange starting hands on top
	for i := 0; i < len(gs.Deck); i++ {
		ShuffleDeck(gs, i)
		ArrangeStartingHand(gs, i, tickets[i].StartingHand)
	}

	// create match object
	m := &match.Match{
		ID:        id,
		ByUser:    make(map[string]*match.Client),
		Whitelist: userIDs,

		GameState:      gs,
		PublishedState: state.DeepCopyGameState(gs),
		EventLog:       []any{},
		NextUpdateID:   1, // starts as 1 because the GameState already is UpdateID 0

		InputEventChan: make(chan *eventpq.InputEvent, 1024),
		EventPQTimer:   nil,
	}
	heap.Init(&m.EventPQueue) // init event handling PQ

	match.PushPhaseEvent(m, time.Now().UnixMilli()+500, state.PhaseDraw) // timer for starting the game, could also broadcast the time here

	go m.EventScheduler() // on init, check the game state, schedule if first if possible

	MatchMap[id] = m

	return m

}
