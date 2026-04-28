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
	Type    int   // 0: card, 1: card eff, 2: ability, 3: other, 4: summon
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
	Items []*Stack
	total int
}

func (pq PriorityQueue) Len() int { return len(pq.Items) }

// earliest moves first
func (pq PriorityQueue) Less(i, j int) bool {

	if pq.Items[i].PopTime != pq.Items[j].PopTime {
		return pq.Items[i].PopTime < pq.Items[j].PopTime
	}

	return pq.Items[i].EffectIndex < pq.Items[j].EffectIndex
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.Items[i], pq.Items[j] = pq.Items[j], pq.Items[i]
	pq.Items[i].index = i
	pq.Items[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*Stack)
	item.EffectIndex = pq.total
	pq.total++
	item.index = len(pq.Items)
	pq.Items = append(pq.Items, item)
}

func (pq *PriorityQueue) Pop() any {
	old := pq.Items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	pq.Items = old[:n-1]
	return item
}

func (pq *PriorityQueue) Update(item *Stack, newPopTime int64) {
	if item == nil || item.index < 0 || item.index >= len(pq.Items) {
		return
	}
	item.PopTime = newPopTime
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Peek() *Stack {
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}
