package state

import (
	"container/heap"
	"fmt"
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

func (pq PriorityQueue) Len() int { return len(pq) }

// earliest moves first
func (pq PriorityQueue) Less(i, j int) bool {

	if pq[i].PopTime != pq[j].PopTime {
		return pq[i].PopTime < pq[j].PopTime
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

func (pq *PriorityQueue) Update(item *Stack, newPopTime int64, newEffectIndex uint64, changeEffectIndex bool) {
	if item == nil || item.index < 0 || item.index >= len(*pq) {
		return
	}

	item.PopTime = newPopTime

	if changeEffectIndex {
		item.EffectIndex = newEffectIndex
	}
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Peek() *Stack {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}

func (pq PriorityQueue) Print() {

	fmt.Println("Printing EffPQ")
	for i := 0; i < pq.Len(); i++ {
		fmt.Println(i, "EffectIndex", pq[i].EffectIndex)
		fmt.Println(i, "index", pq[i].index)
		fmt.Println(i, "PopTime", pq[i].PopTime)
		for j := 0; j < len(pq[i].Effects); j++ {
			fmt.Println(i, j, "A", pq[i].Effects[j].A)
			fmt.Println(i, j, "B", pq[i].Effects[j].B)
			fmt.Println(i, j, "C", pq[i].Effects[j].C)
			fmt.Println(i, j, "Type", pq[i].Effects[j].Type)
		}
	}
	fmt.Println("EOF")

}
