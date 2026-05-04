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
	EffectMana // A: team, B: amount
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

const (
	ChainStepCard      = 0
	ChainStepCardEff   = 1
	ChainStepAbility   = 2
	ChainStepOther     = 3
	ChainStepSummon    = 4
	ChainStepSpell     = 5
	ChainStepRecurring = 6
)

type ChainStep struct {
	Type      int
	CardID    int   // which card this originates from
	StartTime int64 // when it was scheduled (real-time ms)
	PopTime   int64 // when the server pops it (early, to broadcast in time)
	OccurTime int64 // when the effect actually fires (real-time ms)

	Effects         []Effect
	RecurringPeriod int64 // if <= 0, not recurring
}

type Chain struct {
	ChainID     int
	EffectIndex int // the latest effect gets the biggest index

	PopTime int64 // the time we can pop an item from the stack and apply it to the game state

	Steps []ChainStep

	index int // heap position (internal)
}

type PriorityQueue struct {
	Items []*Chain
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
	item := x.(*Chain)
	item.ChainID = pq.total
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

func (pq *PriorityQueue) Update(item *Chain, newPopTime int64) {
	if item == nil || item.index < 0 || item.index >= len(pq.Items) {
		return
	}
	item.PopTime = newPopTime
	heap.Fix(pq, item.index)
}

func (pq PriorityQueue) Peek() *Chain {
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}

// RemoveChainsByCardID removes all chains whose steps belong to cardID.
// Returns the removed chains so callers can emit RemoveChain logs.
func (pq *PriorityQueue) RemoveChainsByCardID(cardID int) []*Chain {
	var removed []*Chain
	kept := pq.Items[:0]

	for _, chain := range pq.Items {
		filtered := chain.Steps[:0]
		for _, step := range chain.Steps {
			if step.CardID != cardID {
				filtered = append(filtered, step)
			}
		}
		chain.Steps = filtered

		if len(chain.Steps) == 0 {
			chain.index = -1
			removed = append(removed, chain)
		} else {
			kept = append(kept, chain)
		}
	}

	pq.Items = kept
	heap.Init(pq)
	return removed
}
