package state

import (
	"container/heap"
)

func (pq *SimulationPQ) RemoveSimItemsByCardIDs(cardIDs map[int]bool) []int {
	var toRemove []*SimulationItem
	for _, item := range pq.Items {
		if cardIDs[item.CardID] {
			toRemove = append(toRemove, item)
		}
	}
	var removedSeqIDs []int
	for _, item := range toRemove {
		removedSeqIDs = append(removedSeqIDs, item.SeqID)
		heap.Remove(pq, item.index)
	}
	return removedSeqIDs
}

func (pq *SimulationPQ) RemoveSimItemByCardID(cardID int) []int {
	var toRemove []*SimulationItem
	for _, item := range pq.Items {
		if item.CardID == cardID {
			toRemove = append(toRemove, item)
		}
	}
	var removedSeqIDs []int
	for _, item := range toRemove {
		removedSeqIDs = append(removedSeqIDs, item.SeqID)
		heap.Remove(pq, item.index)
	}
	return removedSeqIDs
}

const (
	SimItemSummon = iota
	SimItemRecurring
)

type SimulationItem struct {
	SeqID int // the latest effect gets the biggest index
	index int // heap position (internal)

	StartTime int64 // (denominated in SimTime ms)
	PopTime   int64 // the time we can apply to the game state (denominated in SimTime ms)
	OccurTime int64 // (denominated in SimTime ms)

	CardID  int
	Type    int
	Effects []Effect
}

type SimulationPQ struct {
	Items []*SimulationItem
	total int
}

func (pq SimulationPQ) Len() int { return len(pq.Items) }

func (pq SimulationPQ) Less(i, j int) bool {
	if pq.Items[i].PopTime != pq.Items[j].PopTime {
		return pq.Items[i].PopTime < pq.Items[j].PopTime
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
	item.PopTime = newPopTime
	heap.Fix(pq, item.index)
}

func (pq SimulationPQ) Peek() *SimulationItem {
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}
