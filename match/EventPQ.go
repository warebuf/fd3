package match

import (
	"container/heap"
	"time"
)

type Event interface {
	Timestamp() time.Time
}

type InputEvent struct {
	T         time.Time
	Type      string
	TeamIdx   int
	CardIndex int
}

func (e *InputEvent) Timestamp() time.Time { return e.T }

type NewPhaseEvent struct {
	T     time.Time
	Phase int
}

func (e *NewPhaseEvent) Timestamp() time.Time { return e.T }

type EffectEvent struct {
	T time.Time
}

func (e *EffectEvent) Timestamp() time.Time { return e.T }

type ChainEvent struct {
	T time.Time
}

func (e *ChainEvent) Timestamp() time.Time { return e.T }

type PauseEndEvent struct {
	T time.Time
}

func (e *PauseEndEvent) Timestamp() time.Time { return e.T }

type EventItem struct {
	Event Event
	Type  string
	seq   int64
	Index int
}

type EventPQ struct {
	items   []*EventItem
	nextSeq int64
}

func (pq EventPQ) Len() int { return len(pq.items) }

func (pq EventPQ) Less(i, j int) bool {
	ti := pq.items[i].Event.Timestamp()
	tj := pq.items[j].Event.Timestamp()
	if ti.Equal(tj) {
		return pq.items[i].seq < pq.items[j].seq
	}
	return ti.Before(tj)
}

func (pq EventPQ) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

func (pq *EventPQ) Push(x any) {
	item := x.(*EventItem)
	item.Index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *EventPQ) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	pq.items = old[:n-1]
	return item
}

func (pq *EventPQ) Enqueue(ev Event, typ string) *EventItem {
	item := &EventItem{Event: ev, Type: typ, seq: pq.nextSeq}
	pq.nextSeq++
	heap.Push(pq, item)
	return item
}

func (pq *EventPQ) Remove(item *EventItem) {
	heap.Remove(pq, item.Index)
}

func (pq *EventPQ) Fix(item *EventItem) {
	heap.Fix(pq, item.Index)
}

func (pq *EventPQ) Peek() *EventItem {
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0]
}
