package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

type Event struct {
	Time time.Time
}

// have to implement 5 methods for the priority queue
type PriorityQueue []*Event

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].Time.Before(pq[j].Time) }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*Event)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

type Scheduler struct {
	mutex  sync.Mutex
	pq     PriorityQueue
	reinit chan struct{}
	kill   chan struct{}

	timer       *time.Timer
	startTime   time.Time
	currentTime time.Time
	increment   time.Time
}

func CreateScheduler() {
	scheduler := &Scheduler{}
	heap.Init(&scheduler.pq)
	//heap.Push(pq, &Event{time.Now().Add(2 * time.Second), "A"})
}

func (scheduler *Scheduler) AddEvent(t time.Time) {

	scheduler.mutex.Lock()
	scheduler.mutex.Unlock()

}

func (scheduler *Scheduler) Loop() {

	for {

	}

}
