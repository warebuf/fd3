package match

import (
	"time"
)

type Command interface {
	ExecAt() time.Time // when to run
	Apply(m *Match)    // how it changes the match
	String() string    // for logging/metrics
}

type scheduled struct {
	cmd  Command
	when time.Time
	idx  int
}
type eventHeap []*scheduled

func (h eventHeap) Len() int           { return len(h) }
func (h eventHeap) Less(i, j int) bool { return h[i].when.Before(h[j].when) }
func (h eventHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].idx, h[j].idx = i, j }
func (h *eventHeap) Push(x any)        { e := x.(*scheduled); e.idx = len(*h); *h = append(*h, e) }
func (h *eventHeap) Pop() any {
	a := *h
	n := len(a)
	e := a[n-1]
	*h = a[:n-1]
	e.idx = -1
	return e
}
func (h eventHeap) Peek() *scheduled {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}
