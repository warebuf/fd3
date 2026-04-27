package state

import (
	"container/heap"
)

type EffectType int

const (
	EffectAttack EffectType = iota + 1
	EffectActivate
	EffectHeal
	EffectDraw
	EffectMana
	EffectSilence
	EffectGrantAbility
	SummonResolution
	EffectDestroy // A: Destroy Type (0: All, 1: team specific, 2: card specific), B: target (if team specific, team index, if card specific, card index)
)

type Effect struct {
	Type EffectType
	A    int
	B    int
	C    int
}

type StackEntry struct {
	Type    int   // 0: card, 1: card eff, 2: ability, 3: other
	CardID  int   // which card this originates from
	PopTime int64 // when it resolves (real-time ms)

	Effects []Effect
}

type Stack struct {
	EffectIndex int // the latest effect gets the biggest index

	PopTime int64 // the time we can pop an item from the stack and apply it to the game state

	StackEntries []StackEntry

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
