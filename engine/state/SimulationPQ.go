package state

import (
	"container/heap"
)

type SimulationItem struct {
	SeqID int // the latest effect gets the biggest index
	index int // heap position (internal)

	StartTime int64 // (denominated in SimTime ms)
	PopTime   int64 // the time we can apply to the game state (denominated in SimTime ms)
	OccurTime int64 // (denominated in SimTime ms)

	CardID  int
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
