package match

import (
	"container/heap"
)

// Priority Queue stuff

type EventItem struct {
	Event Event
	Type  string
	seq   int64 // strictly increasing tie-breaker
}

type EventPQ []*EventItem

func (pq EventPQ) Len() int { return len(pq) }

func (pq EventPQ) Less(i, j int) bool {
	ti := pq[i].Event.Timestamp()
	tj := pq[j].Event.Timestamp()

	if ti.Equal(tj) {
		return pq[i].seq < pq[j].seq
	}
	return ti.Before(tj)
}

func (pq EventPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *EventPQ) Push(x any) {
	item := x.(*EventItem)
	*pq = append(*pq, item)
}

func (pq *EventPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

// Peek returns the earliest event without removing it.
// Returns (nil, false) if empty.
func (pq EventPQ) Peek() (*EventItem, bool) {
	if len(pq) == 0 {
		return nil, false
	}
	return pq[0], true
}

func (pq *EventPQ) RemoveAllAutoSchedulers() {
	for i := pq.Len() - 1; i >= 0; i-- {
		item := (*pq)[i]

		if _, ok := item.Event.(*AutoScheduler); ok {
			heap.Remove(pq, i)
		}
	}
}
