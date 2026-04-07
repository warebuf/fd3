package stackPQ

import (
	"container/heap"

	"github.com/warebuf/fd3/engine/states/gameState"
)

type Stack struct {
	EffectIndex uint64 // the latest effect gets the biggest index

	DeclareTime int64
	ResolveTime int64

	Effects []gameState.Effect

	index int // heap position (internal)
}

type PriorityQueue []*Stack

func (pq PriorityQueue) Len() int { return len(pq) }

// earliest moves first
func (pq PriorityQueue) Less(i, j int) bool {

	if pq[i].ResolveTime != pq[j].ResolveTime {
		return pq[i].ResolveTime < pq[j].ResolveTime
	}

	return pq[i].EffectIndex < pq[j].EffectIndex
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Stack)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil     // avoid memory leak
	item.index = -1    // mark as removed
	*pq = old[0 : n-1] // shrink
	return item
}

func (pq *PriorityQueue) Update(item *Stack, newResolveTime int64, newEffectIndex uint64) {
	if item == nil || item.index < 0 || item.index >= len(*pq) {
		return
	}

	item.ResolveTime = newResolveTime
	item.EffectIndex = newEffectIndex
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Peek() *Stack {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}
