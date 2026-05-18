package eventpq

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

type ChainResolutionEvent struct {
	T time.Time
}

func (e *ChainResolutionEvent) Timestamp() time.Time { return e.T }

type PauseEndEvent struct {
	T time.Time
}

func (e *PauseEndEvent) Timestamp() time.Time { return e.T }

type RevealEvent struct {
	T time.Time
}

func (e *RevealEvent) Timestamp() time.Time { return e.T }

type EventItem struct {
	Event Event
	Type  string
	Seq   int64
	Index int
}

type EventPQ struct {
	Items   []*EventItem
	NextSeq int64
}

func (pq EventPQ) Len() int { return len(pq.Items) }

func (pq EventPQ) Less(i, j int) bool {
	ti := pq.Items[i].Event.Timestamp()
	tj := pq.Items[j].Event.Timestamp()
	if ti.Equal(tj) {
		return pq.Items[i].Seq < pq.Items[j].Seq
	}
	return ti.Before(tj)
}

func (pq EventPQ) Swap(i, j int) {
	pq.Items[i], pq.Items[j] = pq.Items[j], pq.Items[i]
	pq.Items[i].Index = i
	pq.Items[j].Index = j
}

func (pq *EventPQ) Push(x any) {
	item := x.(*EventItem)
	item.Index = len(pq.Items)
	pq.Items = append(pq.Items, item)
}

func (pq *EventPQ) Pop() any {
	old := pq.Items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	pq.Items = old[:n-1]
	return item
}

func (pq *EventPQ) CreateAndPushEventItem(ev Event, typ string) *EventItem {
	item := &EventItem{Event: ev, Type: typ, Seq: pq.NextSeq}
	pq.NextSeq++
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
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}
