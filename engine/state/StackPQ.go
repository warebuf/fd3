package state

import (
	"container/heap"
)

type Stack struct {
	EffectIndex int // the latest effect gets the biggest index

	PopTime int64 // the time we can pop an item from the stack and apply it to the game state

	Effects []Effect

	index int // heap position (internal)
}

type PriorityQueue struct {
	items []*Stack
	total int
}

func (pq PriorityQueue) Len() int { return len(pq.items) }

// earliest moves first
func (pq PriorityQueue) Less(i, j int) bool {

	if pq.items[i].PopTime != pq.items[j].PopTime {
		return pq.items[i].PopTime < pq.items[j].PopTime
	}

	return pq.items[i].EffectIndex < pq.items[j].EffectIndex
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Stack)
	item.EffectIndex = pq.total
	pq.total++
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	pq.items = old[:n-1]
	return item
}

func (pq *PriorityQueue) Update(item *Stack, newPopTime int64) {
	if item == nil || item.index < 0 || item.index >= len(pq.items) {
		return
	}
	item.PopTime = newPopTime
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Peek() *Stack {
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0]
}
