package match

import "time"

type Event interface {
	Timestamp() time.Time
}

// Priority Queue stuff

type EventItem struct {
	Event Event
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



// Event types

type InputEvent struct {
	T       time.Time
	Type    string
	TeamIdx int
	UnitIdx int
}

func (e *InputEvent) Timestamp() time.Time { return e.T }

type AutoEvent struct {
	T time.Time
}

func (e *AutoEvent) Timestamp() time.Time { return e.T }

type AutoScheduler struct {
	T time.Time
}

func (e *AutoScheduler) Timestamp() time.Time { return e.T }
