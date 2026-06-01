package state

import (
	"container/heap"
)

func (pq *SimulationPQ) RemoveSimItemsByCardIDs(cardIDs map[int]bool) {
	var toRemove []*SimulationItem
	for _, item := range pq.Items {
		if cardIDs[item.CardID] {
			toRemove = append(toRemove, item)
		}
	}
	for _, item := range toRemove {
		heap.Remove(pq, item.index)
	}
}

func (pq *SimulationPQ) RemoveSimItemByCardID(cardID int) {
	var toRemove []*SimulationItem
	for _, item := range pq.Items {
		if item.CardID == cardID {
			toRemove = append(toRemove, item)
		}
	}
	for _, item := range toRemove {
		heap.Remove(pq, item.index)
	}
}

type SimItemType int

const (
	SimItemSummonResolve SimItemType = iota
	SimItemRecurring
	SimItemDraw
	SimItemStructureRegen
)

type SimulationItem struct {
	SeqID     int // PQ tiebreaker (internal)
	AbilityID int // references the AbilityTimer on the card; -1 for non-ability items
	index     int // heap position (internal)

	StartAtSimTime   int64 // when the countdown starts for this item (denominated in SimTime ms)
	PopAtSimTime     int64 // the time we can apply to the game state (denominated in SimTime ms)
	ExecuteAtSimTime int64 // the time the item executes (denominated in SimTime ms)

	CardID  int
	Type    SimItemType
	Effects []Effect
}

type SimulationPQ struct {
	Items []*SimulationItem
	total int
}

func (pq SimulationPQ) Len() int { return len(pq.Items) }

func (pq SimulationPQ) Less(i, j int) bool {
	if pq.Items[i].PopAtSimTime != pq.Items[j].PopAtSimTime {
		return pq.Items[i].PopAtSimTime < pq.Items[j].PopAtSimTime
	}
	return pq.Items[i].SeqID < pq.Items[j].SeqID
}

func (pq SimulationPQ) Swap(i, j int) {
	pq.Items[i], pq.Items[j] = pq.Items[j], pq.Items[i]
	pq.Items[i].index = i
	pq.Items[j].index = j
}

func (pq *SimulationPQ) Push(x any) {
	item := x.(*SimulationItem)
	item.SeqID = pq.total
	pq.total++
	item.index = len(pq.Items)
	pq.Items = append(pq.Items, item)
}

func (pq *SimulationPQ) Pop() any {
	old := pq.Items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	pq.Items = old[:n-1]
	return item
}

func (pq *SimulationPQ) Update(item *SimulationItem, newPopTime int64) {
	if item == nil || item.index < 0 || item.index >= len(pq.Items) {
		return
	}
	item.PopAtSimTime = newPopTime
	heap.Fix(pq, item.index)
}

func (pq SimulationPQ) Peek() *SimulationItem {
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}
